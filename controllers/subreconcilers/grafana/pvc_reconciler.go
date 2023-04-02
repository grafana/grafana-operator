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
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PvcReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *PvcReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {
	logger := log.FromContext(ctx)

	if cr.Spec.PersistentVolumeClaim == nil {
		logger.Info("skip creating persistent volume claim")
		return nil, nil // todo: success condition
	}

	pvc := model.GetGrafanaDataPVC(cr, r.Scheme) // todod inline model
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, pvc, func() error {
		return v1beta1.Merge(pvc, cr.Spec.PersistentVolumeClaim)
	})
	if err != nil {
		return nil, err // todo error condition
	}

	return nil, nil // todo success condition
}
