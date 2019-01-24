package grafana

import (
	"github.com/ghodss/yaml"
	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ResourceHelper struct {
	templateHelper *GrafanaTemplateHelper
	cr             *integreatly.Grafana
}

func newResourceHelper(cr *integreatly.Grafana) *ResourceHelper {
	return &ResourceHelper{
		templateHelper: newTemplateHelper(cr),
		cr:             cr,
	}
}

func (r *ResourceHelper) createResource(template string) (runtime.Object, error) {
	tpl, err := r.templateHelper.loadTemplate(template)
	if err != nil {
		return nil, err
	}

	resource := unstructured.Unstructured{}
	err = yaml.Unmarshal(tpl, &resource)

	if err != nil {
		return nil, err
	}

	return &resource, nil
}