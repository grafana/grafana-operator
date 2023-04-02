package grafana

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceAccountReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {
	sa := model.GetGrafanaServiceAccount(cr, r.Scheme) // todo dinlinemodel

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sa, func() error {
		return v1beta1.Merge(sa, cr.Spec.ServiceAccount)
	})
	if err != nil {
		return nil, err // todo: err condition
	}

	return nil, nil // todo success condition
}
