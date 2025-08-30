package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestParseAdminURL(t *testing.T) {
	tests := []struct {
		name      string
		adminURL  string
		wantHost  string
		wantPath  string
		wantError bool
	}{
		{
			name:      "No Path",
			adminURL:  "https://grafana.example.com",
			wantHost:  "grafana.example.com",
			wantPath:  "api",
			wantError: false,
		},
		{
			name:      "Root as Path",
			adminURL:  "https://grafana.example.com/",
			wantHost:  "grafana.example.com",
			wantPath:  "/api",
			wantError: false,
		},
		{
			name:      "Custom Port",
			adminURL:  "https://grafana.example.com:3000/",
			wantHost:  "grafana.example.com:3000",
			wantPath:  "/api",
			wantError: false,
		},
		{
			name:      "No Path and no Scheme",
			adminURL:  "grafana.example.com",
			wantError: true,
		},
		{
			name:      "No Scheme",
			adminURL:  "grafana.example.com/path",
			wantError: true,
		},
		{
			name:      "Custom Path",
			adminURL:  "https://grafana.example.com/instances/1",
			wantHost:  "grafana.example.com",
			wantPath:  "/instances/1/api",
			wantError: false,
		},
		{
			name:      "Relative Custom Path",
			adminURL:  "https://grafana.example.com/../test",
			wantHost:  "grafana.example.com",
			wantPath:  "/test/api",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAdminURL(tt.adminURL)
			if tt.wantError {
				assert.Error(t, err, "This should be an invalid url input")
			} else {
				require.NoError(t, err, "This should be a valid url")
				assert.Equal(t, tt.wantPath, got.Path, "Path does not match")
				assert.Equal(t, tt.wantHost, got.Host, "Host does not match")
				assert.Contains(t, got.Path, "api", "/api is not appended to path correctly")
			}
		})
	}
}

func TestGetExternalAdminCredentials(t *testing.T) {
	credSecret := &v1.Secret{
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
	err := v1.AddToScheme(s)
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
						AdminUser: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
								Name: "grafana-credentials",
							},
							Key: "user",
						},
						AdminPassword: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
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
						AdminPassword: &v1.SecretKeySelector{
							LocalObjectReference: v1.LocalObjectReference{
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
	credSecret := &v1.Secret{
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
	err := v1.AddToScheme(s)
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
					APIKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
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

func TestGetBearerToken(t *testing.T) {
	t.Parallel()

	now := time.Now()
	body := jwt.Claims{
		ID:        "1234",
		Issuer:    "grafana.operator.com",
		Audience:  jwt.Audience{"https://grafana.operator.com"},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Expiry:    jwt.NewNumericDate(now.Add(time.Duration(60 * float64(time.Second)))),
	}

	// Generate key
	testRSAKey, err := rsa.GenerateKey(rand.Reader, 1024) // nolint:gosec
	require.NotNil(t, testRSAKey)
	require.NoError(t, err)

	sk := jose.SigningKey{
		Key: &jose.JSONWebKey{
			Key:   testRSAKey,
			KeyID: "test-key",
		},
		Algorithm: jose.RS256,
	}

	// Create signer, and sign token
	signer, err := jose.NewSigner(sk, &jose.SignerOptions{EmbedJWK: true})
	require.NoError(t, err)

	origToken, err := jwt.Signed(signer).Claims(body).Serialize()
	require.NoError(t, err)

	// Create tmp file
	tokenFile, err := os.CreateTemp(os.TempDir(), "token-*")
	require.NoError(t, err)
	require.NotNil(t, testRSAKey)

	// Do not read token file purposefully
	jwtCache = nil
	noToken, err := getBearerToken(tokenFile.Name() + "-extra")
	require.Error(t, err)
	require.Empty(t, noToken)
	require.Nil(t, jwtCache)

	// Write token to file
	writtenBytes, err := tokenFile.WriteString(origToken)
	require.Equal(t, len([]byte(origToken)), writtenBytes)
	require.NoError(t, err)

	// Decode token
	token, err := getBearerToken(tokenFile.Name())
	require.NoError(t, err)
	require.Equal(t, origToken, token)
	require.Equal(t, origToken, jwtCache.Token)
	require.False(t, jwtCache.Expiration.IsZero())

	// Expire the cache and re-read token
	jwtCache.Expiration = time.Now().Add(-10 * time.Second)
	token, err = getBearerToken(tokenFile.Name())
	require.NoError(t, err)
	require.Equal(t, origToken, token)
	require.True(t, jwtCache.Expiration.After(time.Now()))

	// Mangle token
	reversedOrigToken := []byte(origToken)
	slices.Reverse(reversedOrigToken)
	reversedWrittenBytes, err := tokenFile.Write(reversedOrigToken)
	require.NoError(t, err)
	require.Equal(t, len([]byte(reversedOrigToken)), reversedWrittenBytes)

	// Successfully get token from cache
	token, err = getBearerToken(tokenFile.Name())
	require.NoError(t, err)
	require.Equal(t, origToken, token)

	// Reset cache and error on mangled token
	jwtCache = nil
	mangledToken, err := getBearerToken(tokenFile.Name())
	require.Error(t, err)
	require.Empty(t, mangledToken)
	require.Nil(t, jwtCache)
}
