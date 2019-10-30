package v1alpha1

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type GrafanaInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.GrafanaList, error)
	Get(name string, options metav1.GetOptions) (*v1alpha1.Grafana, error)
}

type grafanaClient struct {
	restClient rest.Interface
	ns         string
}

func (c *grafanaClient) List(opts metav1.ListOptions) (*v1alpha1.GrafanaList, error) {
	result := v1alpha1.GrafanaList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("grafanas").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)
	return &result, err
}

func (c *grafanaClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.Grafana, error) {
	result := v1alpha1.Grafana{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("grafanas").
		Name(name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(&result)
	return &result, err
}
