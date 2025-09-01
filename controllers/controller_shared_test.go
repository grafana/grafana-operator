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
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reusable objectMetas and CommonSpecs to make test tables less verbose
var (
	objectMetaNoMatchingInstances = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "no-match",
	}
	commonSpecNoMatchingInstances = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"no-matching-instances": "test"},
		},
	}

	objectMetaSuspended = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "suspended",
	}
	commonSpecSuspended = v1beta1.GrafanaCommonSpec{
		Suspend: true,
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"suspended": "test"},
		},
	}

	objectMetaApplyFailed = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "apply-failed",
	}
	commonSpecApplyFailed = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"apply-failed": "test"},
		},
	}

	objectMetaInvalidSpec = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "invalid-spec",
	}
	commonSpecInvalidSpec = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"invalid-spec": "test"},
		},
	}

	objectMetaSynchronized = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "synchronized",
	}
	commonSpecSynchronized = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"synchronized": "test"},
		},
	}
)

func requestFromMeta(obj metav1.ObjectMeta) ctrl.Request {
	GinkgoHelper()

	return ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      obj.Name,
			Namespace: obj.Namespace,
		},
	}
}

func TestLabelsSatisfyMatchExpressions(t *testing.T) {
	tests := []struct {
		name             string
		labels           map[string]string
		matchExpressions []metav1.LabelSelectorRequirement
		want             bool
	}{
		{
			name:             "No labels and no expressions",
			labels:           map[string]string{},
			matchExpressions: []metav1.LabelSelectorRequirement{},
			want:             true,
		},
		{
			name:   "No labels",
			labels: map[string]string{},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpExists,
					Key:      "dashboards",
				},
			},
			want: true,
		},
		{
			name: "No matchExpressions",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{},
			want:             true,
		},
		{
			name: "Matches DoesNotExist",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpDoesNotExist,
					Key:      "dashboards",
				},
			},
			want: false,
		},
		{
			name: "Matches Exists",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpExists,
					Key:      "dashboards",
				},
			},
			want: true,
		},
		{
			name: "Matches In",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "dashboards",
					Values: []string{
						"grafana",
					},
				},
			},
			want: true,
		},
		{
			name: "Matches NotIn",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpNotIn,
					Key:      "dashboards",
					Values: []string{
						"grafana",
					},
				},
			},
			want: false,
		},
		{
			name: "Does not match In",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "dashboards",
					Values: []string{
						"grafana-external",
					},
				},
			},
			want: false,
		},
		{
			name: "Does not match NotIn",
			labels: map[string]string{
				"dashboards": "grafana",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpNotIn,
					Key:      "dashboards",
					Values: []string{
						"grafana-external",
					},
				},
			},
			want: true,
		},
		{
			name: "Matches multiple expressions",
			labels: map[string]string{
				"dashboards":  "grafana",
				"environment": "production",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "dashboards",
					Values: []string{
						"grafana",
					},
				},
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "environment",
					Values: []string{
						"production",
					},
				},
			},
			want: true,
		},
		{
			name: "Does not match one of expressions (matching labels, different value)",
			labels: map[string]string{
				"dashboards":  "grafana",
				"environment": "production",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "dashboards",
					Values: []string{
						"grafana",
					},
				},
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "environment",
					Values: []string{
						"development",
					},
				},
			},
			want: false,
		},
		{
			name: "Does not match any of expressions (different labels)",
			labels: map[string]string{
				"random-label-1": "random-value-1",
				"random-label-2": "random-value-2",
			},
			matchExpressions: []metav1.LabelSelectorRequirement{
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "dashboards",
					Values: []string{
						"grafana",
					},
				},
				{
					Operator: metav1.LabelSelectorOpIn,
					Key:      "environment",
					Values: []string{
						"development",
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := labelsSatisfyMatchExpressions(tt.labels, tt.matchExpressions)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUpdatePluginConfigMap(t *testing.T) {
	tests := []struct {
		name          string
		cm            *corev1.ConfigMap
		value         []byte
		key           string
		deprecatedKey string
		want          bool
		wantCM        *corev1.ConfigMap
	}{
		{
			name:          "empty ConfigMap",
			cm:            &corev1.ConfigMap{},
			value:         []byte("aa"),
			key:           "datasource-a-b",
			deprecatedKey: "b-datasource",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"datasource-a-b": []byte("aa"),
				},
			},
		},
		{
			name: "naming migration",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"deprecated-key": []byte("aa"),
					"new-key":        []byte("aa"),
				},
			},
			value:         []byte("aa"),
			key:           "new-key",
			deprecatedKey: "deprecated-key",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"new-key": []byte("aa"),
				},
			},
		},
		{
			name: "updated list of plugins",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"datasource-a-b": []byte("aa"),
				},
			},
			value:         []byte("bb"),
			key:           "datasource-a-b",
			deprecatedKey: "b-datasource",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"datasource-a-b": []byte("bb"),
				},
			},
		},
		{
			name: "same list of plugins",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"datasource-a-b": []byte("aa"),
				},
			},
			value:         []byte("aa"),
			key:           "datasource-a-b",
			deprecatedKey: "b-datasource",
			want:          false,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"datasource-a-b": []byte("aa"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := updatePluginConfigMap(tt.cm, tt.value, tt.key, tt.deprecatedKey)

			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantCM, tt.cm)
		})
	}
}

func TestMergeReconcileErrors(t *testing.T) {
	tests := []struct {
		name    string
		sources []map[string]string
		want    map[string]string
	}{
		{
			name: "Merge multiple maps",
			sources: []map[string]string{
				{
					"default-grafana": "error1",
				},
				{
					"default-grafana": "error2",
				},
				{
					"default-grafana2": "error3",
				},
			},
			want: map[string]string{
				"default-grafana":  "error1; error2",
				"default-grafana2": "error3",
			},
		},
		{
			name: "Nil maps are properly handled",
			sources: []map[string]string{
				nil,
				{
					"default-grafana": "error1",
				},
			},
			want: map[string]string{
				"default-grafana": "error1",
			},
		},
		{
			name: "Empty maps are properly handled",
			sources: []map[string]string{
				{},
				{},
			},
			want: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeReconcileErrors(tt.sources...)
			assert.Equal(t, tt.want, got)
		})
	}
}

var _ = Describe("GetMatchingInstances functions", Ordered, func() {
	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: "matching-instances",
	}}
	allowFolder := v1beta1.GrafanaFolder{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns.Name,
			Name:      "allow-cross-namespace",
		},
		Spec: v1beta1.GrafanaFolderSpec{
			GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
				AllowCrossNamespaceImport: true,
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"matching-instances": "test"},
				},
			},
		},
	}
	// Create duplicate resources, changing key fields
	denyFolder := allowFolder.DeepCopy()
	denyFolder.Name = "deny-cross-namespace"
	denyFolder.Spec.AllowCrossNamespaceImport = false

	matchAllFolder := allowFolder.DeepCopy()
	matchAllFolder.Name = "match-all-grafanas"
	matchAllFolder.Spec.InstanceSelector = &metav1.LabelSelector{} // InstanceSelector is never nil

	BaseGrafana := v1beta1.Grafana{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns.Name,
			Name:      "instance",
			Labels:    map[string]string{"matching-instances": "test"},
		},
		Spec: v1beta1.GrafanaSpec{},
	}
	matchesNothingGrafana := BaseGrafana.DeepCopy()
	matchesNothingGrafana.Name = "no-labels-instance"
	matchesNothingGrafana.Labels = nil

	// Status update is skipped for this
	unreadyGrafana := BaseGrafana.DeepCopy()
	unreadyGrafana.Name = "unready-instance"

	createCRs := []client.Object{&ns, &allowFolder, denyFolder, matchAllFolder, unreadyGrafana}

	// Pre-create all resources
	BeforeAll(func() { // Necessary to use assertions
		for _, cr := range createCRs {
			Expect(k8sClient.Create(testCtx, cr)).Should(Succeed())
		}

		grafanas := []v1beta1.Grafana{BaseGrafana, *matchesNothingGrafana}
		for _, instance := range grafanas {
			Expect(k8sClient.Create(testCtx, &instance)).NotTo(HaveOccurred())

			// Apply status to pass instance ready check
			instance.Status.Stage = v1beta1.OperatorStageComplete
			instance.Status.StageStatus = v1beta1.OperatorStageResultSuccess
			Expect(k8sClient.Status().Update(testCtx, &instance)).ToNot(HaveOccurred())
		}
	})

	Context("Ensure AllowCrossNamespaceImport is upheld by GetScopedMatchingInstances", func() {
		It("Finds all ready instances when instanceSelector is empty", func() {
			instances, err := GetScopedMatchingInstances(testCtx, k8sClient, matchAllFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).To(HaveLen(2 + 2)) // +2 To account for instances created in controllers/suite_test.go to provoke conditions
		})
		It("Finds all ready and Matching instances", func() {
			instances, err := GetScopedMatchingInstances(testCtx, k8sClient, &allowFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).ToNot(BeEmpty())
			Expect(instances).To(HaveLen(2))
		})
		It("Finds matching and ready and matching instance in namespace", func() {
			instances, err := GetScopedMatchingInstances(testCtx, k8sClient, denyFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).ToNot(BeEmpty())
			Expect(instances).To(HaveLen(1))
		})
	})
})
