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

	f, err := os.CreateTemp(os.TempDir(), "token-*")
	require.NoError(t, err)

	written, err := f.WriteString(testJWT)
	require.Equal(t, len([]byte(testJWT)), written)
	require.NoError(t, err)

	return f, testJWT
}

func createFileWithContent(t *testing.T, content string) *os.File {
	t.Helper()

	f, err := os.CreateTemp(os.TempDir(), "test-*")
	require.NoError(t, err)

	written, err := f.WriteString(content)
	require.Equal(t, len([]byte(content)), written)
	require.NoError(t, err)

	return f
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

	t.Run("non-existent file", func(t *testing.T) {
		jwtCache = nil

		f, err := os.CreateTemp(os.TempDir(), "token-*")
		defer os.Remove(f.Name())

		require.NoError(t, err)

		token, err := getBearerToken(f.Name() + "-dummy")
		require.ErrorContains(t, err, "reading token file at")
		require.Empty(t, token)

		require.Nil(t, jwtCache)
	})

	tests := []struct {
		name          string
		claims        string // raw claims, will get base64-encoded
		encodedClaims string // treated as if it's base64-encoded string
		wantErrText   string
	}{
		{
			name:          "incorrect token structure",
			claims:        "",
			encodedClaims: "extra.parts.in.token",
			wantErrText:   "ServiceAccount JWT token expected to have 3 parts, not 6",
		},
		{
			name:          "broken base64",
			claims:        "",
			encodedClaims: "non-base64-string",
			wantErrText:   "base64 decoding ServiceAccount JWT token",
		},
		{
			name:          "broken json in claims",
			claims:        "{broken-json}",
			encodedClaims: "",
			wantErrText:   "deserializing ServiceAccount JWT claims",
		},
		{
			name:          "no exp in claims",
			claims:        "{}",
			encodedClaims: "",
			wantErrText:   "no expiry found in ServiceAccount JWT claims",
		},
		{
			name:          "broken exp in claims",
			claims:        `{"exp": "abc"}`,
			encodedClaims: "",
			wantErrText:   "token exp claim (expiry) cannot be cast to a float64",
		},
		{
			name:          "token not renewed",
			claims:        `{"exp": 1}`, // 01 Jan 1970
			encodedClaims: "",
			wantErrText:   "token expired at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwtCache = nil

			encodedClaims := base64.RawStdEncoding.EncodeToString([]byte(tt.claims))
			if tt.encodedClaims != "" {
				encodedClaims = tt.encodedClaims
			}

			jwt := fmt.Sprintf("header.%s.signature", encodedClaims)

			f := createFileWithContent(t, jwt)
			defer os.Remove(f.Name())

			token, err := getBearerToken(f.Name())
			require.ErrorContains(t, err, tt.wantErrText)
			require.Empty(t, token)

			require.Nil(t, jwtCache)
		})
	}

	t.Run("decode test token with nil cache", func(t *testing.T) {
		jwtCache = nil

		f, token := createTestJWTFile(t)
		defer os.Remove(f.Name())

		parsedToken, err := getBearerToken(f.Name())
		tokenIsValid(t, token, parsedToken, err)

		assert.NotNil(t, jwtCache)
	})

	t.Run("Read from cache", func(t *testing.T) {
		jwtCache = nil

		f, token := createTestJWTFile(t)
		defer os.Remove(f.Name())

		parsedToken, err := getBearerToken(f.Name())
		tokenIsValid(t, token, parsedToken, err)

		os.Remove(f.Name())
		cachedToken, err := getBearerToken(f.Name())
		tokenIsValid(t, token, cachedToken, err)

		assert.NotNil(t, jwtCache)
	})

	t.Run("expire cache and re-parse token", func(t *testing.T) {
		jwtCache = nil

		f, token := createTestJWTFile(t)
		defer os.Remove(f.Name())

		parsedToken, err := getBearerToken(f.Name())
		tokenIsValid(t, token, parsedToken, err)

		// Store original expiration and overwrite cache
		tokenExpiration := jwtCache.Expiration
		jwtCache.Expiration = time.Now().Add(-60 * time.Second)
		jwtCache.Token = ""

		cachedToken, err := getBearerToken(f.Name())
		tokenIsValid(t, token, cachedToken, err)
		require.Equal(t, tokenExpiration, jwtCache.Expiration)

		assert.NotNil(t, jwtCache)
	})
}
