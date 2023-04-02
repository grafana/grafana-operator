package controllers

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/model"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetMatchingInstances(ctx context.Context, k8sClient client.Client, labelSelector *v1.LabelSelector) (v1beta1.GrafanaList, error) {
	selector, err := v1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return v1beta1.GrafanaList{}, err
	}
	var list v1beta1.GrafanaList
	opts := []client.ListOption{client.MatchingLabelsSelector{selector}}

	err = k8sClient.List(ctx, &list, opts...)
	return list, err
}

func ReconcilePlugins(ctx context.Context, k8sClient client.Client, scheme *runtime.Scheme, grafana *v1beta1.Grafana, plugins v1beta1.PluginList, resource string) error {
	pluginsConfigMap := model.GetPluginsConfigMap(grafana, scheme)
	selector := client.ObjectKey{
		Namespace: pluginsConfigMap.Namespace,
		Name:      pluginsConfigMap.Name,
	}

	err := k8sClient.Get(ctx, selector, pluginsConfigMap)
	if err != nil {
		return err
	}

	val, err := json.Marshal(plugins.Sanitize())
	if err != nil {
		return err
	}

	if pluginsConfigMap.BinaryData == nil {
		pluginsConfigMap.BinaryData = make(map[string][]byte)
	}

	if !bytes.Equal(val, pluginsConfigMap.BinaryData[resource]) {
		pluginsConfigMap.BinaryData[resource] = val
		return k8sClient.Update(ctx, pluginsConfigMap)
	}

	return nil
}

func grafanaOwnedResources(object client.Object) bool {
	for _, owner := range object.GetOwnerReferences() {
		if owner.APIVersion == "grafana.integreatly.org/v1beta1" && owner.Kind == "Grafana" {
			return true
		}
	}
	return false
}

func deploymentReady(object client.Object) bool {
	deploy := object.(*appsv1.Deployment)
	return deploy.Status.ReadyReplicas > 0
}
