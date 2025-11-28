package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CompleteReconciler struct {
	client client.Client
}

func NewCompleteReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &CompleteReconciler{
		client: client,
	}
}

func (r *CompleteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("CompleteReconciler")

	log.V(1).Info("fetching Grafana version from instance")

	version, err := grafanaclient.GetGrafanaVersion(ctx, r.client, cr)
	if err != nil {
		cr.Status.Version = ""
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("failed fetching version from instance: %w", err)
	}

	cr.Status.Version = version

	log.V(1).Info("reconciliation completed")

	return v1beta1.OperatorStageResultSuccess, nil
}
