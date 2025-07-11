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
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	objectMetaNoMatchingInstances = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "no-match",
	}
	commonSpecNoMatchingInstances = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"test": "no-matching-instances"},
		},
	}

	objectMetaSuspended = metav1.ObjectMeta{
		Namespace: "default",
		Name:      "suspended",
	}
	commonSpecSuspended = v1beta1.GrafanaCommonSpec{
		Suspend: true,
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"test": "suspended"},
		},
	}

	objectMetaApplyFailed = metav1.ObjectMeta{
		Namespace: "apply-failed",
		Name:      "apply-failed",
	}
	commonSpecApplyFailed = v1beta1.GrafanaCommonSpec{
		InstanceSelector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"test": "apply-failed"},
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
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "get-matching-test",
		},
	}
	allowFolder := v1beta1.GrafanaFolder{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "grafana.integreatly.org/v1beta1",
			Kind:       "GrafanaFolder",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "allow-cross-namespace",
			Namespace: namespace.Name,
		},
		Spec: v1beta1.GrafanaFolderSpec{
			GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
				AllowCrossNamespaceImport: true,
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "folder",
					},
				},
			},
		},
	}
	// Create duplicate resources, changing key fields
	denyFolder := allowFolder.DeepCopy()
	denyFolder.Name = "deny-cross-namespace"
	denyFolder.Spec.AllowCrossNamespaceImport = false

	matchAllFolder := allowFolder.DeepCopy()
	matchAllFolder.Name = "invalid-match-labels"
	matchAllFolder.Spec.InstanceSelector = &metav1.LabelSelector{} // InstanceSelector is never nil

	DefaultGrafana := v1beta1.Grafana{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "grafana.integreatly.org/v1beta1",
			Kind:       "Grafana",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "instance",
			Namespace: "default",
			Labels: map[string]string{
				"test": "folder",
			},
		},
		Spec: v1beta1.GrafanaSpec{},
	}
	matchesNothingGrafana := DefaultGrafana.DeepCopy()
	matchesNothingGrafana.Name = "match-nothing-instance"
	matchesNothingGrafana.Labels = nil

	secondNamespaceGrafana := DefaultGrafana.DeepCopy()
	secondNamespaceGrafana.Name = "second-namespace-instance"
	secondNamespaceGrafana.Namespace = namespace.Name

	// Status update is skipped for this
	unreadyGrafana := DefaultGrafana.DeepCopy()
	unreadyGrafana.Name = "unready-instance"

	ctx := context.Background()
	testLog := logf.FromContext(ctx).WithSink(logf.NullLogSink{})
	ctx = logf.IntoContext(ctx, testLog)

	// Pre-create all resources
	BeforeAll(func() { // Necessary to use assertions
		Expect(k8sClient.Create(ctx, &namespace)).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, &allowFolder)).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, denyFolder)).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, matchAllFolder)).NotTo(HaveOccurred())
		Expect(k8sClient.Create(ctx, unreadyGrafana)).NotTo(HaveOccurred())

		grafanas := []v1beta1.Grafana{DefaultGrafana, *matchesNothingGrafana, *secondNamespaceGrafana}
		for _, instance := range grafanas {
			Expect(k8sClient.Create(ctx, &instance)).NotTo(HaveOccurred())

			// Apply status to pass instance ready check
			instance.Status.Stage = v1beta1.OperatorStageComplete
			instance.Status.StageStatus = v1beta1.OperatorStageResultSuccess
			Expect(k8sClient.Status().Update(ctx, &instance)).ToNot(HaveOccurred())
		}
	})

	Context("Ensure AllowCrossNamespaceImport is upheld by GetScopedMatchingInstances", func() {
		It("Finds all ready instances when instanceSelector is empty", func() {
			instances, err := GetScopedMatchingInstances(ctx, k8sClient, matchAllFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).To(HaveLen(3 + 1)) // +1 To account for instance created in suite_test.go to provoke ApplyFailed conditions
		})
		It("Finds all ready and Matching instances", func() {
			instances, err := GetScopedMatchingInstances(ctx, k8sClient, &allowFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).ToNot(BeEmpty())
			Expect(instances).To(HaveLen(2))
		})
		It("Finds matching and ready and matching instance in namespace", func() {
			instances, err := GetScopedMatchingInstances(ctx, k8sClient, denyFolder)
			Expect(err).ToNot(HaveOccurred())
			Expect(instances).ToNot(BeEmpty())
			Expect(instances).To(HaveLen(1))
		})
	})
})
