package common

import (
	"context"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DataSourcesState struct {
	ClusterDataSources *v1alpha1.GrafanaDataSourceList
	KnownDataSources   *v1.ConfigMap
}

func NewDataSourcesState() *DataSourcesState {
	return &DataSourcesState{}
}

func (i *DataSourcesState) Read(ctx context.Context, client client.Client, ns string) error {
	err := i.readClusterDataSources(ctx, client, ns)
	if err != nil {
		return err
	}

	err = i.readKnownDataSources(ctx, client, ns)
	if err != nil {
		return err
	}

	return nil
}

func (i *DataSourcesState) readClusterDataSources(ctx context.Context, c client.Client, ns string) error {
	list := &v1alpha1.GrafanaDataSourceList{}
	opts := &client.ListOptions{
		Namespace: ns,
	}

	err := c.List(ctx, list, opts)
	if err != nil {
		i.ClusterDataSources = list
		return err
	}

	i.ClusterDataSources = list
	return nil
}

func (i *DataSourcesState) readKnownDataSources(ctx context.Context, c client.Client, ns string) error {
	dataSources := &v1.ConfigMap{}
	selector := client.ObjectKey{
		Namespace: ns,
		Name:      model.GrafanaDatasourcesConfigMapName,
	}

	err := c.Get(ctx, selector, dataSources)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return err
	}

	i.KnownDataSources = dataSources

	return nil
}
