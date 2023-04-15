package grafana

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/api/core/v1"
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

func GetGrafanaServiceAccountMeta(cr *v1beta1.Grafana) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
	}
}

func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error {
	sa := GetGrafanaServiceAccountMeta(cr)
	if err := controllerutil.SetControllerReference(cr, sa, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, sa, func() error {
		return v1beta1.Merge(sa, cr.Spec.ServiceAccount)
	})
	if err != nil {
		return fmt.Errorf("failed to create of update: %w", err)
	}

	return nil
}
