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
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/metrics"
	"github.com/grafana-operator/grafana-operator/v5/controllers/reconcilers"
	"github.com/grafana-operator/grafana-operator/v5/controllers/reconcilers/grafana"
)

const (
	RequeueDelay = 60 * time.Second
)

// GrafanaReconciler reconciles a Grafana object
type GrafanaReconciler struct {
	client.Client
	Log         logr.Logger
	Scheme      *runtime.Scheme
	Discovery   discovery.DiscoveryInterface
	IsOpenShift bool

	subreconcilers []reconcilers.GrafanaReconciler
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanas/finalizers,verbs=update
//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;create;update;delete;watch
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
//+kubebuilder:rbac:groups="",resources=configmaps;secrets;serviceaccounts;services;persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

func (r *GrafanaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("grafana", req.NamespacedName)

	grafana := &v1beta1.Grafana{}
	err := r.Get(ctx, req.NamespacedName, grafana)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("grafana cr has been deleted", "name", req.NamespacedName)
			return ctrl.Result{}, nil
		}

		log.Error(err, "error getting grafana cr")
		return ctrl.Result{}, err
	}

	metrics.GrafanaReconciles.WithLabelValues(grafana.Name).Inc()

	if grafana.IsExternal() {
		statusChanged := false

		if grafana.Status.AdminUrl != grafana.Spec.External.URL {
			statusChanged = true
			grafana.Status.AdminUrl = grafana.Spec.External.URL
		}

		statusChanged = statusChanged || r.setApiAvailableCondition(ctx, grafana)

		if !statusChanged {
			return ctrl.Result{}, nil
		}

		return r.updateStatus(grafana)
	}

	for i, reconciler := range r.subreconcilers {
		err := reconciler.Reconcile(ctx, grafana)
		if err != nil {
			statusChanged := grafana.SetCondition(metav1.Condition{
				Type:               "Ready",
				Status:             metav1.ConditionFalse,
				ObservedGeneration: grafana.Generation,
				LastTransitionTime: metav1.Now(),
				Reason:             v1beta1.CreateResourceFailedReason,
				Message:            err.Error(),
			})

			metrics.GrafanaFailedReconciles.WithLabelValues(grafana.Name, fmt.Sprint(i)).Inc()

			if statusChanged {
				return r.updateStatus(grafana)
			}
		}
	}

	statusChanged := r.setApiAvailableCondition(ctx, grafana)
	if statusChanged {
		return r.updateStatus(grafana)
	}

	return ctrl.Result{}, nil
}

func (r *GrafanaReconciler) setApiAvailableCondition(ctx context.Context, grafana *v1beta1.Grafana) bool {
	apiAvailable := true
	res, err := http.Get(grafana.Status.AdminUrl)
	if err != nil {
		apiAvailable = false
	}
	res.Body.Close()
	if res.StatusCode >= 400 {
		apiAvailable = false
	}

	if apiAvailable {
		return grafana.SetReadyCondition(metav1.ConditionTrue, v1beta1.GrafanaApiAvailableReason, "Grafana API is available")
	} else {
		return grafana.SetReadyCondition(metav1.ConditionFalse, v1beta1.GrafanaApiUnavailableReason, "Grafana API is unavailable")
	}
}

func (r *GrafanaReconciler) updateStatus(cr *v1beta1.Grafana) (ctrl.Result, error) {
	// TODO: DeepEquals check is not a terrible idea
	err := r.Client.Status().Update(context.Background(), cr)
	if err != nil {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: RequeueDelay,
		}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.subreconcilers = []reconcilers.GrafanaReconciler{
		&grafana.ConfigReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.AdminSecretReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.PvcReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.ServiceAccountReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.ServiceReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.IngressReconciler{
			Client:      mgr.GetClient(),
			Scheme:      mgr.GetScheme(),
			IsOpenShift: r.IsOpenShift,
		},
		&grafana.PluginsReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
		},
		&grafana.DeploymentReconciler{
			Client:      mgr.GetClient(),
			Scheme:      mgr.GetScheme(),
			IsOpenShift: r.IsOpenShift,
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Grafana{}).
		Owns(&appsv1.Deployment{}).
		Owns(&v1.ConfigMap{}).
		Owns(&v1.Secret{}).
		Owns(&v1.Service{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&networkingv1.Ingress{}).
		Owns(&routev1.Route{}).
		Complete(r)
}
