package reconcilers

import (
	"context"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GrafanaSubReconciler interface {
	Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error)
}
