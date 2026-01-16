/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionManifestSynchronized = "ManifestSynchronized"
)

// GrafanaManifestReconciler reconciles a GrafanaManifest object
type GrafanaManifestReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Cfg      *Config
	GVRCache sync.Map
}

func (r *GrafanaManifestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaManifestReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaManifest{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		log.Error(err, LogMsgGettingCR)

		return ctrl.Result{}, fmt.Errorf("%s: %w", LogMsgGettingCR, err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaManifest: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, cr)

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionManifestSynchronized)

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionManifestSynchronized)

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)

	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		err := r.reconcileWithInstance(ctx, &grafana, cr)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	condition := buildSynchronizedCondition("Manifest", conditionManifestSynchronized, cr.Generation, applyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(applyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", applyErrors)
	}

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func getManifestNamespace(cr *v1beta1.GrafanaManifest, instance *v1beta1.Grafana) string {
	if cr.Spec.Template.Metadata.Namespace != "" {
		return cr.Spec.Template.Metadata.Namespace
	}

	if instance.Spec.External != nil && instance.Spec.External.TenantNamespace != "" {
		return instance.Spec.External.TenantNamespace
	}

	return "default"
}

func (r *GrafanaManifestReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaManifest) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaManifest")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, instance := range instances {
		cl, dc, err := grafanaclient.NewDynamicClient(ctx, r.Client, &instance)
		if err != nil {
			return fmt.Errorf("building grafana api client: %w", err)
		}

		gvr, err := r.loadGVR(dc, cr.Spec.Template)
		if err != nil {
			return fmt.Errorf("getting group version resource for resource: %w", err)
		}

		ns := getManifestNamespace(cr, &instance)

		err = cl.Resource(gvr).Namespace(ns).Delete(ctx, cr.Spec.Template.Metadata.Name, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete resource: %w", err)
		}

		if err := instance.RemoveNamespacedResource(ctx, r.Client, cr); err != nil {
			return fmt.Errorf("removing manifest from Grafana CR status: %w", err)
		}
	}

	return nil
}

func (r *GrafanaManifestReconciler) reconcileWithInstance(ctx context.Context, instance *v1beta1.Grafana, cr *v1beta1.GrafanaManifest) error {
	cl, dc, err := grafanaclient.NewDynamicClient(ctx, r.Client, instance)
	if err != nil {
		return fmt.Errorf("building grafana api client: %w", err)
	}

	gvr, err := r.loadGVR(dc, cr.Spec.Template)
	if err != nil {
		return fmt.Errorf("getting group version resource for resource: %w", err)
	}

	ns := getManifestNamespace(cr, instance)
	resourceClient := cl.Resource(gvr).Namespace(ns)

	_, err = resourceClient.Get(ctx, cr.Spec.Template.Metadata.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err := resourceClient.Create(ctx, cr.Spec.Template.ToUnstructured(), metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating resource: %w", err)
		}

		return instance.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
	} else if err != nil {
		return fmt.Errorf("fetching existing resource: %w", err)
	}

	if _, err := resourceClient.Update(ctx, cr.Spec.Template.ToUnstructured(), metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("updating resource: %w", err)
	}

	return nil
}

func (r *GrafanaManifestReconciler) loadGVR(cl *discovery.DiscoveryClient, template v1beta1.GrafanaManifestTemplate) (schema.GroupVersionResource, error) {
	gvr, ok := r.GVRCache.Load(gvrKey(template))
	if ok {
		return gvr.(schema.GroupVersionResource), nil //nolint:errcheck
	}

	_, resources, err := cl.ServerGroupsAndResources()
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("failed to discover groups and resources: %w", err)
	}

	target := template.APIVersion
	for _, res := range resources {
		if res.GroupVersion != target {
			continue
		}

		gv, err := schema.ParseGroupVersion(res.GroupVersion)
		if err != nil {
			return schema.GroupVersionResource{}, fmt.Errorf("failed to parse groupversion returned by server: %w", err)
		}

		for _, api := range res.APIResources {
			// this removes subresources like playlist/status
			if strings.Contains(api.Name, "/") {
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: api.Name,
			}
			r.GVRCache.Store(gvrKey(template), gvr)

			return gvr, nil
		}
	}

	return schema.GroupVersionResource{}, errors.New("group version not found")
}

func gvrKey(cr v1beta1.GrafanaManifestTemplate) string {
	return fmt.Sprintf("%s.%s", cr.APIVersion, cr.Kind)
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaManifestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaManifest{}).
		WithEventFilter(ignoreStatusUpdates()).
		Complete(r)
}
