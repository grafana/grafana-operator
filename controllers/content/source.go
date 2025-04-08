package content

import "github.com/grafana/grafana-operator/v5/api/v1beta1"

type ContentSourceType string

const (
	ContentSourceTypeRawJSON    ContentSourceType = "json"
	ContentSourceTypeGzipJSON   ContentSourceType = "gzipJson"
	ContentSourceJsonnetProject ContentSourceType = "jsonnetProjectWithRuntimeRaw"
	ContentSourceTypeURL        ContentSourceType = "url"
	ContentSourceTypeJsonnet    ContentSourceType = "jsonnet"
	ContentSourceTypeGrafanaCom ContentSourceType = "grafana"
	ContentSourceConfigMap      ContentSourceType = "configmap"
)

func GetSourceTypes(cr v1beta1.GrafanaContentResource) []ContentSourceType {
	var sourceTypes []ContentSourceType

	spec := cr.GrafanaContentSpec()

	if spec.JSON != "" {
		sourceTypes = append(sourceTypes, ContentSourceTypeRawJSON)
	}

	if spec.GzipJSON != nil {
		sourceTypes = append(sourceTypes, ContentSourceTypeGzipJSON)
	}

	if spec.URL != "" {
		sourceTypes = append(sourceTypes, ContentSourceTypeURL)
	}

	if spec.Jsonnet != "" {
		sourceTypes = append(sourceTypes, ContentSourceTypeJsonnet)
	}

	if spec.GrafanaCom != nil {
		sourceTypes = append(sourceTypes, ContentSourceTypeGrafanaCom)
	}

	if spec.ConfigMapRef != nil {
		sourceTypes = append(sourceTypes, ContentSourceConfigMap)
	}

	if spec.JsonnetProjectBuild != nil {
		sourceTypes = append(sourceTypes, ContentSourceJsonnetProject)
	}

	return sourceTypes
}
