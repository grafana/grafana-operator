package fetchers

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

func FetchFromOCI(ctx context.Context, cr v1beta1.GrafanaContentResource, cl client.Client) ([]byte, error) {
	o := cr.GrafanaContentSpec().OCI

	var refStr string

	switch {
	case o.Digest != "":
		refStr = fmt.Sprintf("%s@%s", o.Image, o.Digest)
	case o.Tag != "":
		refStr = fmt.Sprintf("%s:%s", o.Image, o.Tag)
	default:
		return nil, fmt.Errorf("oci source must specify tag or digest on %v/%v", cr.GetNamespace(), cr.GetName())
	}

	var nameOpts []name.Option

	if o.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}

	ref, err := name.ParseReference(refStr, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parse oci reference %q: %w", refStr, err)
	}

	auth := authn.Anonymous

	if o.PullSecretRef != nil {
		auth, err = authFromPullSecret(ctx, cl, cr.GetNamespace(), o.PullSecretRef.Name, ref.Context())
		if err != nil {
			return nil, fmt.Errorf("resolve pull secret: %w", err)
		}
	}

	opts := []remote.Option{remote.WithContext(ctx), remote.WithAuth(auth)}

	if o.Insecure {
		opts = append(opts, remote.WithTransport(insecureTransport()))
	}

	img, err := remote.Image(ref, opts...)
	if err != nil {
		return nil, fmt.Errorf("pull oci image %s: %w", refStr, err)
	}

	rc := mutate.Extract(img)
	defer rc.Close()

	data, err := extractFromTar(rc, o.File)
	if err != nil {
		return nil, fmt.Errorf("extract %q from %s: %w", o.File, refStr, err)
	}

	return data, nil
}

func extractFromTar(r io.Reader, target string) ([]byte, error) {
	tr := tar.NewReader(r)

	for {
		h, err := tr.Next()

		if err == io.EOF {
			return nil, fmt.Errorf("file %q not found", target)
		}

		if err != nil {
			return nil, err
		}

		if h.Typeflag != tar.TypeReg {
			continue
		}

		if filepath.ToSlash(h.Name) == target {
			var buf bytes.Buffer

			if _, err := io.Copy(&buf, tr); err != nil { // #nosec G110 - operator pulls from trusted registries configured by cluster admins
				return nil, err
			}

			return buf.Bytes(), nil
		}
	}
}

// dockerConfigJSON mirrors the relevant subset of the kubernetes.io/dockerconfigjson secret format.
type dockerConfigJSON struct {
	Auths map[string]dockerConfigAuth `json:"auths"`
}

type dockerConfigAuth struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"` // base64("username:password")
}

func authFromPullSecret(ctx context.Context, cl client.Client, namespace, secretName string, repo name.Repository) (authn.Authenticator, error) {
	secret := &corev1.Secret{}

	if err := cl.Get(ctx, client.ObjectKey{Namespace: namespace, Name: secretName}, secret); err != nil {
		return nil, err
	}

	if secret.Type != corev1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("pull secret %s/%s must be type %s, got %s",
			namespace, secretName, corev1.SecretTypeDockerConfigJson, secret.Type)
	}

	raw, ok := secret.Data[corev1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("pull secret %s/%s missing key %s", namespace, secretName, corev1.DockerConfigJsonKey)
	}

	var cfg dockerConfigJSON

	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse pull secret %s/%s: %w", namespace, secretName, err)
	}

	registryHost := repo.RegistryStr()

	for host, a := range cfg.Auths {
		if !hostMatches(host, registryHost) {
			continue
		}

		username, password := a.Username, a.Password

		if a.Auth != "" {
			decoded, err := base64.StdEncoding.DecodeString(a.Auth)
			if err != nil {
				return nil, fmt.Errorf("decode auth field in pull secret %s/%s: %w", namespace, secretName, err)
			}

			parts := strings.SplitN(string(decoded), ":", 2)
			if len(parts) == 2 {
				username, password = parts[0], parts[1]
			}
		}

		return authn.FromConfig(authn.AuthConfig{Username: username, Password: password}), nil
	}

	return authn.Anonymous, nil
}

func insecureTransport() http.RoundTripper {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultTransport
	}

	t := base.Clone()
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 - Linter disabled because InsecureSkipVerify is the wanted behavior for this variable

	return t
}

// hostMatches returns true when the config host key matches the registry hostname.
// Docker config files may store the registry as "https://index.docker.io/v1/" for Docker Hub,
// or plain hostname for other registries.
func hostMatches(configHost, registryHost string) bool {
	if !strings.Contains(configHost, "://") {
		configHost = "https://" + configHost
	}

	u, err := url.Parse(configHost)
	if err != nil {
		return false
	}

	return u.Host == registryHost
}
