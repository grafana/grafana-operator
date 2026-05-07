package grafana

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type PreferencesReconciler struct {
	client client.Client
}

func NewPreferencesReconciler(cl client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PreferencesReconciler{
		client: cl,
	}
}

func (r *PreferencesReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("PreferencesReconciler")

	if cr.Spec.Preferences == nil {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.client, cr)
	if err != nil {
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("building grafana api client: %w", err)
	}

	// PATCH so empty fields don't clobber preferences set elsewhere.
	// Add new fields here as they are added to v1beta1.GrafanaPreferences.
	cmd := &models.PatchPrefsCmd{
		HomeDashboardUID: cr.Spec.Preferences.HomeDashboardUID,
	}

	if _, err := gClient.Org.PatchOrgPreferences(cmd); err != nil {
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("patching org preferences: %w", err)
	}

	log.V(1).Info("preferences applied")

	return v1beta1.OperatorStageResultSuccess, nil
}
