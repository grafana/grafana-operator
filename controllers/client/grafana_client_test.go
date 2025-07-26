package client

import (
	"context"
	"testing"

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
		tests := []struct {
			name string
			spec v1beta1.GrafanaSpec
		}{
			{
				name: "err from empty config",
				spec: v1beta1.GrafanaSpec{
					External: &v1beta1.External{},
					Config: map[string]map[string]string{
						"security": {},
					},
				},
			},
			{
				name: "err when reference is unset or security.admin_user/password is set",
				spec: v1beta1.GrafanaSpec{
					External: &v1beta1.External{},
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
					Spec: tt.spec,
				}

				adminUser, err := getExternalAdminUser(testCtx, client, cr)
				require.Error(t, err)
				assert.Empty(t, adminUser)

				adminPassword, err := getExternalAdminPassword(testCtx, client, cr)
				require.Error(t, err)
				assert.Empty(t, adminPassword)
			})
		}
	})
}

// TODO currently only tests code paths for external grafanas
func TestGetAdminCredentials(t *testing.T) {
	tests := []struct {
		name            string
		spec            v1beta1.GrafanaSpec
		wantCredentials *grafanaAdminCredentials
		wantErr         bool
	}{
		{
			name: "ApiKey from Secret ignoring config",
			spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{
					APIKey: &v1.SecretKeySelector{
						LocalObjectReference: v1.LocalObjectReference{
							Name: "grafana-credentials",
						},
						Key: "token",
					},
				},
				Config: map[string]map[string]string{
					"security": {
						"admin_user":     "root",
						"admin_password": "secret",
					},
				},
			},
			wantCredentials: &grafanaAdminCredentials{
				adminUser:     "",
				adminPassword: "",
				apikey:        "service-account-key",
			},
			wantErr: false,
		},
		{
			name: "fallback to admin user/password",
			spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{},
				Config: map[string]map[string]string{
					"security": {
						"admin_user":     "root",
						"admin_password": "secret",
					},
				},
			},
			wantCredentials: &grafanaAdminCredentials{
				adminUser:     "root",
				adminPassword: "secret",
				apikey:        "",
			},
			wantErr: false,
		},
		{
			name: "err when neither apiKey or admin user/password is set",
			spec: v1beta1.GrafanaSpec{
				External: &v1beta1.External{},
			},
			wantCredentials: nil,
			wantErr:         true,
		},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.Grafana{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "default",
					Name:      "grafana",
				},
				Spec: tt.spec,
			}

			credentials, err := getAdminCredentials(testCtx, client, cr)
			if tt.wantErr {
				require.Error(t, err, "getAdminCredentials() should return an error")
				require.Nil(t, credentials, "credentials should be nil on error")
			} else {
				require.NoError(t, err, "getAdminCredentials() should not return an error")
				require.Equal(t, tt.wantCredentials.apikey, credentials.apikey)
				require.Equal(t, tt.wantCredentials.adminUser, credentials.adminUser)
				require.Equal(t, tt.wantCredentials.adminPassword, credentials.adminPassword)
			}
		})
	}
}
