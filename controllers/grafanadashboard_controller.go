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
	"bytes"
	"context"
	"encoding/json"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	grafanav1beta1 "github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GrafanaDashboard object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx)

	dashboard := &grafanav1beta1.GrafanaDashboard{}
	err := r.Get(ctx, req.NamespacedName, dashboard)

	if err != nil {
		if errors.IsNotFound(err) {
			controllerLog.Info("grafana dashboard cr has been deleted", "name", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		controllerLog.Error(err, "error getting grafana dashboard cr")
		return ctrl.Result{}, err
	}

	// skip dashboards without an instance selector
	if dashboard.Spec.InstanceSelector == nil {
		return ctrl.Result{}, nil
	}

	instances, err := r.getMatchingInstances(ctx, dashboard.Spec.InstanceSelector)
	if err != nil {
		return ctrl.Result{}, err
	}

	if len(instances.Items) == 0 {
		controllerLog.Info("no matching instances found for dashboard", "dashboard", dashboard.Name, "namespace", dashboard.Namespace)
	}

	controllerLog.Info("found matching Grafana instances", "count", len(instances.Items))

	complete := true

	for _, grafana := range instances.Items {
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		if grafana.Status.AdminUrl == "" {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			complete = false
			continue
		}

		// first reconcile the plugins
		// append the requested dashboards to a configmap from where the
		// grafana reconciler will pick them up
		err = r.reconcilePlugins(ctx, &grafana, dashboard)
		if err != nil {
			complete = false
			controllerLog.Error(err, "error reconciling plugins", "dashboard", dashboard.Name, "grafana", grafana.Name)
		}

		// then import the dashboard into the matching grafana instances
		err = r.reconcileDashboard(ctx, &grafana, dashboard)
		if err != nil {
			complete = false
			controllerLog.Error(err, "error reconciling dashboard", "dashboard", dashboard.Name, "grafana", grafana.Name)
		}
	}

	// another reconcile needed?
	if complete {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: RequeueDelayError}, nil
}

func (r *GrafanaDashboardReconciler) reconcileDashboard(ctx context.Context, grafana *grafanav1beta1.Grafana, dashboard *grafanav1beta1.GrafanaDashboard) error {
	if strings.TrimSpace(dashboard.Spec.Json) == "" {
		return nil
	}

	_, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	return nil
}

func (r *GrafanaDashboardReconciler) reconcilePlugins(ctx context.Context, grafana *grafanav1beta1.Grafana, dashboard *grafanav1beta1.GrafanaDashboard) error {
	if dashboard.Spec.Plugins == nil || len(dashboard.Spec.Plugins) == 0 {
		return nil
	}

	pluginsConfigMap := model.GetPluginsConfigMap(grafana, r.Scheme)
	selector := client.ObjectKey{
		Namespace: pluginsConfigMap.Namespace,
		Name:      pluginsConfigMap.Name,
	}

	err := r.Client.Get(ctx, selector, pluginsConfigMap)
	if err != nil {
		return err
	}

	val, err := json.Marshal(dashboard.Spec.Plugins.Sanitize())
	if err != nil {
		return err
	}

	if pluginsConfigMap.BinaryData == nil {
		pluginsConfigMap.BinaryData = make(map[string][]byte)
	}

	if bytes.Compare(val, pluginsConfigMap.BinaryData[dashboard.Name]) != 0 {
		pluginsConfigMap.BinaryData[dashboard.Name] = val
		return r.Client.Update(ctx, pluginsConfigMap)
	}

	return nil
}

func (r *GrafanaDashboardReconciler) getMatchingInstances(ctx context.Context, labelSelector *v1.LabelSelector) (grafanav1beta1.GrafanaList, error) {
	var list grafanav1beta1.GrafanaList
	opts := []client.ListOption{
		client.MatchingLabels(labelSelector.MatchLabels),
	}

	err := r.Client.List(ctx, &list, opts...)
	return list, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&grafanav1beta1.GrafanaDashboard{}).
		Complete(r)
}
