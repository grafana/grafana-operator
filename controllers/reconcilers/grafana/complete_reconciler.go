package grafana

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	conditionAlertmanagerReady = "AlertmanagerReady"

	reasonAlertmanagerReady        = "AlertmanagerReady"
	reasonAlertmanagerCheckFailed  = "AlertmanagerCheckFailed"
	reasonAlertmanagerConfigFailed = "AlertmanagerConfigFailed"
	reasonNoExternalAlertmanager   = "NoExternalAlertmanager"
)

type CompleteReconciler struct {
	client client.Client
}

func NewCompleteReconciler(cl client.Client) reconcilers.OperatorGrafanaReconciler {
	return &CompleteReconciler{
		client: cl,
	}
}

func (r *CompleteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("CompleteReconciler")

	log.V(1).Info("attempting to authenticate with instance")

	_, err := grafanaclient.GetAuthenticationStatus(ctx, r.client, cr)
	if err != nil {
		cr.Status.Version = ""
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("failed to authenticate with instance: %w", err)
	}

	log.V(1).Info("fetching Grafana version from instance")

	version, err := grafanaclient.GetGrafanaVersion(ctx, r.client, cr)
	if err != nil {
		cr.Status.Version = ""
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("failed fetching version from instance: %w", err)
	}

	cr.Status.Version = version

	if err := r.reconcileAlertmanager(ctx, cr); err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	log.V(1).Info("reconciliation completed")

	return v1beta1.OperatorStageResultSuccess, nil
}

// reconcileAlertmanager aligns the instance's alertmanagersChoice with
// spec.alertmanager. Routing alerts to external Alertmanagers ("external")
// additionally requires at least one external Alertmanager to be configured on
// the instance; for "internal" and "all" this precondition is not relevant.
func (r *CompleteReconciler) reconcileAlertmanager(ctx context.Context, cr *v1beta1.Grafana) error {
	log := logf.FromContext(ctx).WithName("CompleteReconciler")

	desired := cr.Spec.Alertmanager
	if desired == "" {
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionAlertmanagerReady)
		return nil
	}

	status, err := grafanaclient.GetAlertmanagerStatus(ctx, r.client, cr)
	if err != nil {
		msg := fmt.Sprintf("failed to query alertmanager status: %s", err)
		r.setAlertmanagerCondition(cr, metav1.ConditionFalse, reasonAlertmanagerCheckFailed, msg)

		return fmt.Errorf("checking alertmanager status: %w", err)
	}

	// Precondition: routing alerts to external Alertmanagers requires at least
	// one to be configured on the instance.
	if desired == v1beta1.AlertmanagerExternal && status.NumExternalAlertmanagers == 0 {
		msg := `spec.alertmanager is "external" but no external Alertmanager is configured on the instance`
		r.setAlertmanagerCondition(cr, metav1.ConditionFalse, reasonNoExternalAlertmanager, msg)

		return errors.New(msg)
	}

	if status.AlertmanagersChoice != desired {
		log.Info("updating alertmanagersChoice", "from", status.AlertmanagersChoice, "to", desired)

		if err := grafanaclient.SetAlertmanagerChoice(ctx, r.client, cr, desired); err != nil {
			msg := fmt.Sprintf("failed to set alertmanagersChoice to %q: %s", desired, err)
			r.setAlertmanagerCondition(cr, metav1.ConditionFalse, reasonAlertmanagerConfigFailed, msg)

			return fmt.Errorf("configuring alertmanagersChoice: %w", err)
		}
	}

	msg := fmt.Sprintf("alertmanagersChoice set to %q (%d external Alertmanager(s) configured)",
		desired, status.NumExternalAlertmanagers)
	r.setAlertmanagerCondition(cr, metav1.ConditionTrue, reasonAlertmanagerReady, msg)

	return nil
}

func (r *CompleteReconciler) setAlertmanagerCondition(cr *v1beta1.Grafana, status metav1.ConditionStatus, reason, message string) {
	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:               conditionAlertmanagerReady,
		Status:             status,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: cr.Generation,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	})
}
