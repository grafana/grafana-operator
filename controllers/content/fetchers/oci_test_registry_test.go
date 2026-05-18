package fetchers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

// fakeRegistry is a minimal in-process implementation of the OCI distribution
// pull API. It supports only the routes oras-go's Resolve + FetchBytes hit, and
// its state is populated directly via pushArtifact (no upload protocol).
type fakeRegistry struct {
	blobs     map[string][]byte // "sha256:..." -> content
	manifests map[string][]byte // "<repo>/<reference>" -> manifest bytes (ref is tag OR digest)

	requireAuth bool
	user, pass  string
}

// newFakeRegistry starts an httptest server backed by an empty registry and
// returns the registry handle and the host:port the server is reachable on.
func newFakeRegistry(t *testing.T) (*fakeRegistry, string) {
	t.Helper()

	r := &fakeRegistry{
		blobs:     map[string][]byte{},
		manifests: map[string][]byte{},
	}

	srv := httptest.NewServer(r.handler())
	t.Cleanup(srv.Close)

	return r, strings.TrimPrefix(srv.URL, "http://")
}

// pushArtifact stores an oras-style artifact: each file becomes a raw blob
// layer annotated with org.opencontainers.image.title. Returns the manifest
// digest so digest-pinned tests can reference it.
func (r *fakeRegistry) pushArtifact(t *testing.T, repo, tag string, files map[string][]byte) string {
	t.Helper()

	layers := make([]ocispec.Descriptor, 0, len(files))

	for name, content := range files {
		d := sha256Digest(content)
		r.blobs[d] = content
		layers = append(layers, ocispec.Descriptor{
			MediaType:   "application/json",
			Digest:      digest.Digest(d),
			Size:        int64(len(content)),
			Annotations: map[string]string{ocispec.AnnotationTitle: name},
		})
	}

	emptyJSON := []byte("{}")
	emptyDigest := sha256Digest(emptyJSON)
	r.blobs[emptyDigest] = emptyJSON

	manifest := ocispec.Manifest{
		MediaType:    ocispec.MediaTypeImageManifest,
		ArtifactType: "application/vnd.grafana.dashboard+json",
		Config: ocispec.Descriptor{
			MediaType: ocispec.MediaTypeEmptyJSON,
			Digest:    digest.Digest(emptyDigest),
			Size:      int64(len(emptyJSON)),
		},
		Layers: layers,
	}
	manifest.SchemaVersion = 2

	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)

	manifestDigest := sha256Digest(manifestBytes)
	r.manifests[repo+"/"+tag] = manifestBytes
	r.manifests[repo+"/"+manifestDigest] = manifestBytes

	return manifestDigest
}

func (r *fakeRegistry) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if r.requireAuth {
			user, pass, ok := req.BasicAuth()
			if !ok || user != r.user || pass != r.pass {
				w.Header().Set("WWW-Authenticate", `Basic realm="registry"`)
				http.Error(w, "unauthorized", http.StatusUnauthorized)

				return
			}
		}

		path := req.URL.Path

		if path == "/v2/" || path == "/v2" {
			w.Header().Set("Content-Type", "application/json")
			writeOrFail(w, []byte("{}"))

			return
		}

		if !strings.HasPrefix(path, "/v2/") {
			http.NotFound(w, req)
			return
		}

		rest := strings.TrimPrefix(path, "/v2/")

		switch {
		case strings.Contains(rest, "/manifests/"):
			idx := strings.LastIndex(rest, "/manifests/")
			repo := rest[:idx]
			ref := rest[idx+len("/manifests/"):]
			r.serveManifest(w, req, repo, ref)
		case strings.Contains(rest, "/blobs/"):
			idx := strings.LastIndex(rest, "/blobs/")
			ref := rest[idx+len("/blobs/"):]
			r.serveBlob(w, req, ref)
		default:
			http.NotFound(w, req)
		}
	})
}

func (r *fakeRegistry) serveManifest(w http.ResponseWriter, req *http.Request, repo, ref string) {
	body, ok := r.manifests[repo+"/"+ref]
	if !ok {
		http.NotFound(w, req)
		return
	}

	d := sha256Digest(body)

	w.Header().Set("Content-Type", ocispec.MediaTypeImageManifest)
	w.Header().Set("Docker-Content-Digest", d)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))

	if req.Method == http.MethodHead {
		return
	}

	writeOrFail(w, body)
}

func (r *fakeRegistry) serveBlob(w http.ResponseWriter, req *http.Request, ref string) {
	body, ok := r.blobs[ref]
	if !ok {
		http.NotFound(w, req)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Docker-Content-Digest", ref)
	w.Header().Set("Content-Length", strconv.Itoa(len(body)))

	if req.Method == http.MethodHead {
		return
	}

	writeOrFail(w, body)
}

// writeOrFail writes to the response, panicking on error so the test surfaces
// a server-side write failure rather than the linter complaining about a
// silently dropped error.
func writeOrFail(w http.ResponseWriter, body []byte) {
	if _, err := w.Write(body); err != nil {
		panic(err)
	}
}

func sha256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
