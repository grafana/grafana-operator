package grafana

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
)

func Test_getResources(t *testing.T) {
	defaultRequests := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse(MemoryRequest),
		v1.ResourceCPU:    resource.MustParse(CpuRequest),
	}

	defaultLimits := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse(MemoryLimit),
		v1.ResourceCPU:    resource.MustParse(CpuLimit),
	}

	defaultLimitsWithoutCPU := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse(MemoryLimit),
	}

	tests := []struct {
		name string
		cr   *v1beta1.Grafana
		want v1.ResourceRequirements
	}{
		{
			name: "Resources entirely not defined should return defaults",
			cr:   &v1beta1.Grafana{},
			want: v1.ResourceRequirements{
				Requests: defaultRequests,
				Limits:   defaultLimits,
			},
		},
		{
			name: "Resource Limits not defined should return defaults",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Deployment: &v1beta1.DeploymentV1{
						Spec: v1beta1.DeploymentV1Spec{
							Template: &v1beta1.DeploymentV1PodTemplateSpec{
								Spec: &v1beta1.DeploymentV1PodSpec{
									Containers: []v1.Container{
										{
											Name: "grafana",
											Resources: v1.ResourceRequirements{
												Requests: defaultRequests,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: v1.ResourceRequirements{
				Requests: defaultRequests,
				Limits:   defaultLimits,
			},
		},
		{
			name: "Only CPU Limit not defined should return default Requests and default Memory Limit",
			cr: &v1beta1.Grafana{
				Spec: v1beta1.GrafanaSpec{
					Deployment: &v1beta1.DeploymentV1{
						Spec: v1beta1.DeploymentV1Spec{
							Template: &v1beta1.DeploymentV1PodTemplateSpec{
								Spec: &v1beta1.DeploymentV1PodSpec{
									Containers: []v1.Container{
										{
											Name: "grafana",
											Resources: v1.ResourceRequirements{
												Requests: defaultRequests,
												Limits:   defaultLimitsWithoutCPU,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			want: v1.ResourceRequirements{
				Requests: defaultRequests,
				Limits:   defaultLimitsWithoutCPU,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.want
			got := getResources(tt.cr)

			assert.Equal(t, want, got)
		})
	}
}
