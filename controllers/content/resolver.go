package content

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/content/cache"
	"github.com/grafana/grafana-operator/v5/controllers/content/fetchers"
	"github.com/grafana/grafana-operator/v5/embeds"
	"github.com/itchyny/gojq"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Resolver struct {
	Client          client.Client
	resource        v1beta1.GrafanaContentResource
	disabledSources []SourceType
}

type Option func(r *Resolver)

func WithDisabledSources(disabledSources []SourceType) Option {
	return func(r *Resolver) {
		r.disabledSources = disabledSources
	}
}

func NewResolver(cr v1beta1.GrafanaContentResource, cl client.Client, opts ...Option) *Resolver {
	resolver := &Resolver{
		Client:   cl,
		resource: cr,
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return resolver
}

func (h *Resolver) Resolve(ctx context.Context) (map[string]any, string, error) {
	j, err := h.fetchContentJSON(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch contents: %w", err)
	}

	model, hash, err := h.getContentModel(j)
	if err != nil {
		return nil, "", fmt.Errorf("failed to extract model: %w", err)
	}

	return model, hash, nil
}

// map data sources that are required in the content model to data sources that exist in the instance
func (h *Resolver) resolveDatasources(contentJSON []byte) ([]byte, error) {
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
func (h *Resolver) fetchContentJSON(ctx context.Context) ([]byte, error) {
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
	case SourceTypeRawJSON:
		return []byte(spec.JSON), nil
	case SourceTypeGzipJSON:
		return cache.Gunzip(spec.GzipJSON)
	case SourceTypeURL:
		return fetchers.FetchFromURL(ctx, h.resource, h.Client, grafanaclient.InsecureTLSConfiguration)
	case SourceTypeJsonnet:
		envs, err := h.getContentEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}

		return fetchers.FetchJsonnet(h.resource, envs, embeds.GrafonnetEmbed)
	case SourceJsonnetProject:
		envs, err := h.getContentEnvs(ctx)
		if err != nil {
			return nil, fmt.Errorf("something went wrong while collecting envs, error: %w", err)
		}

		return fetchers.BuildProjectAndFetchJsonnetFrom(ctx, h.resource, envs)
	case SourceTypeGrafanaCom:
		return fetchers.FetchFromGrafanaCom(ctx, h.resource, h.Client)
	case SourceConfigMap:
		return fetchers.FetchDashboardFromConfigMap(ctx, h.resource, h.Client)
	case SourceOCI:
		return fetchers.FetchFromOCI(ctx, h.resource, h.Client)
	default:
		return nil, fmt.Errorf("unknown source type %v found in content resource %v", sourceTypes[0], h.resource.GetName())
	}
}

func (h *Resolver) getContentEnvs(ctx context.Context) (map[string]string, error) {
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

func (h *Resolver) getReferencedValue(ctx context.Context, cr v1beta1.GrafanaContentResource, source v1beta1.GrafanaContentEnvFromSource) (string, string, error) {
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

// variableOverrideScript sets the template variable named $name to $value and reconciles its
// options[] so the value is selected and valid. The type guards leave dashboards that omit or
// reshape these fields untouched instead of erroring.
const variableOverrideScript = `
def override($v): {selected: true, text: $v, value: $v};

def reconcileOptions($v):
  if (.options | type) != "array" then .
  else
    (.options |= map(if type == "object" then .selected = (.value == $v) else . end))
    | if any(.options[]; type == "object" and .value == $v) then .
      else .options = [override($v)] + .options
      end
  end;

if (.templating | type) == "object" and (.templating.list | type) == "array"
then .templating.list |= map(
  if type == "object" and .name == $name
  then .current = override($value) | reconcileOptions($value)
  else .
  end)
else .
end`

// variableOverrideCode compiles variableOverrideScript once; the script is a constant, so a failure
// here is a bug in this package rather than something a user can provoke.
var variableOverrideCode = sync.OnceValues(func() (*gojq.Code, error) {
	query, err := gojq.Parse(variableOverrideScript)
	if err != nil {
		return nil, fmt.Errorf("parsing variable override script: %w", err)
	}

	code, err := gojq.Compile(query, gojq.WithVariables([]string{"$name", "$value"}))
	if err != nil {
		return nil, fmt.Errorf("compiling variable override script: %w", err)
	}

	return code, nil
})

// contentVariables returns the template variable overrides declared by the resource, or nil for
// content resources that have no variables to override.
func (h *Resolver) contentVariables() []v1beta1.GrafanaContentVariable {
	overrider, ok := h.resource.(v1beta1.GrafanaContentVariableOverrider)
	if !ok {
		return nil
	}

	return overrider.ContentVariables()
}

// overrideVariables sets the default (current) value of named template variables in the content model.
func (h *Resolver) overrideVariables(contentModel map[string]any) (map[string]any, error) {
	variables := h.contentVariables()
	if len(variables) == 0 {
		return contentModel, nil
	}

	code, err := variableOverrideCode()
	if err != nil {
		return nil, err
	}

	for _, v := range variables {
		contentModel, err = applyVariableOverride(code, contentModel, v)
		if err != nil {
			return nil, fmt.Errorf("overriding template variable %q: %w", v.Name, err)
		}
	}

	return contentModel, nil
}

// applyVariableOverride runs variableOverrideScript for a single override. The script has no
// generators at the top level, so it always yields exactly one model.
func applyVariableOverride(code *gojq.Code, contentModel map[string]any, v v1beta1.GrafanaContentVariable) (map[string]any, error) {
	result, ok := code.Run(contentModel, v.Name, v.Value).Next()
	if !ok {
		return nil, errors.New("script returned no result")
	}

	if err, ok := result.(error); ok {
		return nil, fmt.Errorf("evaluating script: %w", err)
	}

	model, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("script returned %T, expected a json object", result)
	}

	return model, nil
}

// getContentModel resolves datasources, updates uid (if needed) and converts raw json to type grafana client accepts
func (h *Resolver) getContentModel(contentJSON []byte) (map[string]any, string, error) {
	contentJSON, err := h.resolveDatasources(contentJSON)
	if err != nil {
		return map[string]any{}, "", err
	}

	hash := sha256.New()
	hash.Write(contentJSON)
	// Variable overrides change the model after it is parsed, so fold them into the hash explicitly.
	for _, v := range h.contentVariables() {
		hash.Write([]byte(v.Name))
		hash.Write([]byte{0})
		hash.Write([]byte(v.Value))
		hash.Write([]byte{0})
	}

	var contentModel map[string]any

	err = json.Unmarshal(contentJSON, &contentModel)
	if err != nil {
		return map[string]any{}, "", err
	}

	contentModel, err = h.overrideVariables(contentModel)
	if err != nil {
		return map[string]any{}, "", err
	}

	// NOTE: id should never be hardcoded in a model, otherwise grafana will try to update a model by id instead of uid.
	//       And, in case the id is non-existent, grafana will respond with 404. https://github.com/grafana/grafana-operator/issues/1108
	contentModel["id"] = nil

	uid, _ := contentModel["uid"].(string) //nolint:errcheck
	contentModel["uid"] = GetGrafanaUID(h.resource, uid)

	return contentModel, fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (h *Resolver) UpdateCache(model map[string]any) error {
	// GetSourceTypes needs to be of length 1 for this function to even be called
	sourceType := GetSourceTypes(h.resource)[0]

	// only cache remote resources
	if sourceType != SourceTypeURL && sourceType != SourceTypeGrafanaCom && sourceType != SourceOCI {
		return nil
	}

	return cache.SetContentCache(h.resource, model)
}
