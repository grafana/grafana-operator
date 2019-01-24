package v1alpha1

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/client/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type GrafanaDashboardsGetter interface {
	GrafanaDashboards(namespace string) GrafanaDashboardInterface
}

type GrafanaDashboardInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.GrafanaDashboardList, error)
	Update(dashboard *v1alpha1.GrafanaDashboard) (result *v1alpha1.GrafanaDashboard, err error)
}

type grafanadashboards struct {
	client rest.Interface
	ns string
}

func newGrafanaDashboards(c *IntegreatlyV1alpha1Client, namespace string) *grafanadashboards {
	return &grafanadashboards{
		client: c.RESTClient(),
		ns: namespace,
	}
}

func (c *grafanadashboards) List(opts metav1.ListOptions) (result *v1alpha1.GrafanaDashboardList, err error) {
	result = &v1alpha1.GrafanaDashboardList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("grafanadashboards").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

func (c *grafanadashboards) Update(dashboard *v1alpha1.GrafanaDashboard) (result *v1alpha1.GrafanaDashboard, err error) {
	result = &v1alpha1.GrafanaDashboard{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("grafanadashboards").
		Name(dashboard.Name).
		Body(dashboard).
		Do().
		Into(result)
	return
}