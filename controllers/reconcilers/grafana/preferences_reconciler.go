package grafana

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rt "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	ConditionPreferencesApplied = "PreferencesApplied"

	reasonPreferencesApplied   = "PreferencesApplied"
	reasonHomeDashboardMissing = "HomeDashboardMissing"
)

type PreferencesReconciler struct {
	client client.Client
}

func NewPreferencesReconciler(cl client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PreferencesReconciler{
		client: cl,
	}
}

func (r *PreferencesReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *rt.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("PreferencesReconciler")

	if cr.Spec.Preferences == nil {
		meta.RemoveStatusCondition(&cr.Status.Conditions, ConditionPreferencesApplied)
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
		Theme:            cr.Spec.Preferences.Theme,
		Timezone:         cr.Spec.Preferences.Timezone,
		WeekStart:        cr.Spec.Preferences.WeekStart,
		Language:         cr.Spec.Preferences.Language,
	}

	if _, err := gClient.Org.PatchOrgPreferences(cmd); err != nil { //nolint:errcheck
		// Grafana returns 404 when HomeDashboardUID points to a dashboard
		// that doesn't exist yet. Treat as a soft failure so the instance
		// stays Ready and dashboard reconciles can still create it; the
		// Grafana CR re-reconciles when a matching GrafanaDashboard appears
		// (see GrafanaReconciler.SetupWithManager).
		if apiErr, ok := errors.AsType[*runtime.APIError](err); ok && apiErr.Code == http.StatusNotFound { //nolint:forbidigo
			log.Info("home dashboard not found, will retry once it exists",
				"uid", cr.Spec.Preferences.HomeDashboardUID)
			setPreferencesPending(&cr.Status.Conditions, cr.Generation, cr.Spec.Preferences.HomeDashboardUID)

			return v1beta1.OperatorStageResultSuccess, nil
		}

		return v1beta1.OperatorStageResultFailed, fmt.Errorf("patching org preferences: %w", err)
	}

	if err := r.applyOrganizationName(ctx, gClient, cr.Spec.Preferences.OrganizationName); err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	setPreferencesApplied(&cr.Status.Conditions, cr.Generation)
	log.V(1).Info("preferences applied")

	return v1beta1.OperatorStageResultSuccess, nil
}

// applyOrganizationName sets the org's display name via PUT /api/org if the
// configured name is non-empty and differs from the current name. Reading the
// current name first avoids generating audit-log noise on every reconcile when
// the name is already what the spec asks for.
func (r *PreferencesReconciler) applyOrganizationName(ctx context.Context, gClient *genapi.GrafanaHTTPAPI, want string) error {
	if want == "" {
		return nil
	}

	log := logf.FromContext(ctx).WithName("PreferencesReconciler")

	current, err := gClient.Org.GetCurrentOrg()
	if err != nil {
		return fmt.Errorf("getting current org: %w", err)
	}

	if current.GetPayload().Name == want {
		return nil
	}

	if _, err := gClient.Org.UpdateCurrentOrg(&models.UpdateOrgForm{Name: want}); err != nil { //nolint:errcheck
		return fmt.Errorf("updating org name: %w", err)
	}

	log.Info("organization name updated", "from", current.GetPayload().Name, "to", want)

	return nil
}

func setPreferencesApplied(conditions *[]metav1.Condition, generation int64) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               ConditionPreferencesApplied,
		Status:             metav1.ConditionTrue,
		Reason:             reasonPreferencesApplied,
		Message:            "Grafana preferences applied",
		ObservedGeneration: generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	})
}

func setPreferencesPending(conditions *[]metav1.Condition, generation int64, uid string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               ConditionPreferencesApplied,
		Status:             metav1.ConditionFalse,
		Reason:             reasonHomeDashboardMissing,
		Message:            fmt.Sprintf("home dashboard %q not found in Grafana; preference will be applied once the dashboard exists", uid),
		ObservedGeneration: generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	})
}
