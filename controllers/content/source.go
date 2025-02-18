package content

import "github.com/grafana/grafana-operator/v5/api/v1beta1"

type ContentSourceType string

const (
	ContentSourceTypeRawJson    ContentSourceType = "json"
	ContentSourceTypeGzipJson   ContentSourceType = "gzipJson"
	ContentSourceJsonnetProject ContentSourceType = "jsonnetProjectWithRuntimeRaw"
	ContentSourceTypeUrl        ContentSourceType = "url"
	ContentSourceTypeJsonnet    ContentSourceType = "jsonnet"
	ContentSourceTypeGrafanaCom ContentSourceType = "grafana"
	ContentSourceConfigMap      ContentSourceType = "configmap"
)

func GetSourceTypes(cr v1beta1.GrafanaContentResource) []ContentSourceType {
	var sourceTypes []ContentSourceType

	spec := cr.GrafanaContentSpec()

	if spec.Json != "" {
		sourceTypes = append(sourceTypes, ContentSourceTypeRawJson)
	}

	if spec.GzipJson != nil {
		sourceTypes = append(sourceTypes, ContentSourceTypeGzipJson)
	}

	if spec.Url != "" {
		sourceTypes = append(sourceTypes, ContentSourceTypeUrl)
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
