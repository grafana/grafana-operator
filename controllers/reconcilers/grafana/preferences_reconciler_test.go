package grafana

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

var _ = Describe("Reconcile Preferences", func() {
	t := GinkgoT()

	It("is a no-op when spec.preferences is nil", func() {
		r := NewPreferencesReconciler(cl)
		cr := &v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				Preferences: nil,
			},
		}

		status, err := r.Reconcile(context.Background(), cr, &v1beta1.OperatorReconcileVars{}, scheme.Scheme)

		require.NoError(t, err)
		assert.Equal(t, v1beta1.OperatorStageResultSuccess, status)
		assert.Nil(t, meta.FindStatusCondition(cr.Status.Conditions, ConditionPreferencesApplied))
	})
})
