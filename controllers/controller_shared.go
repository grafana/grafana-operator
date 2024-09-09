package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	operatorapi "github.com/grafana/grafana-operator/v5/api"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	grafanav1beta1 "github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const grafanaFinalizer = "operator.grafana.com/finalizer"

const (
	conditionNoMatchingInstance = "NoMatchingInstance"
	conditionNoMatchingFolder   = "NoMatchingFolder"
	conditionInvalidSpec        = "InvalidSpec"
)

const annotationAppliedNotificationPolicy = "operator.grafana.com/applied-notificationpolicy"

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

func GetMatchingInstances(ctx context.Context, k8sClient client.Client, labelSelector *metav1.LabelSelector) (v1beta1.GrafanaList, error) {
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

// getFolderUID fetches the folderUID from an existing GrafanaFolder CR declared in the specified namespace
func getFolderUID(ctx context.Context, k8sClient client.Client, ref operatorapi.FolderReferencer) (string, error) {
	if ref.FolderUID() != "" {
		return ref.FolderUID(), nil
	}
	if ref.FolderRef() == "" {
		return "", nil
	}
	folder := &grafanav1beta1.GrafanaFolder{}

	err := k8sClient.Get(ctx, client.ObjectKey{
		Namespace: ref.FolderNamespace(),
		Name:      ref.FolderRef(),
	}, folder)
	if err != nil {
		if kuberr.IsNotFound(err) {
			setNoMatchingFolder(ref.Conditions(), ref.CurrentGeneration(), "NotFound", fmt.Sprintf("Folder with name %s not found in namespace %s", ref.FolderRef(), ref.FolderNamespace()))
			return "", err
		}
		setNoMatchingFolder(ref.Conditions(), ref.CurrentGeneration(), "ErrFetchingFolder", fmt.Sprintf("Failed to fetch folder: %s", err.Error()))
		return "", err
	}
	removeNoMatchingFolder(ref.Conditions())

	return string(folder.ObjectMeta.UID), nil
}

func labelsSatisfyMatchExpressions(labels map[string]string, matchExpressions []metav1.LabelSelectorRequirement) bool {
	// To preserve support for scenario with instanceSelector: {}
	if len(labels) == 0 {
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

func removeNoMatchingInstance(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionNoMatchingInstance)
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

func removeNoMatchingFolder(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionNoMatchingFolder)
}

func setInvalidSpec(conditions *[]metav1.Condition, generation int64, reason, message string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionInvalidSpec,
		Status:             "True",
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
		Reason:  reason,
		Message: message,
	})
}

func removeInvalidSpec(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionInvalidSpec)
}

func ignoreStatusUpdates() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
	}
}

func buildSynchronizedCondition(resource string, syncType string, generation int64, applyErrors map[string]string, total int) metav1.Condition {
	condition := metav1.Condition{
		Type:               syncType,
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	}

	if len(applyErrors) == 0 {
		condition.Status = "True"
		condition.Reason = "ApplySuccessful"
		condition.Message = fmt.Sprintf("%s was successfully applied to %d instances", resource, total)
	} else {
		condition.Status = "False"
		condition.Reason = "ApplyFailed"

		var sb strings.Builder
		for i, err := range applyErrors {
			sb.WriteString(fmt.Sprintf("\n- %s: %s", i, err))
		}

		condition.Message = fmt.Sprintf("%s failed to be applied for %d out of %d instances. Errors:%s", resource, len(applyErrors), total, sb.String())
	}
	return condition
}

func getReferencedValue(ctx context.Context, cl client.Client, cr metav1.ObjectMetaAccessor, source v1beta1.ValueFromSource) (string, string, error) {
	objMeta := cr.GetObjectMeta()
	if source.SecretKeyRef != nil {
		s := &corev1.Secret{}
		err := cl.Get(ctx, client.ObjectKey{Namespace: objMeta.GetNamespace(), Name: source.SecretKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.SecretKeyRef.Key]; ok {
			return string(val), source.SecretKeyRef.Key, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in secret %s", source.SecretKeyRef.Key, source.SecretKeyRef.Name)
		}
	} else {
		s := &corev1.ConfigMap{}
		err := cl.Get(ctx, client.ObjectKey{Namespace: objMeta.GetNamespace(), Name: source.ConfigMapKeyRef.Name}, s)
		if err != nil {
			return "", "", err
		}
		if val, ok := s.Data[source.ConfigMapKeyRef.Key]; ok {
			return val, source.ConfigMapKeyRef.Key, nil
		} else {
			return "", "", fmt.Errorf("missing key %s in configmap %s", source.ConfigMapKeyRef.Key, source.ConfigMapKeyRef.Name)
		}
	}
}
