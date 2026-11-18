package grafana

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"github.com/grafana/grafana-operator/v5/controllers/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

const deploymentProgressDeadlineExceeded = "ProgressDeadlineExceeded"

// How long to wait for a ready replica before failing the stage.
const readyReplicaTimeout = 3 * time.Minute

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

	if !cr.IsExternal() {
		deployment := resources.GetGrafanaDeployment(cr, nil)
		selector := client.ObjectKey{
			Namespace: deployment.Namespace,
			Name:      deployment.Name,
		}

		err := r.client.Get(ctx, selector, deployment)
		if apierrors.IsNotFound(err) {
			return v1beta1.OperatorStageResultInProgress, fmt.Errorf("waiting for deployment %s to be created", deployment.Name)
		}

		if err != nil {
			return v1beta1.OperatorStageResultFailed, fmt.Errorf("failed to fetch deployment: %w", err)
		}

		stageStatus, err := deploymentReadiness(deployment, time.Now())
		if err != nil {
			return stageStatus, err
		}
	}

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

	log.V(1).Info("reconciliation completed")

	return v1beta1.OperatorStageResultSuccess, nil
}

func deploymentReadiness(deployment *appsv1.Deployment, now time.Time) (v1beta1.OperatorStageStatus, error) {
	if deployment.Spec.Replicas != nil && *deployment.Spec.Replicas == 0 {
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("deployment %s is scaled to zero", deployment.Name)
	}

	if deployment.Status.ReadyReplicas > 0 {
		return v1beta1.OperatorStageResultSuccess, nil
	}

	waiting := fmt.Errorf("waiting for deployment %s to have a ready replica", deployment.Name)

	// Conditions for an older generation do not apply to the current rollout.
	if deployment.Status.ObservedGeneration < deployment.Generation {
		return v1beta1.OperatorStageResultInProgress, waiting
	}

	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing && condition.Status == corev1.ConditionFalse &&
			condition.Reason == deploymentProgressDeadlineExceeded {
			return v1beta1.OperatorStageResultFailed, fmt.Errorf("deployment %s has no ready replica: %s", deployment.Name, condition.Message)
		}
	}

	if since, ok := waitingSince(deployment); ok {
		if waited := now.Sub(since); waited > readyReplicaTimeout {
			return v1beta1.OperatorStageResultFailed, fmt.Errorf("deployment %s has had no ready replica for %s", deployment.Name, waited.Round(time.Second))
		}
	}

	return v1beta1.OperatorStageResultInProgress, waiting
}

// The time from which to count the wait for a ready replica. This is the later of the last change to the
// Available condition and the last progress of the rollout. False if the deployment has neither yet.
func waitingSince(deployment *appsv1.Deployment) (time.Time, bool) {
	var since time.Time

	found := false

	for _, condition := range deployment.Status.Conditions {
		var changed metav1.Time

		switch condition.Type {
		case appsv1.DeploymentAvailable:
			changed = condition.LastTransitionTime
		case appsv1.DeploymentProgressing:
			changed = condition.LastUpdateTime
		case appsv1.DeploymentReplicaFailure:
			continue
		default:
			continue
		}

		found = true

		if changed.After(since) {
			since = changed.Time
		}
	}

	return since, found
}
