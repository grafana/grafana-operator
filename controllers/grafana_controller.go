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
	"reflect"

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
	client2 "github.com/grafana-operator/grafana-operator/v5/controllers/client"
	"github.com/grafana-operator/grafana-operator/v5/controllers/metrics"
	"github.com/grafana-operator/grafana-operator/v5/controllers/reconcilers"
	"github.com/grafana-operator/grafana-operator/v5/controllers/reconcilers/grafana"
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

	nextGrafana := grafana.DeepCopy()

	if grafana.IsExternal() {
		if grafana.Status.AdminUrl != grafana.Spec.External.URL {
			nextGrafana.Status.AdminUrl = grafana.Spec.External.URL
		}

		return r.reconcileResult(ctx, grafana, nextGrafana)
	}

	for i, reconciler := range r.subreconcilers {
		err := reconciler.Reconcile(ctx, grafana, nextGrafana)
		if err != nil {
			nextGrafana.SetReadyCondition(metav1.ConditionFalse, v1beta1.CreateResourceFailedReason, err.Error())
			metrics.GrafanaFailedReconciles.WithLabelValues(grafana.Name, fmt.Sprint(i)).Inc()
			break
		}
	}

	return r.reconcileResult(ctx, grafana, nextGrafana)
}

func (r *GrafanaReconciler) updateReadyCondition(ctx context.Context, grafana *v1beta1.Grafana) {
	var availabilityErr error
	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		availabilityErr = err
	} else {
		_, err = grafanaClient.Dashboards()
		if err != nil {
			availabilityErr = err
		}
	}

	if availabilityErr == nil {
		grafana.SetReadyCondition(metav1.ConditionTrue, v1beta1.GrafanaApiAvailableReason, "Grafana API is available")
	} else {
		grafana.SetReadyCondition(metav1.ConditionFalse, v1beta1.GrafanaApiUnavailableReason, fmt.Sprintf("Grafana API is unavailable: %s", availabilityErr))
	}
}

func (r *GrafanaReconciler) reconcileResult(ctx context.Context, grafana *v1beta1.Grafana, nextGrafana *v1beta1.Grafana) (ctrl.Result, error) {
	r.updateReadyCondition(ctx, nextGrafana)
	if !reflect.DeepEqual(grafana.Status, nextGrafana.Status) {
		err := r.Client.Status().Update(context.Background(), nextGrafana)
		if err != nil {
			return ctrl.Result{RequeueAfter: errorRequeueDelay}, fmt.Errorf("failed to update status for grafana %s", grafana.Name)
		}
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
		&grafana.DeploymentReconciler{
			Client:      mgr.GetClient(),
			Scheme:      mgr.GetScheme(),
			IsOpenShift: r.IsOpenShift,
		},
	}

	builder := ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Grafana{})

	if r.IsOpenShift {
		builder = builder.Owns(&routev1.Route{})
	}

	return builder.
		Owns(&appsv1.Deployment{}).
		Owns(&v1.ConfigMap{}).
		Owns(&v1.Secret{}).
		Owns(&v1.Service{}).
		Owns(&v1.ServiceAccount{}).
		Owns(&v1.PersistentVolumeClaim{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}
