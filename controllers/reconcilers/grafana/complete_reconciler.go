package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CompleteReconciler struct{}

func NewCompleteReconciler() reconcilers.OperatorGrafanaReconciler {
	return &CompleteReconciler{}
}

func (r *CompleteReconciler) Reconcile(ctx context.Context, _ *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("CompleteReconciler")
	log.Info("grafana installation complete")

	return v1beta1.OperatorStageResultSuccess, nil
}
