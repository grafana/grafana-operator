package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CompleteReconciler struct{}

func NewCompleteReconciler() reconcilers.OperatorGrafanaReconciler {
	return &CompleteReconciler{}
}

func (r *CompleteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx).WithName("CompleteReconciler")
	logger.Info("grafana installation complete")
	return v1beta1.OperatorStageResultSuccess, nil
}
