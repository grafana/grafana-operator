package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	client := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(emptySecret, secretWithData).
		Build()

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
			name: "non-existent secret",
			keySelector: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "non-existent-secret",
				},
				Key: key,
			},
			wantErrText: "not found",
		},
		{
			name: "empty credential secret",
			keySelector: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: emptySecret.Name,
				},
				Key: key,
			},
			wantErrText: "empty credential secret",
		},
		{
			name: "credentials not found",
			keySelector: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secretWithData.Name,
				},
				Key: "non-existent-key",
			},
			wantErrText: "credentials not found in secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetValueFromSecretKey(testCtx, client, namespace, tt.keySelector)
			require.ErrorContains(t, err, tt.wantErrText)
			assert.Nil(t, got)
		})
	}

	t.Run("correct value", func(t *testing.T) {
		want := []byte(value)

		keySelector := &corev1.SecretKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{
				Name: secretWithData.Name,
			},
			Key: key,
		}

		got, err := GetValueFromSecretKey(testCtx, client, namespace, keySelector)
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})
}
