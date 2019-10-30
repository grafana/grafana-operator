package v1alpha1

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type GrafanaV1Alpha1Interface interface {
	Grafanas(namespace string) GrafanaInterface
}

type GrafanaV1Alpha1Client struct {
	restClient rest.Interface
}

func NewForConfigOrDie(c *rest.Config) *GrafanaV1Alpha1Client {
	config := *c
	config.ContentConfig.GroupVersion = &schema.GroupVersion{Group: v1alpha1.GroupName, Version: v1alpha1.GroupVersion}
	config.APIPath = "/apis"
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	config.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		panic(err)
	}
	return &GrafanaV1Alpha1Client{restClient: client}
}

func (c *GrafanaV1Alpha1Client) Grafanas(namespace string) GrafanaInterface {
	return &grafanaClient{
		restClient: c.restClient,
		ns:         namespace,
	}
}
