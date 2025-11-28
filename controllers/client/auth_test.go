package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetExternalAdminCredentials(t *testing.T) {
	credSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "grafana-credentials",
		},
		Data: map[string][]byte{
			"user": []byte("root"),
			"pass": []byte("secret"),
		},
	}

	testCtx := context.Background()
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	client := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(credSecret).
		Build()

	t.Run("with defined credentials", func(t *testing.T) {
		tests := []struct {
			name          string
			spec          v1beta1.GrafanaSpec
			wantAdminUser string
			wantAdminPass string
		}{
			{
				name: "User and Password from Secret",
				spec: v1beta1.GrafanaSpec{
					External: &v1beta1.External{
						AdminUser: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "grafana-credentials",
							},
							Key: "user",
						},
						AdminPassword: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "grafana-credentials",
							},
							Key: "pass",
						},
					},
				},
				wantAdminUser: "root",
				wantAdminPass: "secret",
			},
			{
				name: "User from config and Password from Secret",
				spec: v1beta1.GrafanaSpec{
					External: &v1beta1.External{
						AdminPassword: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "grafana-credentials",
							},
							Key: "pass",
						},
					},
					Config: map[string]map[string]string{
						"security": {
							"admin_user": "root",
						},
					},
				},
				wantAdminUser: "root",
				wantAdminPass: "secret",
			},
			{
				name: "User and Password from config",
				spec: v1beta1.GrafanaSpec{
					External: &v1beta1.External{},
					Config: map[string]map[string]string{
						"security": {
							"admin_user":     "root",
							"admin_password": "secret",
						},
					},
				},
				wantAdminUser: "root",
				wantAdminPass: "secret",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cr := &v1beta1.Grafana{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "grafana",
					},
					Spec: tt.spec,
				}

				adminUser, err := getExternalAdminUser(testCtx, client, cr)
				require.NoError(t, err)

				adminPassword, err := getExternalAdminPassword(testCtx, client, cr)
				require.NoError(t, err)

				assert.Equal(t, tt.wantAdminUser, adminUser)
				assert.Equal(t, tt.wantAdminPass, adminPassword)
			})
		}
	})

	t.Run("with undefined credentials", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{},
			},
		}

		adminUser, err := getExternalAdminUser(testCtx, client, cr)
		require.Error(t, err)
		assert.Empty(t, adminUser)

		adminPassword, err := getExternalAdminPassword(testCtx, client, cr)
		require.Error(t, err)
		assert.Empty(t, adminPassword)
	})
}

// TODO currently only tests code paths for external grafanas
func TestGetAdminCredentials(t *testing.T) {
	credSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "grafana-credentials",
		},
		Data: map[string][]byte{
			"token": []byte("service-account-key"),
			"user":  []byte("root"),
			"pass":  []byte("secret"),
		},
	}

	testCtx := context.Background()
	s := runtime.NewScheme()
	err := corev1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	client := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(credSecret).
		Build()

	t.Run("with defined credentials", func(t *testing.T) {
		tests := []struct {
			name     string
			external *v1beta1.External
			want     *grafanaAdminCredentials
		}{
			{
				name: "apiKey is preferred",
				external: &v1beta1.External{
					APIKey: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: "grafana-credentials",
						},
						Key: "token",
					},
				},
				want: &grafanaAdminCredentials{
					adminUser:     "",
					adminPassword: "",
					apikey:        "service-account-key",
				},
			},
			{
				name:     "fallback to admin user/password",
				external: &v1beta1.External{},
				want: &grafanaAdminCredentials{
					adminUser:     "root",
					adminPassword: "secret",
					apikey:        "",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cr := &v1beta1.Grafana{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "grafana",
					},
					Spec: v1beta1.GrafanaSpec{
						Config: map[string]map[string]string{
							"security": {
								"admin_user":     "root",
								"admin_password": "secret",
							},
						},
						External: tt.external,
					},
				}

				got, err := getAdminCredentials(testCtx, client, cr)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})

	t.Run("with undefined credentials", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "grafana",
			},
			Spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{},
			},
		}

		got, err := getAdminCredentials(testCtx, client, cr)
		require.Error(t, err)

		assert.Nil(t, got)
	})
}

func createTestJWTFile(t *testing.T) (*os.File, string) {
	t.Helper()

	now := time.Now().Add(60 * time.Second).Unix()
	claims := fmt.Sprintf(`{"exp": %g}`, float64(now))
	encodedClaims := base64.RawStdEncoding.EncodeToString([]byte(claims))

	testJWT := fmt.Sprintf("header.%s.signature", encodedClaims)

	tokenFile, err := os.CreateTemp(os.TempDir(), "token-*")
	require.NoError(t, err)

	written, err := tokenFile.WriteString(testJWT)
	require.Equal(t, len([]byte(testJWT)), written)
	require.NoError(t, err)

	return tokenFile, testJWT
}

func tokenIsValid(t *testing.T, expectedToken, token string, err error) {
	t.Helper()
	require.NoError(t, err)
	require.NotNil(t, jwtCache)
	require.Equal(t, expectedToken, token)
	require.Equal(t, expectedToken, jwtCache.Token)
	require.False(t, jwtCache.Expiration.IsZero())
	require.True(t, jwtCache.Expiration.After(time.Now()))
}

func TestGetBearerToken(t *testing.T) {
	t.Parallel()

	t.Run("error on empty token file", func(t *testing.T) {
		jwtCache = nil

		// Empty file
		tokenFile, err := os.CreateTemp(os.TempDir(), "token-*")
		require.NoError(t, err)

		noToken, err := getBearerToken(tokenFile.Name() + "-dummy")
		require.Error(t, err)
		require.Empty(t, noToken)
		require.Nil(t, jwtCache)
	})

	t.Run("decode test token with nil cache", func(t *testing.T) {
		jwtCache = nil

		tokenFile, token := createTestJWTFile(t)

		parsedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, parsedToken, err)
	})

	t.Run("Read from cache", func(t *testing.T) {
		jwtCache = nil

		tokenFile, token := createTestJWTFile(t)
		parsedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, parsedToken, err)

		os.Remove(tokenFile.Name())
		cachedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, cachedToken, err)
	})

	t.Run("Reset cache and error on mangled token", func(t *testing.T) {
		jwtCache = nil

		tokenFile, token := createTestJWTFile(t)
		parsedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, parsedToken, err)

		// Mangle token
		_, err = tokenFile.WriteString("Invalid.JWT.Token")
		require.NoError(t, err)

		jwtCache = nil
		emptyToken, err := getBearerToken(tokenFile.Name())
		require.Error(t, err)
		require.Empty(t, emptyToken)
	})

	t.Run("expire cache and re-parse token", func(t *testing.T) {
		jwtCache = nil

		tokenFile, token := createTestJWTFile(t)

		parsedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, parsedToken, err)

		// Store original expiration and overwrite cache
		tokenExpiration := jwtCache.Expiration
		jwtCache.Expiration = time.Now().Add(-60 * time.Second)
		jwtCache.Token = ""

		cachedToken, err := getBearerToken(tokenFile.Name())
		tokenIsValid(t, token, cachedToken, err)
		require.Equal(t, tokenExpiration, jwtCache.Expiration)
	})
}
