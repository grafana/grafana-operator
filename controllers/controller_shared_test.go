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
	"errors"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			value:         []byte("1"),
			key:           "a",
			deprecatedKey: "b",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
		},
		{
			name: "naming migration",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"deprecated": []byte("1"),
				},
			},
			value:         []byte("1"),
			key:           "new",
			deprecatedKey: "deprecated",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"new": []byte("1"),
				},
			},
		},
		{
			name: "updated list of plugins",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
			value:         []byte("2"),
			key:           "a",
			deprecatedKey: "b",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("2"),
				},
			},
		},
		{
			name: "same list of plugins",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
			value:         []byte("1"),
			key:           "a",
			deprecatedKey: "b",
			want:          false,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
		},
		{
			name: "removed plugin (nil)",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
			value:         nil,
			key:           "a",
			deprecatedKey: "b",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{},
			},
		},
		{
			name: "removed plugin (empty slice)",
			cm: &corev1.ConfigMap{
				BinaryData: map[string][]byte{
					"a": []byte("1"),
				},
			},
			value:         []byte{},
			key:           "a",
			deprecatedKey: "b",
			want:          true,
			wantCM: &corev1.ConfigMap{
				BinaryData: map[string][]byte{},
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

func TestGetGrafanaRefValue(t *testing.T) {
	tests := []struct {
		name     string
		instance *v1beta1.Grafana
		selector *corev1.ObjectFieldSelector
		want     string
		wantErr  error
	}{
		{
			name: "Get simple value",
			instance: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Version: "0.0.1",
				},
			},
			selector: &corev1.ObjectFieldSelector{
				FieldPath: "spec.version",
			},
			want: "0.0.1",
		},
		{
			name: "Get value with dots",
			instance: &v1beta1.Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"foo.bar/baz": "some-value",
					},
				},
			},
			selector: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.annotations.[foo.bar/baz]",
			},
			want: "some-value",
		},
		{
			name:     "Get missing value",
			instance: &v1beta1.Grafana{},
			selector: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.annotations.[foo.bar/baz]",
			},
			wantErr: errors.New("field 'metadata.annotations.[foo.bar/baz]' not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getGrafanaRefValue(tt.instance, tt.selector)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

var _ = Describe("GetMatchingInstances functions", Ordered, func() {
	t := GinkgoT()

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
			err := cl.Create(testCtx, cr)
			require.NoError(t, err)
		}

		grafanas := []v1beta1.Grafana{BaseGrafana, *matchesNothingGrafana}
		for _, instance := range grafanas {
			err := cl.Create(testCtx, &instance)
			require.NoError(t, err)

			// Apply status to pass instance ready check
			instance.Status.Stage = v1beta1.OperatorStageComplete
			instance.Status.StageStatus = v1beta1.OperatorStageResultSuccess

			err = cl.Status().Update(testCtx, &instance)
			require.NoError(t, err)
		}
	})

	Context("Ensure AllowCrossNamespaceImport is upheld by GetScopedMatchingInstances", func() {
		It("Finds all ready instances when instanceSelector is empty", func() {
			instances, err := GetScopedMatchingInstances(testCtx, cl, matchAllFolder)
			require.NoError(t, err)
			assert.NotEmpty(t, instances)
			assert.Len(t, instances, 2+2) // +2 To account for instances created in controllers/suite_test.go to provoke conditions
		})
		It("Finds all ready and Matching instances", func() {
			instances, err := GetScopedMatchingInstances(testCtx, cl, &allowFolder)
			require.NoError(t, err)
			assert.NotEmpty(t, instances)
			assert.Len(t, instances, 2)
		})
		It("Finds matching and ready and matching instance in namespace", func() {
			instances, err := GetScopedMatchingInstances(testCtx, cl, denyFolder)
			require.NoError(t, err)
			assert.NotEmpty(t, instances)
			assert.Len(t, instances, 1)
		})
	})
})

type errTypeA struct{}

func (e errTypeA) Error() string {
	return "a"
}

type errTypeB struct{}

func (e errTypeB) Error() string {
	return "B"
}

func TestIsErrorType(t *testing.T) {
	var (
		errA errTypeA
		errB errTypeB
	)

	t.Run("same type", func(t *testing.T) {
		want := true
		got := IsErrorType[errTypeA](errA)
		assert.Equal(t, want, got)
	})

	t.Run("different type", func(t *testing.T) {
		want := false
		got := IsErrorType[errTypeA](errB)
		assert.Equal(t, want, got)
	})
}

func TestIsNotErrorType(t *testing.T) {
	var (
		errA errTypeA
		errB errTypeB
	)

	t.Run("same type", func(t *testing.T) {
		want := false
		got := IsNotErrorType[errTypeA](errA)
		assert.Equal(t, want, got)
	})

	t.Run("different type", func(t *testing.T) {
		want := true
		got := IsNotErrorType[errTypeA](errB)
		assert.Equal(t, want, got)
	})
}
