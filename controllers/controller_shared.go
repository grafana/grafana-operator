package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"slices"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const grafanaFinalizer = "operator.grafana.com/finalizer"

const (
	conditionNoMatchingInstance = "NoMatchingInstance"
	conditionNoMatchingFolder   = "NoMatchingFolder"
)

func GetMatchingInstances(ctx context.Context, k8sClient client.Client, labelSelector *v1.LabelSelector) (v1beta1.GrafanaList, error) {
	if labelSelector == nil {
		return v1beta1.GrafanaList{}, nil
	}

	var list v1beta1.GrafanaList
	opts := []client.ListOption{
		client.MatchingLabels(labelSelector.MatchLabels),
	}
	err := k8sClient.List(ctx, &list, opts...)

	var selectedList v1beta1.GrafanaList

	for _, instance := range list.Items {
		selected := labelsSatisfyMatchExpressions(instance.Labels, labelSelector.MatchExpressions)
		if selected {
			selectedList.Items = append(selectedList.Items, instance)
		}
	}

	return selectedList, err
}

func labelsSatisfyMatchExpressions(labels map[string]string, matchExpressions []metav1.LabelSelectorRequirement) bool {
	// To preserve support for scenario with instanceSelector: {}
	if len(labels) == 0 {
		return true
	}

	if len(matchExpressions) == 0 {
		return true
	}

	for _, matchExpression := range matchExpressions {
		selected := false

		if label, ok := labels[matchExpression.Key]; ok {
			switch matchExpression.Operator {
			case metav1.LabelSelectorOpDoesNotExist:
				selected = false
			case metav1.LabelSelectorOpExists:
				selected = true
			case metav1.LabelSelectorOpIn:
				selected = slices.Contains(matchExpression.Values, label)
			case metav1.LabelSelectorOpNotIn:
				selected = !slices.Contains(matchExpression.Values, label)
			}
		}

		// All matchExpressions must evaluate to true in order to satisfy the conditions
		if !selected {
			return false
		}
	}

	return true
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

func setNoMatchingInstance(conditions *[]metav1.Condition, generation int64, reason, message string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionNoMatchingInstance,
		Status:             "True",
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
		Reason:  reason,
		Message: message,
	})
}

func setNoMatchingFolder(conditions *[]metav1.Condition, generation int64, reason, message string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionNoMatchingFolder,
		Status:             "True",
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
		Reason:  reason,
		Message: message,
	})
}

func removeNoMatchingInstance(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionNoMatchingInstance)
}

func ignoreStatusUpdates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}
