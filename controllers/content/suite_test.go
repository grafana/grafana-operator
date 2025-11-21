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

package content

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
)

// NopContentResource is intended for testing only.
type NopContentResource struct {
	metav1.TypeMeta
	metav1.ObjectMeta

	Spec   v1beta1.GrafanaContentSpec
	Status v1beta1.GrafanaContentStatus
}

func (in *NopContentResource) GrafanaContentSpec() *v1beta1.GrafanaContentSpec {
	return &in.Spec
}

func (in *NopContentResource) GrafanaContentStatus() *v1beta1.GrafanaContentStatus {
	return &in.Status
}

func (in *NopContentResource) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}

	return nil
}

func (in *NopContentResource) DeepCopy() *NopContentResource {
	if in == nil {
		return nil
	}

	out := new(NopContentResource)
	in.DeepCopyInto(out)

	return out
}

func (in *NopContentResource) DeepCopyInto(out *NopContentResource) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func TestAPIs(t *testing.T) {
	if testing.Short() {
		t.Skip("-short was passed, skipping Content")
	}

	RunSpecs(t, "Content Suite")
}

var _ = BeforeSuite(func() {
	t := GinkgoT()

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	k8sClient, err := client.New(cfg, client.Options{Scheme: scheme.Scheme})
	require.NoError(t, err)
	require.NotNil(t, k8sClient)
})
