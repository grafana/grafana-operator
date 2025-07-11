/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dashboard Reconciler: Provoke Conditions", func() {
	tests := []struct {
		name          string
		cr            *v1beta1.GrafanaDashboard
		wantCondition string
		wantReason    string
	}{
		{
			name: "Suspended Condition",
			cr: &v1beta1.GrafanaDashboard{
				ObjectMeta: objectMetaSuspended,
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaCommonSpec:  commonSpecSuspended,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionSuspended,
			wantReason:    conditionReasonApplySuspended,
		},
		{
			name: "NoMatchingInstances Condition",
			cr: &v1beta1.GrafanaDashboard{
				ObjectMeta: objectMetaNoMatchingInstances,
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaCommonSpec:  commonSpecNoMatchingInstances,
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{JSON: "{}"},
				},
			},
			wantCondition: conditionNoMatchingInstance,
			wantReason:    conditionReasonEmptyAPIReply,
		},
	}

	for _, test := range tests {
		It(test.name, func() {
			err := k8sClient.Create(testCtx, test.cr)
			Expect(err).ToNot(HaveOccurred())

			// Reconciliation Request
			req := requestFromMeta(test.cr.ObjectMeta)

			// Reconcile
			r := GrafanaDashboardReconciler{Client: k8sClient}
			_, err = r.Reconcile(testCtx, req)
			Expect(err).ShouldNot(HaveOccurred())

			resultCr := &v1beta1.GrafanaDashboard{}
			Expect(r.Get(testCtx, req.NamespacedName, resultCr)).Should(Succeed())

			// Verify Condition
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Type", test.wantCondition)))
			Expect(resultCr.Status.Conditions).Should(ContainElement(HaveField("Reason", test.wantReason)))
		})
	}
})
