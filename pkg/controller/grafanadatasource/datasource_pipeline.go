package grafanadatasource

import (
	"github.com/ghodss/yaml"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

type DatasourcePipeline interface {
	ProcessDatasource(known *v1.ConfigMap) error
}

type DatasourcePipelineImpl struct {
	datasource *v1alpha1.GrafanaDataSource
	contents   string
	hash       string
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
	i.contents = string(bytes)
	return nil
}

func (i *DatasourcePipelineImpl) append(known *v1.ConfigMap) {
	if known.Data == nil {
		known.Data = map[string]string{}
	}

	known.Data[i.datasource.Filename()] = i.contents
}
