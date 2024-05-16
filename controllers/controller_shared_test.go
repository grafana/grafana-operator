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

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
