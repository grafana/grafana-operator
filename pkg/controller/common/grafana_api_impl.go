package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/grafana-tools/sdk"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"os"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var grafanalog = logf.Log.WithName("grafana_helper")

type GrafanaApiData struct {
	*GrafanaData
	clients map[string]*sdk.Client
	config  *ControllerConfig
}

func NewGrafanaApiImpl(grafanaData *GrafanaData) *GrafanaApiData {
	// initially supports only 1 grafana in same namespace as operator
	// in future might support many different grafana instances in many namespaces
	return &GrafanaApiData{
		GrafanaData: grafanaData,
		config:      GetControllerConfig(),
		clients:     make(map[string]*sdk.Client),
	}
}

func (h GrafanaApiData) UpdateDashboard(ctx context.Context, dashboard *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	client, err := h.getClient(dashboard.Spec.GrafanaRef)
	if err != nil {
		return resp, err
	}
	board := sdk.Board{}
	if err := json.Unmarshal([]byte(dashboard.Spec.Dashboard.Json), &board); err != nil {
		fmt.Printf("Failed to unmarshal dashboard: %v", err)
		return resp, err
	}
	dashboardResp, err := client.SetDashboard(board, true)
	if err != nil {
		fmt.Printf("Failed to set dashboard: %v", err)
		return resp, err
	}
	dashboard.Status.Slug = *dashboardResp.Slug
	resp.Succeeded = true
	return resp, nil
}

func (h GrafanaApiData) DashboardIsKnown(ctx context.Context, dashboard *v1alpha1.GrafanaDashboard) (bool, error) {
	client, err := h.getClient(dashboard.Spec.GrafanaRef)
	if err != nil {
		return false, err
	}
	_, _, err = client.GetDashboard(dashboard.Status.Slug)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (h GrafanaApiData) DeleteDashboard(ctx context.Context, dashboard *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	client, err := h.getClient(dashboard.Spec.GrafanaRef)
	if err != nil {
		return resp, err
	}
	_, err = client.DeleteDashboard(dashboard.Status.Slug)
	if err != nil {
		return resp, err
	}
	resp.Succeeded = true
	return resp, nil
}

func (h GrafanaApiData) GetDashboard(ctx context.Context, dashboard *v1alpha1.GrafanaDashboard) (ResponseStatus, error) {
	resp := ResponseStatus{}
	return resp, errors.New("not implemented")
}

func (h GrafanaApiData) UpdateDataSource(ctx context.Context, datasource *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	client, err := h.getClient(datasource.Spec.GrafanaRef)
	if err != nil {
		return resp, err
	}

	ds := dataSourceFromApi(datasource)
	datasourceResp, err := client.CreateDatasource(ds)
	if err != nil {
		fmt.Printf("Failed to set dashboard: %v", err)
		return resp, err
	}
	fmt.Printf("response == %s\n", *datasourceResp.Message)

	resp.Succeeded = true
	datasource.Status.ID = *(datasourceResp.ID)
	return resp, nil
}
func (h GrafanaApiData) GetDataSource(ctx context.Context, datasource *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	client, err := h.getClient(datasource.Spec.GrafanaRef)
	if err != nil {
		return resp, err
	}
	_, err = client.GetDatasourceByName(datasource.Spec.DataSource.Name)
	if err != nil {
		return resp, err
	}
	return resp, errors.New("not implemented")
}
func (h GrafanaApiData) DeleteDataSource(ctx context.Context, datasource *v1alpha1.GrafanaDataSource) (ResponseStatus, error) {
	resp := ResponseStatus{}
	client, err := h.getClient(datasource.Spec.GrafanaRef)
	if err != nil {
		return resp, err
	}
	_, err = client.DeleteDatasource(datasource.Status.ID)
	if err != nil {
		fmt.Printf("Failed to set dashboard: %+v", err)
		return resp, err
	}

	resp.Succeeded = true
	return resp, nil
}

func (h GrafanaApiData) DataSourceIsKnown(ctx context.Context, datasource *v1alpha1.GrafanaDataSource) (bool, error) {
	// custom resource holds multiple datasources,
	// DataSourceIsKnown return true only if all datasources are present in api
	client, err := h.getClient(datasource.Spec.GrafanaRef)
	if err != nil {
		return false, err
	}
	_, err = client.GetDatasourceByName(datasource.Spec.DataSource.Name)
	if err != nil {
		fmt.Printf("Datasource is not Knonwn, err = %v\n", err)
		return false, nil
	}
	fmt.Printf("Datasource is Knonwn")
	return true, nil
}

func dataSourceFromApi(from *v1alpha1.GrafanaDataSource) sdk.Datasource {
	return sdk.Datasource{
		OrgID:             1,
		Name:              from.Spec.DataSource.Name,
		Type:              from.Spec.DataSource.Type,
		Access:            from.Spec.DataSource.Access,
		URL:               from.Spec.DataSource.Url,
		Password:          &from.Spec.DataSource.Password,
		User:              &from.Spec.DataSource.User,
		Database:          &from.Spec.DataSource.Database,
		BasicAuth:         &from.Spec.DataSource.BasicAuth,
		BasicAuthUser:     &from.Spec.DataSource.BasicAuthUser,
		BasicAuthPassword: &from.Spec.DataSource.BasicAuthPassword,
		IsDefault:         from.Spec.DataSource.IsDefault,
		JSONData:          from.Spec.DataSource.JsonData,
	}
}

func (h GrafanaApiData) getClient(grafanaName string) (*sdk.Client, error) {
	// currently supports grafana in same namespace as operator only

	if val, ok := h.clients[grafanaName]; ok {
		return val, nil
	}
	grafana, err := h.kubeHelper.grafanaClient.Grafanas(os.Getenv("WATCH_NAMESPACE")).Get(grafanaName, v1.GetOptions{})
	if err != nil {
		fmt.Printf("Failed to get grafana %#v", err)
		return nil, err
	}
	url := "http://" + GrafanaServiceName + "." + grafana.Namespace + ":3000/"
	creds := grafana.Spec.AdminUser + ":" + grafana.Spec.AdminPassword
	client := sdk.NewClient(url, creds, &http.Client{})
	h.clients[grafanaName] = client
	return client, nil
}
