package reconcilers

import (
	"context"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

type GrafanaReconciler interface {
	Reconcile(ctx context.Context, cr *v1beta1.Grafana, next *v1beta1.Grafana) error
}
