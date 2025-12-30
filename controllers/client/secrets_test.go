package client

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetValueFromSecretKey(t *testing.T) {
	const (
		namespace = "default"
		key       = "mykey"
		value     = "myvalue"
	)

	emptySecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "empty-secret",
		},
		Data: nil,
	}

	secretWithData := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      "secret",
		},
		Data: map[string][]byte{
			key: []byte(value),
		},
	}

	testCtx := t.Context()

	cl := tk8s.GetFakeClient(t, emptySecret, secretWithData)

	tests := []struct {
		name        string
		keySelector *corev1.SecretKeySelector
		wantErrText string
	}{
		{
			name:        "empty secret key selector",
			keySelector: nil,
			wantErrText: "empty secret key selector",
		},
		{
			name:        "non-existent secret",
			keySelector: tk8s.GetSecretKeySelector(t, "non-existent-secret", key),
			wantErrText: "not found",
		},
		{
			name:        "empty credential secret",
			keySelector: tk8s.GetSecretKeySelector(t, emptySecret.Name, key),
			wantErrText: "empty credential secret",
		},
		{
			name:        "credentials not found",
			keySelector: tk8s.GetSecretKeySelector(t, secretWithData.Name, "non-existent-key"),
			wantErrText: "credentials not found in secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueFromSecretKey(testCtx, cl, namespace, tt.keySelector)
			require.ErrorContains(t, err, tt.wantErrText)
			assert.Nil(t, got)
		})
	}

	t.Run("correct value", func(t *testing.T) {
		want := []byte(value)

		keySelector := tk8s.GetSecretKeySelector(t, secretWithData.Name, key)

		got, err := GetValueFromSecretKey(testCtx, cl, namespace, keySelector)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
