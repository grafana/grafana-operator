package client

import (
	"context"
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fakeClient struct {
	client.Client
}

func (c fakeClient) Get(_ context.Context, _ client.ObjectKey, ref client.Object, _ ...client.GetOption) error {
	s := ref.(*v1.Secret) //nolint:errcheck
	s.Data = map[string][]byte{
		"fake": []byte("something"),
	}
	return nil
}

func Benchmark_GenGrafanaClient(b *testing.B) {
	ctx := context.TODO()
	var (
		fc  fakeClient
		grf = &v1beta1.Grafana{
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{
					APIKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "secret",
						},
						Key: "fake",
					},
					TLS: &v1beta1.TLSConfig{
						InsecureSkipVerify: true,
					},
				},
			},
			Status: v1beta1.GrafanaStatus{
				AdminURL: "https://grafana.example.com",
			},
		}
	)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		cl, err := NewGeneratedGrafanaClient(ctx, fc, grf)
		if err != nil {
			b.Fatal(err)
		}
		if cl == nil {
			b.Fatal("client is nil")
		}
	}
}
