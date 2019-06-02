package grafana

import (
	"github.com/ghodss/yaml"
	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// An abstraction type that allows easier access to the contents of an
// unstructured resource
type UnstructuredResourceMap struct {
	Values map[string]interface{}
}

// Go one level deeper in the unstructured resource map
func (m UnstructuredResourceMap) access(key string) *UnstructuredResourceMap {
	return &UnstructuredResourceMap{
		Values: m.Values[key].(map[string]interface{}),
	}
}

// Return value of the current level in the unstructured resource map
func (m UnstructuredResourceMap) get(key string) interface{} {
	return m.Values[key]
}

// Set value in an unstructured resource map
func (m UnstructuredResourceMap) set(key string, value interface{}) {
	m.Values[key] = value
}

// Create a new unstructured resource map
func newUnstructuredResourceMap(unstructured *unstructured.Unstructured) *UnstructuredResourceMap {
	return &UnstructuredResourceMap{
		Values: unstructured.UnstructuredContent(),
	}
}

// Helps with creating kubernetes resources from yaml templates
type ResourceHelper struct {
	templateHelper *TemplateHelper
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
