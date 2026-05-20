package fetchers

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"slices"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	corev1 "k8s.io/api/core/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

func FetchFromOCI(ctx context.Context, cr v1beta1.GrafanaContentResource, cl client.Client) ([]byte, error) {
	o := cr.GrafanaContentSpec().OCI

	parsed, err := registry.ParseReference(o.Reference)
	if err != nil {
		return nil, fmt.Errorf("parse oci reference %q: %w", o.Reference, err)
	}

	if parsed.Reference == "" {
		return nil, fmt.Errorf("oci reference must include tag or digest on %v/%v", cr.GetNamespace(), cr.GetName())
	}

	repo, err := remote.NewRepository(parsed.Registry + "/" + parsed.Repository)
	if err != nil {
		return nil, fmt.Errorf("parse oci reference %q: %w", o.Reference, err)
	}

	if o.InsecurePlainHTTP {
		repo.PlainHTTP = true
	}

	var credFunc auth.CredentialFunc

	if o.PullSecretRef != nil {
		credFunc, err = authFromPullSecret(ctx, cl, cr.GetNamespace(), o.PullSecretRef.Name, repo.Reference.Registry)
		if err != nil {
			return nil, fmt.Errorf("resolve pull secret: %w", err)
		}
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewCache(),
		Credential: credFunc,
	}

	_, manifestBytes, err := oras.FetchBytes(ctx, repo, parsed.Reference, oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, fmt.Errorf("pull oci manifest %s: %w", o.Reference, err)
	}

	var manifest ocispec.Manifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest for %s: %w", o.Reference, err)
	}

	target := filepath.ToSlash(o.Path)

	// Pass 1: oras-style artifact. Each file is its own raw-blob layer carrying an
	// org.opencontainers.image.title annotation; match the annotation to the requested path.
	for _, layer := range manifest.Layers {
		title := layer.Annotations[ocispec.AnnotationTitle]
		if filepath.ToSlash(title) != target {
			continue
		}

		blob, err := content.FetchAll(ctx, repo, layer)
		if err != nil {
			return nil, fmt.Errorf("fetch layer %s for %s: %w", layer.Digest, o.Reference, err)
		}

		return blob, nil
	}

	// Pass 2: container-image style. Each layer is a (gzipped) tarball of a filesystem.
	// Walk layers in reverse so upper layers win, matching the container filesystem
	// overlay semantics. Non-tar layers (e.g. an oras-style raw blob that did not match
	// pass 1) are silently skipped.
	for _, layer := range slices.Backward(manifest.Layers) {
		data, found, err := fetchFileFromImageLayer(ctx, repo, layer, target)
		if err != nil {
			return nil, fmt.Errorf("scan layer %s of %s: %w", layer.Digest, o.Reference, err)
		}

		if found {
			return data, nil
		}
	}

	return nil, fmt.Errorf("file %q not found in %s", o.Path, o.Reference)
}

// fetchFileFromImageLayer streams a single layer blob and looks for target inside it,
// treating the blob as a (possibly gzipped) tarball. Returns (data, true, nil) on hit,
// (nil, false, nil) when the layer is not a tar or simply does not contain target.
func fetchFileFromImageLayer(ctx context.Context, repo *remote.Repository, layer ocispec.Descriptor, target string) ([]byte, bool, error) {
	rc, err := repo.Fetch(ctx, layer)
	if err != nil {
		return nil, false, err
	}
	defer rc.Close()

	var tarReader io.Reader = rc

	if isGzipMediaType(layer.MediaType) {
		gz, gzErr := gzip.NewReader(rc)
		if gzErr != nil {
			return nil, false, nil
		}

		defer gz.Close()

		tarReader = gz
	}

	tr := tar.NewReader(tarReader)

	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil, false, nil
		}

		if err != nil {
			// Not a tar (e.g. raw artifact blob). Treat as not-found in this layer.
			return nil, false, nil
		}

		if h.Typeflag != tar.TypeReg {
			continue
		}

		name := strings.TrimPrefix(filepath.ToSlash(h.Name), "/")
		name = strings.TrimPrefix(name, "./")

		if name == target {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tr); err != nil { // #nosec G110 - operator pulls from trusted registries configured by cluster admins
				return nil, false, err
			}

			return buf.Bytes(), true, nil
		}
	}
}

// isGzipMediaType reports whether an OCI layer media type denotes gzip-wrapped content.
func isGzipMediaType(mt string) bool {
	return strings.HasSuffix(mt, "+gzip") || strings.HasSuffix(mt, ".gz") || strings.HasSuffix(mt, ".gzip")
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

func authFromPullSecret(ctx context.Context, cl client.Client, namespace, secretName, registryHost string) (auth.CredentialFunc, error) {
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

		return auth.StaticCredential(registryHost, auth.Credential{Username: username, Password: password}), nil
	}

	return nil, nil
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
