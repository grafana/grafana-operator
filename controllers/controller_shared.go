package controllers

import (
	"context"
	"fmt"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getMatchingInstances(ctx context.Context, k8sClient client.Client, labelSelector *v1.LabelSelector) (v1beta1.GrafanaList, error) {
	selector, err := v1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return v1beta1.GrafanaList{}, err
	}
	var list v1beta1.GrafanaList
	opts := []client.ListOption{client.MatchingLabelsSelector{Selector: selector}}

	err = k8sClient.List(ctx, &list, opts...)
	if err != nil {
		return v1beta1.GrafanaList{}, fmt.Errorf("failed to get grafana instances matching %v: %w", labelSelector, err)
	}

	if len(list.Items) == 0 {
		return v1beta1.GrafanaList{}, fmt.Errorf("no matching grafana instances matching: %v", labelSelector)
	}

	return list, err
}

func updateGrafanaStatusPlugins(ctx context.Context, k8sClient client.Client, grafana *v1beta1.Grafana, plugins v1beta1.PluginList) error {
	plugins, err := grafana.Status.Plugins.ConsolidatedConcat(plugins)
	if err != nil {
		return fmt.Errorf("failed to extend plugin list with plugins %v: %w", plugins, err)
	}

	grafana.Status.Plugins = plugins
	if err := k8sClient.Status().Update(ctx, grafana); err != nil {
		return fmt.Errorf("failed to update plugin list in grafana status: %w", err)
	}

	return nil
}

func grafanaOwnedResources(object client.Object) bool {
	return getGrafanaOwner(object.GetOwnerReferences()) != nil
}

func deploymentReady(object client.Object) bool {
	deploy := object.(*appsv1.Deployment)
	return deploy.Status.ReadyReplicas > 0
}

func getGrafanaOwner(ownerReferences []v1.OwnerReference) *v1.OwnerReference {
	for _, owner := range ownerReferences {
		if owner.APIVersion == v1beta1.GroupVersion.String() && owner.Kind == "Grafana" {
			return &owner
		}
	}
	return nil
}
