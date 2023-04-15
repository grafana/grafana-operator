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

type PvcReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func GetGrafanaDataPVCMeta(cr *v1beta1.Grafana) *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-data", cr.Name),
			Namespace: cr.Namespace,
		},
	}
}

func (r *PvcReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error {
	if cr.Spec.PersistentVolumeClaim == nil {
		return nil
	}

	pvc := GetGrafanaDataPVCMeta(cr)
	if err := controllerutil.SetControllerReference(cr, pvc, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		return v1beta1.Merge(pvc, cr.Spec.PersistentVolumeClaim)
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}

	return nil
}
