package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Reconcile AdminSecret", func() {
	It("runs successfully with disabled default admin secret", func() {
		r := NewAdminSecretReconciler(k8sClient)
		cr := &v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				DisableDefaultAdminSecret: true,
			},
		}

		vars := &v1beta1.OperatorReconcileVars{}
		status, err := r.Reconcile(context.Background(), cr, vars, scheme.Scheme)

		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(v1beta1.OperatorStageResultSuccess))
	})
})
