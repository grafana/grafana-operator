package templates

import (
	"context"

	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	"github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type handler struct {
	*api.Runtime
}

func NewHandler(rt *api.Runtime) *handler {
	return &handler{rt}
}

func (d *handler) getTemplateList(object runtime.Object) (err error) {
	if err = d.Client.List(context.Background(),
		object,
		client.MatchingLabels{"grafana": "templates"},
		client.InNamespace(d.Namespace.Namespace),
	); err != nil {
		return
	}
	return
}

func (d *handler) CopyDashboards(namespace string) (err error) {
	dl := &v1alpha1.GrafanaDashboardList{}
	if err = d.getTemplateList(dl); err != nil {
		return
	}
	for _, i := range dl.Items {
		ds := i.DeepCopy()
		ds.SetNamespace(namespace)
		ds.SetResourceVersion("")
		if err = d.Client.Create(context.Background(), ds); err != nil {
			if err, ok := err.(*errors.StatusError); ok {
				if err := d.Client.Delete(context.Background(), ds); err == nil {
					return d.CopyDashboards(namespace)
				}
				return err
			}
		}
	}
	return
}

func (d *handler) CopyDatasources(namespace string, principal *models.Principal) (err error) {
	dl := &v1alpha1.GrafanaDataSourceList{}
	if err = d.getTemplateList(dl); err != nil {
		return
	}
	oc, err := getServiceClient(principal.AccountName, principal.Domain)
	if err != nil {
		return
	}
	roles := []applicationcredentials.Role{
		applicationcredentials.Role{Name: "monitoring_viewer"},
	}

	ac, err := createApplicationCredentials(&oc, principal.ID, roles)
	if err != nil {
		return
	}
	for i := 0; i < len(dl.Items); i++ {
		ds := dl.Items[i].DeepCopy()
		ds.SetNamespace(namespace)
		ds.SetResourceVersion("")
		if err = d.createDatasources(ds, ac); err != nil {
			return
		}
	}
	return
}

func (d *handler) createDatasources(ds *v1alpha1.GrafanaDataSource, ac *applicationcredentials.ApplicationCredential) (err error) {
	for i := 0; i < len(ds.Spec.Datasources); i++ {
		ds.Spec.Datasources[i].OrgId = 1
		if ds.Spec.Datasources[i].BasicAuth {
			ds.Spec.Datasources[i].BasicAuthUser = "*" + ac.ID
			ds.Spec.Datasources[i].BasicAuthPassword = ac.Secret
		}
	}
	if err = d.Client.Create(context.Background(), ds); err != nil {
		if err, ok := err.(*errors.StatusError); ok {
			if err := d.Client.Delete(context.Background(), ds); err == nil {
				return d.createDatasources(ds, ac)
			}
			return err
		}
	}
	return
}
