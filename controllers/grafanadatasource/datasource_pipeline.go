package grafanadatasource

import (
	"github.com/ghodss/yaml"
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type DatasourcePipeline interface {
	ProcessDatasource(known *v1.ConfigMap) error
}

type DatasourcePipelineImpl struct {
	datasource *v1alpha1.GrafanaDataSource
	contents   string
}

func NewDatasourcePipeline(ds *v1alpha1.GrafanaDataSource) DatasourcePipeline {
	return &DatasourcePipelineImpl{
		datasource: ds,
	}
}

func (i *DatasourcePipelineImpl) ProcessDatasource(known *v1.ConfigMap) error {
	// parse the datasource to make sure it is value
	err := i.parse()
	if err != nil {
		return errors.Wrap(err, "error parsing datasource")
	}

	// append the datasource to the configmap
	i.append(known)
	return nil
}

func applyCustomJsonData(datasource v1alpha1.GrafanaDataSourceFields, unmarshaledDatasource map[string]interface{}) (needsRemarshaling bool) {
	if datasource.CustomJsonData != nil {
		needsRemarshaling = true
		unmarshaledDatasource["jsonData"] = unmarshaledDatasource["customJsonData"]
		delete(unmarshaledDatasource, "customJsonData")
	}
	if datasource.CustomSecureJsonData != nil {
		needsRemarshaling = true
		unmarshaledDatasource["secureJsonData"] = unmarshaledDatasource["customSecureJsonData"]
		delete(unmarshaledDatasource, "customSecureJsonData")
	}
	return
}

func (i *DatasourcePipelineImpl) parse() error {
	datasources := struct {
		ApiVersion  int                                `json:"apiVersion"`
		Datasources []v1alpha1.GrafanaDataSourceFields `json:"datasources"`
	}{
		ApiVersion:  DatasourcesApiVersion,
		Datasources: i.datasource.Spec.Datasources,
	}

	bytes, err := yaml.Marshal(datasources)
	if err != nil {
		return err
	}
	unmarshaledDatasources := make(map[string]interface{})
	if err = yaml.Unmarshal(bytes, &unmarshaledDatasources); err != nil {
		return err
	}
	unmarshaledDatasourceList, ok := unmarshaledDatasources["datasources"].([]interface{})
	needsRemarshaling := false
	if ok {
		for ix, datasource := range i.datasource.Spec.Datasources {
			unmarshaledDatasource, ok := unmarshaledDatasourceList[ix].(map[string]interface{})
			if !ok {
				continue
			}
			needsRemarshaling = needsRemarshaling || applyCustomJsonData(datasource, unmarshaledDatasource)
		}
	}
	if needsRemarshaling {
		bytes, err = yaml.Marshal(unmarshaledDatasources)
		if err != nil {
			return err
		}
	}
	i.contents = string(bytes)
	return nil
}

func (i *DatasourcePipelineImpl) append(known *v1.ConfigMap) {
	if known.Data == nil {
		known.Data = map[string]string{}
	}

	known.Data[i.datasource.Filename()] = i.contents
}
