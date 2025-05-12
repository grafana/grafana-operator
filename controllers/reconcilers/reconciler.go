package reconcilers

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

type OperatorGrafanaReconciler interface {
	Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars) (v1beta1.OperatorStageStatus, error)
}
