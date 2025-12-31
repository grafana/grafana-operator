/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/grafana/grafana-openapi-client-go/client/datasources"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/spyzhov/ajson"

	genapi "github.com/grafana/grafana-openapi-client-go/client"
	grafanaclient "github.com/grafana/grafana-operator/v5/controllers/client"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
)

const (
	conditionDatasourceSynchronized = "DatasourceSynchronized"
	conditionReasonInvalidModel     = "InvalidModel"
)

// GrafanaDatasourceReconciler reconciles a GrafanaDatasource object
type GrafanaDatasourceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *Config
}

func (r *GrafanaDatasourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx).WithName("GrafanaDatasourceReconciler")
	ctx = logf.IntoContext(ctx, log)

	cr := &v1beta1.GrafanaDatasource{}

	err := r.Get(ctx, req.NamespacedName, cr)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("error getting grafana datasource cr: %w", err)
	}

	if cr.GetDeletionTimestamp() != nil {
		// Check if resource needs clean up
		if controllerutil.ContainsFinalizer(cr, grafanaFinalizer) {
			if err := r.finalize(ctx, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to finalize GrafanaDatasource: %w", err)
			}

			if err := removeFinalizer(ctx, r.Client, cr); err != nil {
				return ctrl.Result{}, fmt.Errorf("failed to remove finalizer: %w", err)
			}
		}

		return ctrl.Result{}, nil
	}

	defer UpdateStatus(ctx, r.Client, cr)

	if cr.Spec.Suspend {
		setSuspended(&cr.Status.Conditions, cr.Generation, conditionReasonApplySuspended)
		return ctrl.Result{}, nil
	}

	removeSuspended(&cr.Status.Conditions)

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)
		cr.Status.NoMatchingInstances = true

		return ctrl.Result{}, fmt.Errorf("failed fetching instances: %w", err)
	}

	if len(instances) == 0 {
		setNoMatchingInstancesCondition(&cr.Status.Conditions, cr.Generation, err)
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)
		cr.Status.NoMatchingInstances = true

		return ctrl.Result{}, ErrNoMatchingInstances
	}

	removeNoMatchingInstance(&cr.Status.Conditions)
	cr.Status.NoMatchingInstances = false

	log.Info("found matching Grafana instances for datasource", "count", len(instances))

	uid := cr.GetGrafanaUID()
	log = log.WithValues("uid", uid)
	ctx = logf.IntoContext(ctx, log)

	if cr.IsUpdatedUID() {
		log.Info("datasource uid got updated, deleting datasources with the old uid")

		if err := r.deleteOldDatasource(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}

		// Clean up uid, so further reconcilications can track changes there
		cr.Status.UID = ""

		// Force requeue for datasource creation
		return ctrl.Result{Requeue: true}, nil
	}

	datasource, hash, err := r.buildDatasourceModel(ctx, cr)
	if err != nil {
		setInvalidSpec(&cr.Status.Conditions, cr.Generation, conditionReasonInvalidModel, err.Error())
		meta.RemoveStatusCondition(&cr.Status.Conditions, conditionDatasourceSynchronized)

		return ctrl.Result{}, fmt.Errorf("building datasource model: %w", err)
	}

	removeInvalidSpec(&cr.Status.Conditions)

	pluginErrors := make(map[string]string)
	applyErrors := make(map[string]string)

	for _, grafana := range instances {
		if grafana.IsInternal() {
			// first reconcile the plugins
			// append the requested datasources to a configmap from where the
			// grafana reconciler will pick them up
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, cr.Spec.Plugins, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
			if err != nil {
				pluginErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
			}
		}

		// then import the datasource into the matching grafana instances
		err = r.onDatasourceCreated(ctx, &grafana, cr, datasource, hash)
		if err != nil {
			applyErrors[fmt.Sprintf("%s/%s", grafana.Namespace, grafana.Name)] = err.Error()
		}
	}

	if len(pluginErrors) > 0 {
		err := fmt.Errorf("%v", pluginErrors)
		log.Error(err, "failed to apply plugins to all instances")
	}

	allApplyErrors := mergeReconcileErrors(applyErrors, pluginErrors)

	condition := buildSynchronizedCondition("Datasource", conditionDatasourceSynchronized, cr.Generation, allApplyErrors, len(instances))
	meta.SetStatusCondition(&cr.Status.Conditions, condition)

	if len(allApplyErrors) > 0 {
		return ctrl.Result{}, fmt.Errorf("failed to apply to all instances: %v", allApplyErrors)
	}

	cr.Status.Hash = hash
	cr.Status.LastMessage = "" //nolint:staticcheck
	cr.Status.UID = cr.GetGrafanaUID()

	return ctrl.Result{RequeueAfter: r.Cfg.requeueAfter(cr.Spec.ResyncPeriod)}, nil
}

func (r *GrafanaDatasourceReconciler) deleteOldDatasource(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	for _, grafana := range instances {
		found, uid := grafana.Status.Datasources.Find(cr.Namespace, cr.Name)
		if !found {
			continue
		}

		gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		_, err = gClient.Datasources.DeleteDataSourceByUID(*uid) //nolint

		var notFound *datasources.GetDataSourceByUIDNotFound
		if err != nil {
			if !errors.As(err, &notFound) {
				return fmt.Errorf("deleting datasource to update uid %s: %w", *uid, err)
			}
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) finalize(ctx context.Context, cr *v1beta1.GrafanaDatasource) error {
	log := logf.FromContext(ctx)
	log.Info("Finalizing GrafanaDatasource")

	instances, err := GetScopedMatchingInstances(ctx, r.Client, cr)
	if err != nil {
		return fmt.Errorf("fetching instances: %w", err)
	}

	uid := cr.GetGrafanaUID()

	for _, grafana := range instances {
		gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, &grafana)
		if err != nil {
			return err
		}

		_, err = gClient.Datasources.DeleteDataSourceByUID(uid) //nolint:errcheck
		if err != nil {
			var notFound *datasources.DeleteDataSourceByUIDNotFound
			if !errors.As(err, &notFound) {
				return fmt.Errorf("deleting datasource %s: %w", uid, err)
			}
		}

		if grafana.IsInternal() {
			err = ReconcilePlugins(ctx, r.Client, r.Scheme, &grafana, nil, cr.GetPluginConfigMapKey(), cr.GetPluginConfigMapDeprecatedKey())
			if err != nil {
				return fmt.Errorf("reconciling plugins: %w", err)
			}
		}

		// Update grafana instance Status
		err = grafana.RemoveNamespacedResource(ctx, r.Client, cr)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) onDatasourceCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDatasource, datasource *models.UpdateDataSourceCommand, hash string) error {
	if grafana.IsExternal() && cr.Spec.Plugins != nil {
		return fmt.Errorf("external grafana instances don't support plugins, please remove spec.plugins from your datasource cr")
	}

	if cr.Spec.Datasource == nil {
		return nil
	}

	gClient, err := grafanaclient.NewGeneratedGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	exists, uid, err := r.Exists(gClient, datasource.UID, datasource.Name)
	if err != nil {
		return err
	}

	if exists && cr.Unchanged(hash) {
		if err := r.syncCorrelations(ctx, gClient, cr); err != nil {
			return fmt.Errorf("syncing correlations: %w", err)
		}

		return nil
	}

	encoded, err := json.Marshal(datasource)
	if err != nil {
		return fmt.Errorf("representing datasource as JSON: %w", err)
	}

	if exists {
		var body models.UpdateDataSourceCommand
		if err := json.Unmarshal(encoded, &body); err != nil {
			return fmt.Errorf("representing data source as update command: %w", err)
		}

		datasource.UID = uid
		_, err := gClient.Datasources.UpdateDataSourceByUID(datasource.UID, &body) //nolint
		if err != nil {
			return err
		}
	} else {
		var body models.AddDataSourceCommand
		if err := json.Unmarshal(encoded, &body); err != nil {
			return fmt.Errorf("representing data source as create command: %w", err)
		}
		_, err = gClient.Datasources.AddDataSource(&body) //nolint
		if err != nil {
			return err
		}
	}

	if err := r.syncCorrelations(ctx, gClient, cr); err != nil {
		return fmt.Errorf("syncing correlations: %w", err)
	}

	// Update grafana instance Status
	return grafana.AddNamespacedResource(ctx, r.Client, cr, cr.NamespacedResource())
}

func (r *GrafanaDatasourceReconciler) Exists(gClient *genapi.GrafanaHTTPAPI, uid, name string) (bool, string, error) {
	items, err := gClient.Datasources.GetDataSources()
	if err != nil {
		return false, "", fmt.Errorf("fetching data sources: %w", err)
	}

	for _, datasource := range items.Payload {
		if datasource.UID == uid || datasource.Name == name {
			return true, datasource.UID, nil
		}
	}

	return false, "", nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDatasourceReconciler) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	const (
		secretIndexKey    string = ".metadata.secret"
		configMapIndexKey string = ".metadata.configMap"
	)

	// Index the datasources by the Secret references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &v1beta1.GrafanaDatasource{}, secretIndexKey,
		r.indexSecretSource()); err != nil {
		return fmt.Errorf("failed setting secret index fields: %w", err)
	}

	// Index the datasources by the ConfigMap references they (may) point at.
	if err := mgr.GetCache().IndexField(ctx, &v1beta1.GrafanaDatasource{}, configMapIndexKey,
		r.indexConfigMapSource()); err != nil {
		return fmt.Errorf("failed setting configmap index fields: %w", err)
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDatasource{}, builder.WithPredicates(
			ignoreStatusUpdates(),
		)).
		Watches(
			&corev1.Secret{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(secretIndexKey)),
		).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.requestsForChangeByField(configMapIndexKey)),
		).
		Complete(r)
}

func (r *GrafanaDatasourceReconciler) indexSecretSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		datasource, ok := o.(*v1beta1.GrafanaDatasource)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaDatasource, got %T", o))
		}

		var secretRefs []string

		for _, valueFrom := range datasource.Spec.ValuesFrom {
			if valueFrom.ValueFrom.SecretKeyRef != nil {
				secretRefs = append(secretRefs, fmt.Sprintf("%s/%s", datasource.Namespace, valueFrom.ValueFrom.SecretKeyRef.Name))
			}
		}

		return secretRefs
	}
}

func (r *GrafanaDatasourceReconciler) indexConfigMapSource() func(o client.Object) []string {
	return func(o client.Object) []string {
		datasource, ok := o.(*v1beta1.GrafanaDatasource)
		if !ok {
			panic(fmt.Sprintf("Expected a GrafanaDatasource, got %T", o))
		}

		var configMapRefs []string

		for _, valueFrom := range datasource.Spec.ValuesFrom {
			if valueFrom.ValueFrom.ConfigMapKeyRef != nil {
				configMapRefs = append(configMapRefs, fmt.Sprintf("%s/%s", datasource.Namespace, valueFrom.ValueFrom.ConfigMapKeyRef.Name))
			}
		}

		return configMapRefs
	}
}

func (r *GrafanaDatasourceReconciler) requestsForChangeByField(indexKey string) handler.MapFunc {
	return func(ctx context.Context, o client.Object) []reconcile.Request {
		var list v1beta1.GrafanaDatasourceList
		if err := r.List(ctx, &list, client.MatchingFields{
			indexKey: fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
		}); err != nil {
			logf.FromContext(ctx).Error(err, "failed to list datasources for watch mapping")
			return nil
		}

		var reqs []reconcile.Request
		for _, datasource := range list.Items {
			reqs = append(reqs, reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: datasource.Namespace,
				Name:      datasource.Name,
			}})
		}

		return reqs
	}
}

func (r *GrafanaDatasourceReconciler) buildDatasourceModel(ctx context.Context, cr *v1beta1.GrafanaDatasource) (*models.UpdateDataSourceCommand, string, error) {
	log := logf.FromContext(ctx)
	// Overwrite OrgID to ensure the field is useless
	cr.Spec.Datasource.OrgID = nil

	initialBytes, err := json.Marshal(cr.Spec.Datasource)
	if err != nil {
		return nil, "", fmt.Errorf("encoding existing datasource model as json: %w", err)
	}

	jsonRoot, err := ajson.Unmarshal(initialBytes)
	if err != nil {
		return nil, "", fmt.Errorf("parsing marshaled json as abstract json: %w", err)
	}

	if err := jsonRoot.AppendObject("uid", ajson.StringNode("", cr.GetGrafanaUID())); err != nil {
		return nil, "", fmt.Errorf("overriding uid: %w", err)
	}

	for _, override := range cr.Spec.ValuesFrom {
		val, key, err := getReferencedValue(ctx, r.Client, cr.Namespace, override.ValueFrom)
		if err != nil {
			return nil, "", fmt.Errorf("getting referenced value: %w", err)
		}

		nodes, err := jsonRoot.JSONPath("$." + override.TargetPath)
		if err != nil {
			return nil, "", fmt.Errorf("getting nodes to override: %w", err)
		}

		for _, n := range nodes {
			currentValue, err := n.GetString()
			if err != nil {
				return nil, "", fmt.Errorf("cannot replace non string field: %w", err)
			}

			substitution := strings.ReplaceAll(currentValue, fmt.Sprintf("${%v}", key), val)
			substitution = strings.ReplaceAll(substitution, fmt.Sprintf("$%v", key), val)
			log.V(1).Info("overriding value", "key", override.TargetPath, "value", val)

			if err := n.SetString(substitution); err != nil {
				return nil, "", fmt.Errorf("setting new value for field: %w", err)
			}
		}
	}

	newBytes, err := ajson.Marshal(jsonRoot)
	if err != nil {
		return nil, "", fmt.Errorf("encoding expanded datasource model as json: %w", err)
	}

	// TODO models.DataSource has SecureJsonData field now, verify if below is still true
	// We use UpdateDataSourceCommand here because models.DataSource lacks the SecureJsonData field
	var res models.UpdateDataSourceCommand
	if err = json.Unmarshal(newBytes, &res); err != nil {
		return nil, "", fmt.Errorf("deserializing expanded datasource model from json: %w", err)
	}

	// TODO Remove hashing along with the Status.Hash field
	hash := sha256.New()
	hash.Write(newBytes)

	return &res, fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func correlationKey(targetUID, label string) string {
	return fmt.Sprintf("%s:%s", targetUID, label)
}

func (r *GrafanaDatasourceReconciler) syncCorrelations(ctx context.Context, gClient *genapi.GrafanaHTTPAPI, cr *v1beta1.GrafanaDatasource) error {
	log := logf.FromContext(ctx)
	sourceUID := cr.GetGrafanaUID()

	if len(cr.Spec.Correlations) == 0 {
		return r.deleteAllCorrelations(ctx, gClient, sourceUID)
	}

	existingCorrelations, err := gClient.Datasources.GetCorrelationsBySourceUID(sourceUID)
	if err != nil {
		var notFound *datasources.GetCorrelationsBySourceUIDNotFound
		if !errors.As(err, &notFound) {
			return fmt.Errorf("fetching existing correlations: %w", err)
		}
	}

	existingByKey := make(map[string]*models.Correlation)
	if existingCorrelations != nil {
		for _, c := range existingCorrelations.Payload {
			key := correlationKey(c.TargetUID, c.Label)
			existingByKey[key] = c
		}
	}

	desiredKeys := make(map[string]struct{})
	for _, c := range cr.Spec.Correlations {
		key := correlationKey(c.TargetUID, c.Label)
		desiredKeys[key] = struct{}{}
	}

	for key, existing := range existingByKey {
		if _, found := desiredKeys[key]; !found {
			log.Info("Deleting correlation", "uid", existing.UID, "targetUID", existing.TargetUID, "label", existing.Label)
			_, err := gClient.Datasources.DeleteCorrelation(sourceUID, existing.UID) //nolint
			if err != nil {
				var notFound *datasources.DeleteCorrelationNotFound
				if !errors.As(err, &notFound) {
					return fmt.Errorf("deleting correlation %s: %w", existing.UID, err)
				}
			}
		}
	}

	for _, desired := range cr.Spec.Correlations {
		key := correlationKey(desired.TargetUID, desired.Label)
		if existing, found := existingByKey[key]; found {
			if err := r.updateCorrelationByUID(gClient, sourceUID, existing.UID, desired); err != nil {
				return fmt.Errorf("updating correlation (targetUID=%s, label=%s): %w", desired.TargetUID, desired.Label, err)
			}
		} else {
			if err := r.createCorrelation(gClient, sourceUID, desired); err != nil {
				return fmt.Errorf("creating correlation (targetUID=%s, label=%s): %w", desired.TargetUID, desired.Label, err)
			}
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) deleteAllCorrelations(ctx context.Context, gClient *genapi.GrafanaHTTPAPI, sourceUID string) error {
	log := logf.FromContext(ctx)

	existingCorrelations, err := gClient.Datasources.GetCorrelationsBySourceUID(sourceUID)
	if err != nil {
		var notFound *datasources.GetCorrelationsBySourceUIDNotFound
		if errors.As(err, &notFound) {
			return nil
		}

		return fmt.Errorf("fetching existing correlations: %w", err)
	}

	if existingCorrelations == nil {
		return nil
	}

	for _, c := range existingCorrelations.Payload {
		log.Info("Deleting correlation", "uid", c.UID)
		_, err := gClient.Datasources.DeleteCorrelation(sourceUID, c.UID) //nolint
		if err != nil {
			var notFound *datasources.DeleteCorrelationNotFound
			if !errors.As(err, &notFound) {
				return fmt.Errorf("deleting correlation %s: %w", c.UID, err)
			}
		}
	}

	return nil
}

func (r *GrafanaDatasourceReconciler) createCorrelation(gClient *genapi.GrafanaHTTPAPI, sourceUID string, c v1beta1.GrafanaDatasourceCorrelation) error {
	cmd := &models.CreateCorrelationCommand{
		TargetUID:   c.TargetUID,
		Label:       c.Label,
		Description: c.Description,
		Type:        models.CorrelationType(c.Type),
	}

	if c.Config != nil {
		cmd.Config = r.buildCorrelationConfig(c.Config)
	}

	_, err := gClient.Datasources.CreateCorrelation(sourceUID, cmd) //nolint

	return err
}

func (r *GrafanaDatasourceReconciler) updateCorrelationByUID(gClient *genapi.GrafanaHTTPAPI, sourceUID, correlationUID string, desired v1beta1.GrafanaDatasourceCorrelation) error {
	cmd := &models.UpdateCorrelationCommand{
		Label:       desired.Label,
		Description: desired.Description,
	}

	if desired.Config != nil {
		cmd.Config = r.buildCorrelationConfigUpdate(desired.Config)
	}

	params := datasources.NewUpdateCorrelationParams().
		WithSourceUID(sourceUID).
		WithCorrelationUID(correlationUID).
		WithBody(cmd)

	_, err := gClient.Datasources.UpdateCorrelation(params) //nolint

	return err
}

func (r *GrafanaDatasourceReconciler) buildCorrelationConfig(config *v1beta1.GrafanaDatasourceCorrelationConfig) *models.CorrelationConfig {
	field := config.Field
	result := &models.CorrelationConfig{
		Field: &field,
		Type:  models.CorrelationType(config.Type),
	}

	if config.Target != nil {
		result.Target = convertJSONToInterface(config.Target)
	}

	if len(config.Transformations) > 0 {
		result.Transformations = make(models.Transformations, len(config.Transformations))
		for i, t := range config.Transformations {
			result.Transformations[i] = &models.Transformation{
				Type:       t.Type,
				Field:      t.Field,
				Expression: t.Expression,
				MapValue:   t.MapValue,
			}
		}
	}

	return result
}

func (r *GrafanaDatasourceReconciler) buildCorrelationConfigUpdate(config *v1beta1.GrafanaDatasourceCorrelationConfig) *models.CorrelationConfigUpdateDTO {
	result := &models.CorrelationConfigUpdateDTO{
		Field: config.Field,
	}

	if config.Target != nil {
		result.Target = convertJSONToInterface(config.Target)
	}

	if len(config.Transformations) > 0 {
		result.Transformations = make([]*models.Transformation, len(config.Transformations))
		for i, t := range config.Transformations {
			result.Transformations[i] = &models.Transformation{
				Type:       t.Type,
				Field:      t.Field,
				Expression: t.Expression,
				MapValue:   t.MapValue,
			}
		}
	}

	return result
}

func convertJSONToInterface(j *apiextensionsv1.JSON) any {
	if j == nil || j.Raw == nil {
		return nil
	}

	var result any
	if err := json.Unmarshal(j.Raw, &result); err != nil {
		return nil
	}

	return result
}
