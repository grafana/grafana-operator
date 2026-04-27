package resources

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetInheritedLabels(t *testing.T) {
	tests := []struct {
		name            string
		labels          map[string]string
		inheritedLabels map[string]string
		want            map[string]string // NOTE: will get merged with GetCommonLabels
	}{
		{
			name:            "nil",
			labels:          nil,
			inheritedLabels: nil,
			want:            map[string]string{},
		},
		{
			name: "existing and inherited labels are merged",
			labels: map[string]string{
				"existing": "ex",
			},
			inheritedLabels: map[string]string{
				"inherited": "in",
			},
			want: map[string]string{
				"existing":  "ex",
				"inherited": "in",
			},
		},
		{
			name: "applyset labels are ignored",
			labels: map[string]string{
				"existing": "ex",
			},
			inheritedLabels: map[string]string{
				"inherited":                      "in",
				"applyset.kubernetes.io/part-of": "applyset-abc",
			},
			want: map[string]string{
				"existing":  "ex",
				"inherited": "in",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// To replicate SetInheritedLabels() behavior
			maps.Copy(tt.want, GetCommonLabels())

			obj := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "service",
					Namespace: "default",
					Labels:    tt.labels,
				},
			}

			SetInheritedLabels(obj, tt.inheritedLabels)
			got := obj.GetLabels()

			assert.Equal(t, tt.want, got)
		})
	}
}
