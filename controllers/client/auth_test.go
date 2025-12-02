package client

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"testing/synctest"
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

func createFileWithContent(t *testing.T, content string) *os.File {
	t.Helper()

	f, err := os.CreateTemp(os.TempDir(), "test-*")
	require.NoError(t, err)

	written, err := f.WriteString(content)
	require.Equal(t, len([]byte(content)), written)
	require.NoError(t, err)

	return f
}

func getFakeToken(t *testing.T, exp time.Time) string {
	t.Helper()

	claims := fmt.Sprintf(`{"exp": %g}`, float64(exp.Unix()))
	encodedClaims := base64.RawStdEncoding.EncodeToString([]byte(claims))

	jwt := fmt.Sprintf("header.%s.signature", encodedClaims)

	return jwt
}

func tokenAndCacheAreValid(t *testing.T, got string, wantToken string, wantExp time.Time) {
	t.Helper()

	assert.NotEmpty(t, got)

	require.NotNil(t, jwtCache)
	assert.Equal(t, wantToken, got)
	assert.Equal(t, wantToken, jwtCache.Token)
	assert.Equal(t, wantExp.Unix(), jwtCache.Expiration.Unix())
	assert.False(t, wantExp.IsZero())
	assert.True(t, jwtCache.Expiration.After(time.Now()))
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

	t.Run("token expiration and renewal", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			jwtCache = nil

			now := time.Now()

			exp1 := now.Add(60 * time.Second)
			exp2 := now.Add(120 * time.Second)

			jwt1 := getFakeToken(t, exp1)
			jwt2 := getFakeToken(t, exp2)

			f1 := createFileWithContent(t, jwt1)
			defer os.Remove(f1.Name())

			f2 := createFileWithContent(t, jwt2)
			defer os.Remove(f2.Name())

			// Empty cache at first, we expect it to be populated with the data derived from jwt1
			wantToken := jwt1
			wantExp := exp1.Add(tokenExpirationCompensation)

			got, err := getBearerToken(f1.Name())
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)

			// New token is already available, but the old one hasn't expired yet, so fetch jwt1 from cache
			wantToken = jwt1
			wantExp = exp1.Add(tokenExpirationCompensation)

			got, err = getBearerToken(f2.Name()) // returns the old token despite the new file
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)

			// Token expiration and renewal
			time.Sleep(60 * time.Second)
			synctest.Wait()

			wantToken = jwt2
			wantExp = exp2.Add(tokenExpirationCompensation)

			got, err = getBearerToken(f2.Name()) // the new token this time
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)
		})
	})
}
