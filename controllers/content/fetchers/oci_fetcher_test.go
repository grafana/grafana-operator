package fetchers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
)

// makeDockerConfigJSON builds a kubernetes.io/dockerconfigjson secret payload for registryHost.
func makeDockerConfigJSON(t *testing.T, registryHost, username, password string) []byte {
	t.Helper()

	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	cfg := map[string]any{
		"auths": map[string]any{
			registryHost: map[string]any{"auth": auth},
		},
	}

	raw, err := json.Marshal(cfg)
	require.NoError(t, err)

	return raw
}

// ociDashboard builds a GrafanaDashboard CR pointing at an OCI source.
//
// InsecurePlainHTTP is always true: the test registry is plain HTTP (httptest.NewServer),
// and the production code requires InsecurePlainHTTP=true to skip HTTPS. The HTTPS path
// is exercised by oras-go's own test suite and not re-tested here.
func ociDashboard(reference, path string, pullSecretRef *corev1.LocalObjectReference) *v1beta1.GrafanaDashboard {
	return &v1beta1.GrafanaDashboard{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: v1beta1.GrafanaDashboardSpec{
			GrafanaContentSpec: v1beta1.GrafanaContentSpec{
				OCI: &v1beta1.GrafanaContentOCI{
					Reference:         reference,
					Path:              path,
					PullSecretRef:     pullSecretRef,
					InsecurePlainHTTP: true,
				},
			},
		},
	}
}

func TestFetchFromOCI(t *testing.T) {
	const wantJSON = `{"title":"test","panels":[]}`

	t.Run("happy path tag", func(t *testing.T) {
		reg, host := newFakeRegistry(t)
		reg.pushArtifact(t, "team/boards", "v1", map[string][]byte{"board.json": []byte(wantJSON)})

		cr := ociDashboard(host+"/team/boards:v1", "board.json", nil)

		got, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.NoError(t, err)
		assert.JSONEq(t, wantJSON, string(got))
	})

	t.Run("happy path digest", func(t *testing.T) {
		reg, host := newFakeRegistry(t)
		digest := reg.pushArtifact(t, "team/boards", "v1", map[string][]byte{"board.json": []byte(wantJSON)})

		cr := ociDashboard(host+"/team/boards@"+digest, "board.json", nil)

		got, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.NoError(t, err)
		assert.JSONEq(t, wantJSON, string(got))
	})

	t.Run("nested file path", func(t *testing.T) {
		reg, host := newFakeRegistry(t)
		reg.pushArtifact(t, "team/boards", "v1", map[string][]byte{"subdir/board.json": []byte(wantJSON)})

		cr := ociDashboard(host+"/team/boards:v1", "subdir/board.json", nil)

		got, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.NoError(t, err)
		assert.JSONEq(t, wantJSON, string(got))
	})

	t.Run("file not in artifact", func(t *testing.T) {
		reg, host := newFakeRegistry(t)
		reg.pushArtifact(t, "team/boards", "v1", map[string][]byte{"board.json": []byte(wantJSON)})

		cr := ociDashboard(host+"/team/boards:v1", "absent.json", nil)

		_, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "absent.json")
	})

	t.Run("reference without tag or digest", func(t *testing.T) {
		_, host := newFakeRegistry(t)

		cr := ociDashboard(host+"/team/boards", "board.json", nil)

		_, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "tag or digest")
	})

	t.Run("invalid image reference", func(t *testing.T) {
		cr := ociDashboard("not a valid image!!:v1", "board.json", nil)

		_, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse oci reference")
	})

	t.Run("pull secret happy path", func(t *testing.T) {
		reg, host := newFakeRegistry(t)
		reg.requireAuth = true
		reg.user, reg.pass = "user", "pass"
		reg.pushArtifact(t, "team/boards", "v1", map[string][]byte{"board.json": []byte(wantJSON)})

		rawCfg := makeDockerConfigJSON(t, host, "user", "pass")
		pullSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "regcred", Namespace: "default"},
			Type:       corev1.SecretTypeDockerConfigJson,
			Data:       map[string][]byte{corev1.DockerConfigJsonKey: rawCfg},
		}
		cl := tk8s.GetFakeClient(t, pullSecret)

		cr := ociDashboard(host+"/team/boards:v1", "board.json",
			&corev1.LocalObjectReference{Name: "regcred"})

		got, err := FetchFromOCI(context.Background(), cr, cl)
		require.NoError(t, err)
		assert.JSONEq(t, wantJSON, string(got))
	})

	t.Run("missing pull secret", func(t *testing.T) {
		_, host := newFakeRegistry(t)

		cr := ociDashboard(host+"/team/boards:v1", "board.json",
			&corev1.LocalObjectReference{Name: "does-not-exist"})

		_, err := FetchFromOCI(context.Background(), cr, tk8s.GetFakeClient(t))
		require.Error(t, err)
	})

	t.Run("wrong secret type", func(t *testing.T) {
		_, host := newFakeRegistry(t)
		opaqueSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "bad-secret", Namespace: "default"},
			Type:       corev1.SecretTypeOpaque,
			Data:       map[string][]byte{"some-key": []byte("value")},
		}
		cl := tk8s.GetFakeClient(t, opaqueSecret)

		cr := ociDashboard(host+"/team/boards:v1", "board.json",
			&corev1.LocalObjectReference{Name: "bad-secret"})

		_, err := FetchFromOCI(context.Background(), cr, cl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), string(corev1.SecretTypeDockerConfigJson))
	})

	t.Run("malformed pull secret body", func(t *testing.T) {
		_, host := newFakeRegistry(t)
		badSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "malformed", Namespace: "default"},
			Type:       corev1.SecretTypeDockerConfigJson,
			Data:       map[string][]byte{corev1.DockerConfigJsonKey: []byte("not json{")},
		}
		cl := tk8s.GetFakeClient(t, badSecret)

		cr := ociDashboard(host+"/team/boards:v1", "board.json",
			&corev1.LocalObjectReference{Name: "malformed"})

		_, err := FetchFromOCI(context.Background(), cr, cl)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse pull secret")
	})
}

func TestHostMatches(t *testing.T) {
	cases := []struct {
		configHost   string
		registryHost string
		want         bool
	}{
		{"ghcr.io", "ghcr.io", true},
		{"https://ghcr.io", "ghcr.io", true},
		{"https://index.docker.io/v1/", "index.docker.io", true},
		{"gcr.io", "ghcr.io", false},
		{"https://registry.example.com", "registry.example.com", true},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s~%s", tc.configHost, tc.registryHost), func(t *testing.T) {
			assert.Equal(t, tc.want, hostMatches(tc.configHost, tc.registryHost))
		})
	}
}
