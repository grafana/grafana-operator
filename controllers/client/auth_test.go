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
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/resources"
	"github.com/grafana/grafana-operator/v5/pkg/tk8s"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestGetContainerEnvCredentials(t *testing.T) {
	const (
		username    = "root"
		usernameKey = "user"
		password    = "secret"
		passwordKey = "password"
		secretName  = "grafana-credentials" //nolint:gosec
		nonExistent = "non-existent"
	)

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      secretName,
		},
		Data: map[string][]byte{
			usernameKey: []byte(username),
			passwordKey: []byte(password),
		},
	}

	ctx := t.Context()
	s := runtime.NewScheme()

	err := corev1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	err = appsv1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	err = v1beta1.AddToScheme(s)
	require.NoError(t, err, "adding scheme")

	c := fake.NewClientBuilder().
		WithScheme(s).
		WithObjects(secret).Build()

	t.Run("non-existent deployment", func(t *testing.T) {
		cr := &v1beta1.Grafana{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      "grafana-env-credentials",
			},
		}

		got, err := getContainerEnvCredentials(ctx, c, cr)
		require.ErrorContains(t, err, "not found")

		assert.Nil(t, got)
	})

	tests := []struct {
		name        string
		envs        []corev1.EnvVar
		want        *grafanaAdminCredentials
		wantErrText string
	}{
		{
			name: "plaintext",
			envs: []corev1.EnvVar{
				{
					Name:  config.GrafanaAdminUserEnvVar,
					Value: username,
				},
				{
					Name:  config.GrafanaAdminPasswordEnvVar,
					Value: password,
				},
			},
			want: &grafanaAdminCredentials{
				adminUser:     username,
				adminPassword: password,
			},
		},
		{
			name: "no credential envs",
			envs: []corev1.EnvVar{
				{
					Name:  "a",
					Value: "b",
				},
			},
			want: &grafanaAdminCredentials{},
		},
		{
			name: "from secret",
			envs: []corev1.EnvVar{
				{
					Name:      config.GrafanaAdminUserEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, usernameKey),
				},
				{
					Name:      config.GrafanaAdminPasswordEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, passwordKey),
				},
			},
			want: &grafanaAdminCredentials{
				adminUser:     username,
				adminPassword: password,
			},
		},
		// error cases
		{
			name: "non-existent secret",
			envs: []corev1.EnvVar{
				{
					Name:      config.GrafanaAdminUserEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, nonExistent, usernameKey),
				},
				{
					Name:      config.GrafanaAdminPasswordEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, nonExistent, passwordKey),
				},
			},
			want:        nil,
			wantErrText: "not found",
		},
		{
			name: "non-existent username key",
			envs: []corev1.EnvVar{
				{
					Name:      config.GrafanaAdminUserEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, nonExistent),
				},
				{
					Name:      config.GrafanaAdminPasswordEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, passwordKey),
				},
			},
			want:        nil,
			wantErrText: "credentials not found in secret",
		},
		{
			name: "non-existent password key",
			envs: []corev1.EnvVar{
				{
					Name:      config.GrafanaAdminUserEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, usernameKey),
				},
				{
					Name:      config.GrafanaAdminPasswordEnvVar,
					ValueFrom: tk8s.GetEnvVarSecretSource(t, secretName, nonExistent),
				},
			},
			want:        nil,
			wantErrText: "credentials not found in secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "grafana-env-credentials",
				},
			}

			deployment := resources.GetGrafanaDeployment(cr, nil)
			deployment.Spec.Template.Spec.Containers = []corev1.Container{
				{
					Name: "grafana", // TODO: switch to const once it's done inside getContainerEnvCredentials
					Env:  tt.envs,
				},
			}

			createAndCleanupResources(t, ctx, c, []client.Object{
				cr, deployment,
			})

			got, err := getContainerEnvCredentials(ctx, c, cr)
			if tt.wantErrText == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErrText)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func createAndCleanupResources(t *testing.T, ctx context.Context, c client.WithWatch, objects []client.Object) {
	t.Helper()

	for _, obj := range objects {
		err := c.Create(ctx, obj)
		require.NoError(t, err)

		t.Cleanup(func() {
			c.Delete(ctx, obj) //nolint:errcheck
		})
	}
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

	t.Cleanup(func() {
		os.Remove(f.Name())
	})

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
			f2 := createFileWithContent(t, jwt2)

			// Empty cache at first, we expect it to be populated with the data derived from jwt1
			wantToken := jwt1
			wantExp := exp1.Add(tokenExpirationCompensation)

			got, err := getBearerToken(f1.Name())
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)

			// New token is already available, but the old one hasn't expired yet, so we fetch jwt1 from cache
			wantToken = jwt1
			wantExp = exp1.Add(tokenExpirationCompensation)

			got, err = getBearerToken(f2.Name()) // returns the old token despite the new file
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)

			// Token expiration and renewal
			time.Sleep(60 * time.Second)

			wantToken = jwt2
			wantExp = exp2.Add(tokenExpirationCompensation)

			got, err = getBearerToken(f2.Name()) // returns the new token this time
			require.NoError(t, err)

			tokenAndCacheAreValid(t, got, wantToken, wantExp)
		})
	})
}
