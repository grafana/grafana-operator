package grafana

import (
	"context"
	"testing"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Complete reconciler", func() {
	t := GinkgoT()

	It("is in progress while the deployment has no ready replica", func() {
		ctx := context.Background()

		cr := &v1beta1.Grafana{
			Namespace: "default",
			Name:      "complete-unready-grafana",
		}

		err := cl.Create(ctx, cr)
		require.NoError(t, err)

		r := NewDeploymentReconciler(cl, false)
		vars := &v1beta1.OperatorReconcileVars{}

		status, err := r.Reconcile(ctx, cr, vars, scheme.Scheme)
		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)

		cr.Status.Version = config.GrafanaVersion

		r = NewCompleteReconciler(cl)

		status, err = r.Reconcile(ctx, cr, vars, scheme.Scheme)
		require.ErrorContains(t, err, "waiting for deployment complete-unready-grafana-deployment to have a ready replica")
		assert.Equal(t, v1beta1.OperatorStageResultInProgress, status)
		assert.Equal(t, config.GrafanaVersion, cr.Status.Version)
	})
})

func TestDeploymentReadiness(t *testing.T) {
	deadlineExceededCondition := appsv1.DeploymentCondition{
		Type:    appsv1.DeploymentProgressing,
		Status:  corev1.ConditionFalse,
		Reason:  deploymentProgressDeadlineExceeded,
		Message: `ReplicaSet "grafana-deployment-abc" has timed out progressing.`,
	}

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	unavailableCondition := func(unavailableFor time.Duration) appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentAvailable,
			Status:             corev1.ConditionFalse,
			Reason:             "MinimumReplicasUnavailable",
			LastTransitionTime: metav1.NewTime(now.Add(-unavailableFor)),
		}
	}

	progressingCondition := func(lastUpdate time.Duration) appsv1.DeploymentCondition {
		return appsv1.DeploymentCondition{
			Type:           appsv1.DeploymentProgressing,
			Status:         corev1.ConditionTrue,
			Reason:         "ReplicaSetUpdated",
			LastUpdateTime: metav1.NewTime(now.Add(-lastUpdate)),
		}
	}

	tests := []struct {
		name        string
		generation  int64
		replicas    *int32
		status      appsv1.DeploymentStatus
		want        v1beta1.OperatorStageStatus
		wantErrText string
	}{
		{
			name:     "ready replica",
			replicas: new(int32(1)),
			status:   appsv1.DeploymentStatus{Replicas: 1, ReadyReplicas: 1},
			want:     v1beta1.OperatorStageResultSuccess,
		},
		{
			name:        "pod starting",
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{Replicas: 1},
			want:        v1beta1.OperatorStageResultInProgress,
			wantErrText: "waiting for deployment grafana-deployment to have a ready replica",
		},
		{
			name:        "pod starting within timeout",
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{Replicas: 1, Conditions: []appsv1.DeploymentCondition{unavailableCondition(readyReplicaTimeout)}},
			want:        v1beta1.OperatorStageResultInProgress,
			wantErrText: "waiting for deployment grafana-deployment to have a ready replica",
		},
		{
			name:        "pod not ready after timeout",
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{Replicas: 1, Conditions: []appsv1.DeploymentCondition{unavailableCondition(readyReplicaTimeout + time.Second)}},
			want:        v1beta1.OperatorStageResultFailed,
			wantErrText: "deployment grafana-deployment has had no ready replica for 3m1s",
		},
		{
			name:     "available without a ready replica after timeout",
			replicas: new(int32(1)),
			status: appsv1.DeploymentStatus{Replicas: 1, Conditions: []appsv1.DeploymentCondition{{
				Type:               appsv1.DeploymentAvailable,
				Status:             corev1.ConditionTrue,
				Reason:             "MinimumReplicasAvailable",
				LastTransitionTime: metav1.NewTime(now.Add(-30 * time.Minute)),
			}}},
			want:        v1beta1.OperatorStageResultFailed,
			wantErrText: "deployment grafana-deployment has had no ready replica for 30m0s",
		},
		{
			name:     "new rollout after timeout",
			replicas: new(int32(1)),
			status: appsv1.DeploymentStatus{Replicas: 1, Conditions: []appsv1.DeploymentCondition{
				unavailableCondition(30 * time.Minute),
				progressingCondition(time.Minute),
			}},
			want:        v1beta1.OperatorStageResultInProgress,
			wantErrText: "waiting for deployment grafana-deployment to have a ready replica",
		},
		{
			name:        "timeout for older generation",
			generation:  2,
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{ObservedGeneration: 1, Replicas: 1, Conditions: []appsv1.DeploymentCondition{unavailableCondition(30 * time.Minute)}},
			want:        v1beta1.OperatorStageResultInProgress,
			wantErrText: "waiting for deployment grafana-deployment to have a ready replica",
		},
		{
			name:        "progress deadline exceeded",
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{Replicas: 1, Conditions: []appsv1.DeploymentCondition{deadlineExceededCondition}},
			want:        v1beta1.OperatorStageResultFailed,
			wantErrText: `deployment grafana-deployment has no ready replica: ReplicaSet "grafana-deployment-abc" has timed out progressing.`,
		},
		{
			name:        "progress deadline exceeded for older generation",
			generation:  2,
			replicas:    new(int32(1)),
			status:      appsv1.DeploymentStatus{ObservedGeneration: 1, Replicas: 1, Conditions: []appsv1.DeploymentCondition{deadlineExceededCondition}},
			want:        v1beta1.OperatorStageResultInProgress,
			wantErrText: "waiting for deployment grafana-deployment to have a ready replica",
		},
		{
			name:     "progress deadline exceeded with a ready replica",
			replicas: new(int32(2)),
			status:   appsv1.DeploymentStatus{Replicas: 2, ReadyReplicas: 1, Conditions: []appsv1.DeploymentCondition{deadlineExceededCondition}},
			want:     v1beta1.OperatorStageResultSuccess,
		},
		{
			name:        "scaled to zero",
			replicas:    new(int32(0)),
			status:      appsv1.DeploymentStatus{},
			want:        v1beta1.OperatorStageResultFailed,
			wantErrText: "deployment grafana-deployment is scaled to zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deployment := &appsv1.Deployment{
				Name:       "grafana-deployment",
				Generation: tt.generation,
				Spec:       appsv1.DeploymentSpec{Replicas: tt.replicas},
				Status:     tt.status,
			}

			got, err := deploymentReadiness(deployment, now)
			if tt.wantErrText == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErrText)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
