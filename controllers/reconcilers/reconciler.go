package reconcilers

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

type OperatorGrafanaReconciler interface {
	Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error)
}
