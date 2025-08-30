package client

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	httptransport "github.com/go-openapi/runtime/client"
	genapi "github.com/grafana/grafana-openapi-client-go/client"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/metrics"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceAccountTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token" // nolint:gosec
)

type grafanaAdminCredentials struct {
	adminUser     string
	adminPassword string
	apikey        string
}

type JWTCache struct {
	Token      string
	Expiration time.Time
}

var jwtCache *JWTCache

// Revoke tokens early expecting them to be rotated hourly, see 'ExpirationSeconds' in KEP1205
// Should mitigate mid-reconcile expiration
const tokenExpirationCompensation = -30 * time.Second

// getBearerToken will read JWT token from given file and cache it until it expires.
// accepts filepath arg for testing
func getBearerToken(bearerTokenPath string) (string, error) {
	// Return cached token if not expired
	if jwtCache != nil && jwtCache.Expiration.After(time.Now()) {
		return jwtCache.Token, nil
	}

	b, err := os.ReadFile(bearerTokenPath)
	if err != nil {
		return "", fmt.Errorf("reading token file at %s, %w", bearerTokenPath, err)
	}

	token := string(b)

	// List of accepted JWT signing algorithms from: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#:~:text=oidc-signing-algs
	t, err := jwt.ParseSigned(token, []jose.SignatureAlgorithm{
		jose.RS256, jose.RS384, jose.RS512,
		jose.ES256, jose.ES384, jose.ES512,
		jose.PS256, jose.PS384, jose.PS512,
	})
	if err != nil {
		return "", err
	}

	claims := jwt.Claims{}

	// TODO fetch JWKS from https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT_HTTPS}/openid/v1/jwks
	// Then verify token using the keys
	err = t.UnsafeClaimsWithoutVerification(&claims)
	if err != nil {
		return "", fmt.Errorf("decoding ServiceAccount token %w", err)
	}

	tokenExpiration := claims.Expiry.Time()
	if tokenExpiration.Add(tokenExpirationCompensation).Before(time.Now()) {
		return "", fmt.Errorf("token expired at %s, expected %s to be renewed. Tokens are considered expired 30 seconds early", tokenExpiration.String(), bearerTokenPath)
	}

	jwtCache = &JWTCache{
		Token:      token,
		Expiration: tokenExpiration.Add(tokenExpirationCompensation),
	}

	return token, nil
}

func getExternalAdminUser(ctx context.Context, c client.Client, cr *v1beta1.Grafana) (string, error) {
	if cr.Spec.External != nil && cr.Spec.External.AdminUser != nil {
		adminUser, err := GetValueFromSecretKey(ctx, cr.Spec.External.AdminUser, c, cr.Namespace)
		if err != nil {
			return "", err
		}

		return string(adminUser), nil
	}

	adminUser := cr.GetConfigSectionValue("security", "admin_user")
	if adminUser != "" {
		return adminUser, nil
	}

	return "", fmt.Errorf("authentication undefined, set apiKey or userName for external instance: %s/%s", cr.Namespace, cr.Name)
}

func getExternalAdminPassword(ctx context.Context, c client.Client, cr *v1beta1.Grafana) (string, error) {
	if cr.Spec.External != nil && cr.Spec.External.AdminPassword != nil {
		adminPassword, err := GetValueFromSecretKey(ctx, cr.Spec.External.AdminPassword, c, cr.Namespace)
		if err != nil {
			return "", err
		}

		return string(adminPassword), nil
	}

	adminPassword := cr.GetConfigSectionValue("security", "admin_password")
	if adminPassword != "" {
		return adminPassword, nil
	}

	// If username is defined, we can assume apiKey will not be used
	return "", fmt.Errorf("password not set for external instance: %s/%s", cr.Namespace, cr.Name)
}

func getAdminCredentials(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*grafanaAdminCredentials, error) {
	credentials := &grafanaAdminCredentials{}

	if grafana.Spec.Client != nil && grafana.Spec.Client.UseKubeAuth {
		t, err := getBearerToken(serviceAccountTokenPath)
		if err != nil {
			return nil, err
		}

		credentials.apikey = t

		return credentials, nil
	}

	if grafana.IsExternal() {
		// prefer api key if present
		if grafana.Spec.External.APIKey != nil {
			apikey, err := GetValueFromSecretKey(ctx, grafana.Spec.External.APIKey, c, grafana.Namespace)
			if err != nil {
				return nil, err
			}

			credentials.apikey = string(apikey)

			return credentials, nil
		}

		var err error

		credentials.adminUser, err = getExternalAdminUser(ctx, c, grafana)
		if err != nil {
			return nil, err
		}

		credentials.adminPassword, err = getExternalAdminPassword(ctx, c, grafana)
		if err != nil {
			return nil, err
		}

		return credentials, nil
	}

	deployment := model.GetGrafanaDeployment(grafana, nil)
	selector := client.ObjectKey{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}

	err := c.Get(ctx, selector, deployment)
	if err != nil {
		return nil, err
	}

	for _, container := range deployment.Spec.Template.Spec.Containers {
		for _, env := range container.Env {
			if env.Name == config.GrafanaAdminUserEnvVar {
				if env.Value != "" {
					credentials.adminUser = env.Value
					continue
				}

				if env.ValueFrom != nil {
					if env.ValueFrom.SecretKeyRef != nil {
						usernameFromSecret, err := GetValueFromSecretKey(ctx, env.ValueFrom.SecretKeyRef, c, grafana.Namespace)
						if err != nil {
							return nil, err
						}

						credentials.adminUser = string(usernameFromSecret)
					}
				}
			}

			if env.Name == config.GrafanaAdminPasswordEnvVar {
				if env.Value != "" {
					credentials.adminPassword = env.Value
					continue
				}

				if env.ValueFrom != nil {
					if env.ValueFrom.SecretKeyRef != nil {
						passwordFromSecret, err := GetValueFromSecretKey(ctx, env.ValueFrom.SecretKeyRef, c, grafana.Namespace)
						if err != nil {
							return nil, err
						}

						credentials.adminPassword = string(passwordFromSecret)
					}
				}
			}
		}
	}

	return credentials, nil
}

func InjectAuthHeaders(ctx context.Context, c client.Client, grafana *v1beta1.Grafana, req *http.Request) error {
	creds, err := getAdminCredentials(ctx, c, grafana)
	if err != nil {
		return fmt.Errorf("fetching admin credentials: %w", err)
	}

	if creds.apikey != "" {
		req.Header.Set("Authorization", "Bearer "+creds.apikey)
	} else {
		req.SetBasicAuth(creds.adminUser, creds.adminPassword)
	}

	return nil
}

func ParseAdminURL(adminURL string) (*url.URL, error) {
	gURL, err := url.Parse(adminURL)
	if err != nil {
		return nil, fmt.Errorf("parsing url for client: %w", err)
	}

	if gURL.Host == "" {
		return nil, fmt.Errorf("invalid Grafana adminURL, url must contain protocol and host")
	}

	gURL = gURL.JoinPath("/api")

	return gURL, nil
}

func NewGeneratedGrafanaClient(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*genapi.GrafanaHTTPAPI, error) {
	var timeout time.Duration
	if grafana.Spec.Client != nil && grafana.Spec.Client.TimeoutSeconds != nil {
		timeout = max(time.Duration(*grafana.Spec.Client.TimeoutSeconds), 0)
	} else {
		timeout = 10
	}

	tlsConfig, err := buildTLSConfiguration(ctx, c, grafana)
	if err != nil {
		return nil, err
	}

	gURL, err := ParseAdminURL(grafana.Status.AdminURL)
	if err != nil {
		return nil, err
	}

	transport := NewInstrumentedRoundTripper(grafana.IsExternal(), tlsConfig, metrics.GrafanaAPIRequests.MustCurryWith(prometheus.Labels{
		"instance_namespace": grafana.Namespace,
		"instance_name":      grafana.Name,
	}))
	if grafana.Spec.Client != nil && grafana.Spec.Client.Headers != nil {
		transport.(*instrumentedRoundTripper).addHeaders(grafana.Spec.Client.Headers) //nolint:errcheck
	}

	// Secrets and ConfigMaps are not cached by default, get credentials as the last step.
	credentials, err := getAdminCredentials(ctx, c, grafana)
	if err != nil {
		return nil, err
	}

	cfg := &genapi.TransportConfig{
		Schemes:  []string{gURL.Scheme},
		BasePath: gURL.Path,
		Host:     gURL.Host,
		// APIKey is an optional API key or service account token.
		APIKey: credentials.apikey,
		// NumRetries contains the optional number of attempted retries
		NumRetries: 0,
		TLSConfig:  tlsConfig,
		Client: &http.Client{
			Transport: transport,
			Timeout:   timeout * time.Second,
		},
	}
	if credentials.adminUser != "" {
		cfg.BasicAuth = url.UserPassword(credentials.adminUser, credentials.adminPassword)
	}

	cl := genapi.NewHTTPClientWithConfig(nil, cfg)

	runtime, ok := cl.Transport.(*httptransport.Runtime)
	if !ok {
		return nil, fmt.Errorf("casting client transport into *httptransport.Runtime to overwrite the default context")
	}

	runtime.Context = ctx

	return cl, nil
}
