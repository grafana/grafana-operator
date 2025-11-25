package content

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/controllers/content/fetchers"
	"github.com/grafana/grafana-operator/v5/embeds"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsUpdatedUID(cr v1beta1.GrafanaContentResource, uid string) bool {
	status := cr.GrafanaContentStatus()
	// This indicates an implementation error
	if status == nil {
		return false
	}

	// Resource has just been created, status is not yet updated
	if status.UID == "" {
		return false
	}

	uid = CustomUIDOrUID(cr, uid)

	return status.UID != uid
}

// Wrapper around CustomUID, contentModelUID or default metadata.uid
func CustomUIDOrUID(cr v1beta1.GrafanaContentResource, contentUID string) string {
	if spec := cr.GrafanaContentSpec(); spec != nil {
		if spec.CustomUID != "" {
			return spec.CustomUID
		}
	}

	if contentUID != "" {
		return contentUID
	}

	return string(cr.GetUID())
}

func HasChanged(cr v1beta1.GrafanaContentResource, hash string) bool {
	return !Unchanged(cr, hash)
}

// Unchanged checks if the stored content hash on the status matches the input
func Unchanged(cr v1beta1.GrafanaContentResource, hash string) bool {
	status := cr.GrafanaContentStatus()
	// This indicates an implementation error
	if status == nil {
		return true
	}

	return status.Hash == hash
}

type ContentResolver struct {
	Client          client.Client
	resource        v1beta1.GrafanaContentResource
	disabledSources []ContentSourceType
}

type Option func(r *ContentResolver)

func WithDisabledSources(disabledSources []ContentSourceType) Option {
	return func(r *ContentResolver) {
		r.disabledSources = disabledSources
	}
}

func NewContentResolver(cr v1beta1.GrafanaContentResource, client client.Client, opts ...Option) *ContentResolver {
	resolver := &ContentResolver{
		Client:   client,
		resource: cr,
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return resolver
}

func (h *ContentResolver) Resolve(ctx context.Context) (map[string]any, string, error) {
	json, err := h.fetchContentJSON(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch contents: %w", err)
	}

	model, hash, err := h.getContentModel(json)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract model: %w", err)
	}

	return model, hash, nil
}

// map data sources that are required in the content model to data sources that exist in the instance
func (h *ContentResolver) resolveDatasources(contentJSON []byte) ([]byte, error) {
	spec := h.resource.GrafanaContentSpec()
	if len(spec.Datasources) == 0 {
		return contentJSON, nil
	}

	for _, input := range spec.Datasources {
		if input.DatasourceName == "" || input.InputName == "" {
			return nil, fmt.Errorf("invalid datasource input rule in content resource %v/%v, input or datasource empty", h.resource.GetNamespace(), h.resource.GetName())
		}

		searchValue := fmt.Sprintf("${%s}", input.InputName)
		contentJSON = bytes.ReplaceAll(contentJSON, []byte(searchValue), []byte(input.DatasourceName))
	}

	return contentJSON, nil
}

// fetchContentJSON delegates obtaining the json definition to one of the known fetchers, for example
// from embedded raw json or from a url
func (h *ContentResolver) fetchContentJSON(ctx context.Context) ([]byte, error) {
	sourceTypes := GetSourceTypes(h.resource)

	if len(sourceTypes) == 0 {
		return nil, fmt.Errorf("no source type provided for content resource %v", h.resource.GetName())
	}

	if len(sourceTypes) > 1 {
		return nil, fmt.Errorf("more than one source types found for content resource %v", h.resource.GetName())
	}

	if slices.Contains(h.disabledSources, sourceTypes[0]) {
		return nil, fmt.Errorf("source type %v is disabled for content resource %v", sourceTypes[0], h.resource.GetName())
	}

	spec := h.resource.GrafanaContentSpec()

	switch sourceTypes[0] {
	case ContentSourceTypeRawJSON:
		return []byte(spec.JSON), nil
	case ContentSourceTypeGzipJSON:
		return cache.Gunzip([]byte(spec.GzipJSON))
	case ContentSourceTypeURL:
		return fetchers.FetchFromURL(ctx, h.resource, h.Client, grafanaclient.InsecureTLSConfiguration)
	case ContentSourceTypeJsonnet:
		envs, err := h.getContentEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}

		return fetchers.FetchJsonnet(h.resource, envs, embeds.GrafonnetEmbed)
	case ContentSourceJsonnetProject:
		envs, err := h.getContentEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}

		return fetchers.BuildProjectAndFetchJsonnetFrom(h.resource, envs)
	case ContentSourceTypeGrafanaCom:
		return fetchers.FetchFromGrafanaCom(ctx, h.resource, h.Client)
	case ContentSourceConfigMap:
		return fetchers.FetchDashboardFromConfigMap(h.resource, h.Client)
	default:
		return nil, fmt.Errorf("unknown source type %v found in content resource %v", sourceTypes[0], h.resource.GetName())
	}
}

func (h *ContentResolver) getContentEnvs(ctx context.Context) (map[string]string, error) {
	spec := h.resource.GrafanaContentSpec()

	envs := make(map[string]string)

	if spec.EnvsFrom != nil {
		for _, ref := range spec.EnvsFrom {
			key, val, err := h.getReferencedValue(ctx, h.resource, ref)
			if err != nil {
				return nil, fmt.Errorf("something went wrong processing envs, error: %w", err)
			}

			envs[key] = val
		}
	}

	if spec.Envs != nil {
		for _, ref := range spec.Envs {
			if ref.Value != "" {
				envs[ref.Name] = ref.Value
			} else {
				_, val, err := h.getReferencedValue(ctx, h.resource, ref.ValueFrom)
				if err != nil {
					return nil, fmt.Errorf("something went wrong processing referenced env %s, error: %w", ref.Name, err)
				}

				envs[ref.Name] = val
			}
		}
	}

	return envs, nil
}

func (h *ContentResolver) getReferencedValue(ctx context.Context, cr v1beta1.GrafanaContentResource, source v1beta1.GrafanaContentEnvFromSource) (string, string, error) {
	if source.SecretKeyRef != nil {
		s := &corev1.Secret{}

		err := h.Client.Get(ctx, client.ObjectKey{Namespace: cr.GetNamespace(), Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}

		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return source.SecretKeyRef.Key, string(val), nil
		} else {
			return "", "", fmt.Errorf("missing key %s in secret %s", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
		}
	}

	if source.ConfigMapKeyRef != nil {
		s := &corev1.ConfigMap{}

		err := h.Client.Get(ctx, client.ObjectKey{Namespace: cr.GetNamespace(), Name: source.ConfigMapKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}

		if val, ok := s.Data[source.ConfigMapKeyRef.Key]; ok {
			return source.ConfigMapKeyRef.Key, val, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in configmap %s", source.ConfigMapKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	}

	return "", "", fmt.Errorf("source couldn't be parsed source: %s", source)
}

// getContentModel resolves datasources, updates uid (if needed) and converts raw json to type grafana client accepts
func (h *ContentResolver) getContentModel(contentJSON []byte) (map[string]any, string, error) {
	contentJSON, err := h.resolveDatasources(contentJSON)
	if err != nil {
		return map[string]any{}, "", err
	}

	hash := sha256.New()
	hash.Write(contentJSON)

	var contentModel map[string]any

	err = json.Unmarshal(contentJSON, &contentModel)
	if err != nil {
		return map[string]any{}, "", err
	}

	// NOTE: id should never be hardcoded in a model, otherwise grafana will try to update a model by id instead of uid.
	//       And, in case the id is non-existent, grafana will respond with 404. https://github.com/grafana/grafana-operator/issues/1108
	contentModel["id"] = nil

	uid, _ := contentModel["uid"].(string) //nolint:errcheck
	contentModel["uid"] = CustomUIDOrUID(h.resource, uid)

	return contentModel, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
