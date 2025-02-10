package content

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaClient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/controllers/content/fetchers"
	"github.com/grafana/grafana-operator/v5/embeds"
	v1 "k8s.io/api/core/v1"
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

func NewContentResolver(cr v1beta1.GrafanaContentResource, client client.Client, opts ...Option) (*ContentResolver, error) {
	// Perform these error checks once in initialization; we assume in the function calls
	// that the spec and status will be non-nil as a result.
	if cr.GrafanaContentSpec() == nil || cr.GrafanaContentStatus() == nil {
		return nil, fmt.Errorf("resource does not properly implement content spec or status fields; this indicates a bug in implementation")
	}

	resolver := &ContentResolver{
		Client:   client,
		resource: cr,
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return resolver, nil
}

func (h *ContentResolver) Resolve(ctx context.Context) (map[string]interface{}, string, error) {
	json, err := h.fetchContentJson(ctx)
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
func (h *ContentResolver) resolveDatasources(contentJson []byte) ([]byte, error) {
	spec := h.resource.GrafanaContentSpec()
	if len(spec.Datasources) == 0 {
		return contentJson, nil
	}

	for _, input := range spec.Datasources {
		if input.DatasourceName == "" || input.InputName == "" {
			return nil, fmt.Errorf("invalid datasource input rule in content resource %v/%v, input or datasource empty", h.resource.GetNamespace(), h.resource.GetName())
		}

		searchValue := fmt.Sprintf("${%s}", input.InputName)
		contentJson = bytes.ReplaceAll(contentJson, []byte(searchValue), []byte(input.DatasourceName))
	}

	return contentJson, nil
}

// fetchContentJson delegates obtaining the dashboard json definition to one of the known fetchers, for example
// from embedded raw json or from a url
func (h *ContentResolver) fetchContentJson(ctx context.Context) ([]byte, error) {
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
	case ContentSourceTypeRawJson:
		return []byte(spec.Json), nil
	case ContentSourceTypeGzipJson:
		return cache.Gunzip([]byte(spec.GzipJson))
	case ContentSourceTypeUrl:
		return fetchers.FetchDashboardFromUrl(ctx, h.resource, h.Client, grafanaClient.InsecureTLSConfiguration)
	case ContentSourceTypeJsonnet:
		envs, err := h.getDashboardEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}
		return fetchers.FetchJsonnet(h.resource, envs, embeds.GrafonnetEmbed)
	case ContentSourceJsonnetProject:
		envs, err := h.getDashboardEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}
		return fetchers.BuildProjectAndFetchJsonnetFrom(h.resource, envs)
	case ContentSourceTypeGrafanaCom:
		return fetchers.FetchDashboardFromGrafanaCom(ctx, h.resource, h.Client)
	case ContentSourceConfigMap:
		return fetchers.FetchDashboardFromConfigMap(h.resource, h.Client)
	default:
		return nil, fmt.Errorf("unknown source type %v found in content resource %v", sourceTypes[0], h.resource.GetName())
	}
}

func (h *ContentResolver) getDashboardEnvs(ctx context.Context) (map[string]string, error) {
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
				val, key, err := h.getReferencedValue(ctx, h.resource, ref.ValueFrom)
				if err != nil {
					return nil, fmt.Errorf("something went wrong processing referenced env %s, error: %w", ref.Name, err)
				}
				envs[key] = val
			}
		}
	}
	return envs, nil
}

func (h *ContentResolver) getReferencedValue(ctx context.Context, cr v1beta1.GrafanaContentResource, source v1beta1.GrafanaContentEnvFromSource) (string, string, error) {
	if source.SecretKeyRef != nil {
		s := &v1.Secret{}
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
		s := &v1.ConfigMap{}
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
func (h *ContentResolver) getContentModel(contentJson []byte) (map[string]interface{}, string, error) {
	contentJson, err := h.resolveDatasources(contentJson)
	if err != nil {
		return map[string]interface{}{}, "", err
	}

	hash := sha256.New()
	hash.Write(contentJson)

	var contentModel map[string]interface{}
	err = json.Unmarshal(contentJson, &contentModel)
	if err != nil {
		return map[string]interface{}{}, "", err
	}

	// NOTE: id should never be hardcoded in a model, otherwise grafana will try to update a model by id instead of uid.
	//       And, in case the id is non-existent, grafana will respond with 404. https://github.com/grafana/grafana-operator/issues/1108
	contentModel["id"] = nil

	uid, _ := contentModel["uid"].(string) //nolint:errcheck
	contentModel["uid"] = CustomUIDOrUID(h.resource, uid)

	return contentModel, fmt.Sprintf("%x", hash.Sum(nil)), nil
}
