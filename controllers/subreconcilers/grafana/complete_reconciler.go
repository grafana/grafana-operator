package grafana

import (
	"context"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type CompleteReconciler struct{}

func (r *CompleteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {
	logger := log.FromContext(ctx)
	logger.Info("grafana installation complete")
	// todo check api availability
	return nil, nil // todo: success condition
}
