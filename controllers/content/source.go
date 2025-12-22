package content

import "github.com/grafana/grafana-operator/v5/api/v1beta1"

type SourceType string

const (
	SourceTypeRawJSON    SourceType = "json"
	SourceTypeGzipJSON   SourceType = "gzipJson"
	SourceJsonnetProject SourceType = "jsonnetProjectWithRuntimeRaw"
	SourceTypeURL        SourceType = "url"
	SourceTypeJsonnet    SourceType = "jsonnet"
	SourceTypeGrafanaCom SourceType = "grafana"
	SourceConfigMap      SourceType = "configmap"
)

func GetSourceTypes(cr v1beta1.GrafanaContentResource) []SourceType {
	var sourceTypes []SourceType

	spec := cr.GrafanaContentSpec()

	if spec.JSON != "" {
		sourceTypes = append(sourceTypes, SourceTypeRawJSON)
	}

	if spec.GzipJSON != nil {
		sourceTypes = append(sourceTypes, SourceTypeGzipJSON)
	}

	if spec.URL != "" {
		sourceTypes = append(sourceTypes, SourceTypeURL)
	}

	if spec.Jsonnet != "" {
		sourceTypes = append(sourceTypes, SourceTypeJsonnet)
	}

	if spec.GrafanaCom != nil {
		sourceTypes = append(sourceTypes, SourceTypeGrafanaCom)
	}

	if spec.ConfigMapRef != nil {
		sourceTypes = append(sourceTypes, SourceConfigMap)
	}

	if spec.JsonnetProjectBuild != nil {
		sourceTypes = append(sourceTypes, SourceJsonnetProject)
	}

	return sourceTypes
}
