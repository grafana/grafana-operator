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
	"github.com/grafana/grafana-operator/v5/controllers/model"
	corev1 "k8s.io/api/core/v1"
	kuberr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	// Synchronization size and timeout values
	syncBatchSize    = 100
	initialSyncDelay = 10 * time.Second
	RequeueDelay     = 10 * time.Second

	// condition types
	conditionNoMatchingInstance             = "NoMatchingInstance"
	conditionNoMatchingFolder               = "NoMatchingFolder"
	conditionInvalidSpec                    = "InvalidSpec"
	conditionNotificationPolicyLoopDetected = "NotificationPolicyLoopDetected"
	conditionSuspended                      = "Suspended"

	// condition reasons
	conditionApplySuccessful = "ApplySuccessful"
	conditionApplyFailed     = "ApplyFailed"
	conditionApplySuspended  = "ApplySuspended"

	// Finalizer
	grafanaFinalizer = "operator.grafana.com/finalizer"
)

//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

// Allow slower initial retry on any failure
// Significantly slower compared to the default exponential backoff
func defaultRateLimiter() workqueue.TypedRateLimiter[reconcile.Request] {
	return workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](RequeueDelay, 120*time.Second)
}

// Only matching instances in the scope of the resource are returned
// Resources with allowCrossNamespaceImport expands the scope to the entire cluster
// Intended to be used in reconciler functions
func GetScopedMatchingInstances(ctx context.Context, k8sClient client.Client, cr v1beta1.CommonResource) ([]v1beta1.Grafana, error) {
	log := logf.FromContext(ctx)
	instanceSelector := cr.MatchLabels()

	// Should never happen, sanity check
	if instanceSelector == nil {
		return []v1beta1.Grafana{}, nil
	}

	opts := []client.ListOption{
		// Matches all instances when MatchLabels is undefined
		client.MatchingLabels(instanceSelector.MatchLabels),
	}

	if !cr.AllowCrossNamespace() {
		// Only query resource namespace
		opts = append(opts, client.InNamespace(cr.MatchNamespace()))
	}

	var list v1beta1.GrafanaList
	err := k8sClient.List(ctx, &list, opts...)
	if err != nil {
		return []v1beta1.Grafana{}, err
	}

	if len(list.Items) == 0 {
		return []v1beta1.Grafana{}, nil
	}

	selectedList := make([]v1beta1.Grafana, 0, len(list.Items))
	var unreadyInstances []string
	for _, instance := range list.Items {
		// Matches all instances when MatchExpressions is undefined
		selected := labelsSatisfyMatchExpressions(instance.Labels, instanceSelector.MatchExpressions)
		if !selected {
			continue
		}

		// Readiness checks omits instances that have not completed a full reconciliation successfully.
		// This allows speeding up reconciliations and reduces the amount of noisy errors.
		// A toggle to skip the readiness check allows additional testing of reconcilers, like provoking the ApplyFailed synchronization condition.
		doReadinessCheck := true
		if instance.Annotations["grafana-operator/skip-readiness-check"] == "true" { //nolint:goconst
			doReadinessCheck = false
		}

		// admin url is required to interact with Grafana
		// the instance or route might not yet be ready
		if doReadinessCheck && (instance.Status.Stage != v1beta1.OperatorStageComplete || instance.Status.StageStatus != v1beta1.OperatorStageResultSuccess) {
			unreadyInstances = append(unreadyInstances, instance.Name)
			continue
		}
		selectedList = append(selectedList, instance)
	}
	if len(unreadyInstances) > 0 {
		log.Info("Grafana instances not ready, excluded from matching", "instances", unreadyInstances)
	}
	if len(selectedList) == 0 {
		log.Info("None of the available Grafana instances matched the selector, skipping reconciliation", "AllowCrossNamespaceImport", cr.AllowCrossNamespace())
	}

	return selectedList, nil
}

// getFolderUID fetches the folderUID from an existing GrafanaFolder CR declared in the specified namespace
func getFolderUID(ctx context.Context, k8sClient client.Client, ref operatorapi.FolderReferencer) (string, error) {
	if ref.FolderUID() != "" {
		return ref.FolderUID(), nil
	}
	if ref.FolderRef() == "" {
		return "", nil
	}
	folder := &v1beta1.GrafanaFolder{}

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

	return folder.CustomUIDOrUID(), nil
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

// TODO Refactor to use scheme from k8sClient.Scheme() as it's the same anyways
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

	// Even though model.GetPluginsConfigMap already sets an owner reference, it gets overwritten
	// when we fetch the actual contents of the ConfigMap using k8sClient, so we need to set it here again
	controllerutil.SetControllerReference(grafana, pluginsConfigMap, scheme) //nolint:errcheck

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

// Correctly determine cause of no matching instance from error
func setNoMatchingInstancesCondition(conditions *[]metav1.Condition, generation int64, err error) {
	var reason, message string
	if err != nil {
		reason = "ErrFetchingInstances"
		message = fmt.Sprintf("error occurred during fetching of instances: %s", err.Error())
	} else {
		reason = "EmptyAPIReply"
		message = "None of the available Grafana instances matched the selector, skipping reconciliation"
	}
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionNoMatchingInstance,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: generation,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
	})
}

func removeNoMatchingInstance(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionNoMatchingInstance)
}

func setNoMatchingFolder(conditions *[]metav1.Condition, generation int64, reason, message string) {
	meta.SetStatusCondition(conditions, metav1.Condition{
		Type:               conditionNoMatchingFolder,
		Status:             metav1.ConditionTrue,
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
		Status:             metav1.ConditionTrue,
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

func setSuspended(conditions *[]metav1.Condition, generation int64, reason string) {
	*conditions = []metav1.Condition{{
		Type:               conditionSuspended,
		Reason:             reason,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: generation,
		LastTransitionTime: metav1.Time{
			Time: time.Now(),
		},
		Message: "Resource changes are ignored",
	}}
}

func removeSuspended(conditions *[]metav1.Condition) {
	meta.RemoveStatusCondition(conditions, conditionSuspended)
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
		condition.Status = metav1.ConditionTrue
		condition.Reason = conditionApplySuccessful
		condition.Message = fmt.Sprintf("%s was successfully applied to %d instances", resource, total)
	} else {
		condition.Status = metav1.ConditionFalse
		condition.Reason = conditionApplyFailed

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

// Add finalizer through a MergePatch
// Avoids updating the entire object and only changes the finalizers
func addFinalizer(ctx context.Context, cl client.Client, cr client.Object) error {
	// Only update when changed
	if controllerutil.AddFinalizer(cr, grafanaFinalizer) {
		return patchFinalizers(ctx, cl, cr)
	}
	return nil
}

// Remove finalizer through a MergePatch
// Avoids updating the entire object and only changes the finalizers
func removeFinalizer(ctx context.Context, cl client.Client, cr client.Object) error {
	// Only update when changed
	if controllerutil.RemoveFinalizer(cr, grafanaFinalizer) {
		return patchFinalizers(ctx, cl, cr)
	}
	return nil
}

// Helper func for add/remove, avoid using directly
func patchFinalizers(ctx context.Context, cl client.Client, cr client.Object) error {
	crFinalizers := cr.GetFinalizers()

	// Create patch using slice
	patch, err := json.Marshal(map[string]any{"metadata": map[string]any{"finalizers": crFinalizers}})
	if err != nil {
		return err
	}
	return cl.Patch(ctx, cr, client.RawPatch(types.MergePatchType, patch))
}

func addAnnotation(ctx context.Context, cl client.Client, cr client.Object, key string, value string) error {
	crAnnotations := cr.GetAnnotations()

	if crAnnotations[key] == value {
		return nil
	}

	// Add key to map and create patch
	crAnnotations[key] = value
	patch, err := json.Marshal(map[string]any{"metadata": map[string]any{"annotations": crAnnotations}})
	if err != nil {
		return err
	}

	return cl.Patch(ctx, cr, client.RawPatch(types.MergePatchType, patch))
}

func removeAnnotation(ctx context.Context, cl client.Client, cr client.Object, key string) error {
	crAnnotations := cr.GetAnnotations()
	if crAnnotations[key] == "" {
		return nil
	}

	// Escape slash '/' according to RFC6901
	// We could also escape tilde '~', but that is not a valid character in annotation keys.
	key = strings.ReplaceAll(key, "/", "~1")
	patch, err := json.Marshal([]any{map[string]any{
		"op":   "remove",
		"path": "/metadata/annotations/" + key,
	}})
	if err != nil {
		return err
	}

	// MergePatchType only removes map keys when the value is null. JSONPatchType allows removing anything under a path.
	// Differs from removeFinalizer where we overwrite an array.
	return cl.Patch(ctx, cr, client.RawPatch(types.JSONPatchType, patch))
}

func mergeReconcileErrors(sources ...map[string]string) map[string]string {
	merged := make(map[string]string)

	for _, source := range sources {
		if source == nil {
			source = make(map[string]string)
		}

		for k, v := range source {
			if merged[k] == "" {
				merged[k] = v
			} else {
				merged[k] = fmt.Sprintf("%v; %v", merged[k], v)
			}
		}
	}

	return merged
}

func UpdateStatus(ctx context.Context, cl client.Client, cr v1beta1.CommonResource) {
	log := logf.FromContext(ctx)

	cr.CommonStatus().LastResync = metav1.Time{Time: time.Now()}
	if err := cl.Status().Update(ctx, cr); err != nil {
		log.Error(err, "updating status")
	}
	if meta.IsStatusConditionTrue(cr.CommonStatus().Conditions, conditionNoMatchingInstance) {
		if err := removeFinalizer(ctx, cl, cr); err != nil {
			log.Error(err, "failed to remove finalizer")
		}
	} else {
		if err := addFinalizer(ctx, cl, cr); err != nil {
			log.Error(err, "failed to set finalizer")
		}
	}
}
