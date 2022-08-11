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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	client2 "github.com/grafana-operator/grafana-operator-experimental/controllers/client"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	grapi "github.com/grafana/grafana-api-golang-client"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

const (
	statusCodeDashboardExistsInFolder = "412"
	dashboardFinalizer                = "integreatly.org/finalizer"
)

// GrafanaDashboardReconciler reconciles a GrafanaDashboard object
type GrafanaDashboardReconciler struct {
	Client    client.Client
	Log       logr.Logger
	Scheme    *runtime.Scheme
	Discovery discovery.DiscoveryInterface
}

//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=grafana.integreatly.org,resources=grafanadashboards/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
func (r *GrafanaDashboardReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	controllerLog := log.FromContext(ctx)
	r.Log = controllerLog

	dashboard := &v1beta1.GrafanaDashboard{}
	err := r.Client.Get(ctx, client.ObjectKey{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, dashboard)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.onDashboardDeleted(ctx, req.Namespace, req.Name)
			if err != nil {
				return ctrl.Result{RequeueAfter: RequeueDelayError}, err
			}
			return ctrl.Result{}, nil
		}
		controllerLog.Error(err, "error getting grafana dashboard cr")
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	// skip dashboards without an instance selector
	if dashboard.Spec.InstanceSelector == nil {
		controllerLog.Info("no instance selector found for dashboard, nothing to do", "name", &dashboard.Name, "namespace", &dashboard.Namespace)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, nil
	}

	instances, err := r.getMatchingInstances(ctx, dashboard.Spec.InstanceSelector)
	if err != nil {
		controllerLog.Error(err, "could not find matching instance", "name", &dashboard.Name)
		return ctrl.Result{RequeueAfter: RequeueDelayError}, err
	}

	if len(instances.Items) == 0 {
		controllerLog.Info("no matching instances found for dashboard", "dashboard", &dashboard.Name, "namespace", &dashboard.Namespace)
	}

	controllerLog.Info("found matching Grafana instances", "count", len(instances.Items))

	for _, grafana := range instances.Items {
		// an admin url is required to interact with grafana
		// the instance or route might not yet be ready
		if grafana.Status.AdminUrl == "" || grafana.Status.Stage != v1beta1.OperatorStageComplete || grafana.Status.StageStatus != v1beta1.OperatorStageResultSuccess {
			controllerLog.Info("grafana instance not ready", "grafana", grafana.Name)
			continue
		}

		// first reconcile the plugins
		// append the requested dashboards to a configmap from where the
		// grafana reconciler will pick them upi
		err = r.reconcilePlugins(ctx, &grafana, dashboard)
		if err != nil {
			controllerLog.Error(err, "error reconciling plugins", "dashboard", dashboard.Name, "grafana", grafana.Name)
		}

		// then import the dashboard into the matching grafana instances
		err = r.onDashboardCreated(ctx, &grafana, dashboard)
		if err != nil {
			controllerLog.Error(err, "error reconciling dashboard", "dashboard", dashboard.Name, "grafana", grafana.Name)
		}
	}

	return ctrl.Result{RequeueAfter: RequeueDelaySuccess}, nil
}

func (r *GrafanaDashboardReconciler) onDashboardDeleted(ctx context.Context, namespace string, name string) error {
	list := v1beta1.GrafanaList{}
	opts := []client.ListOption{}
	err := r.Client.List(ctx, &list, opts...)
	if err != nil {
		return err
	}

	for _, grafana := range list.Items {
		if found, uid := grafana.FindDashboardByNamespaceAndName(namespace, name); found {
			grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, &grafana)
			if err != nil {
				return err
			}

			err = grafanaClient.DeleteDashboardByUID(uid)
			if err != nil {
				if !strings.Contains(err.Error(), "status: 404") {
					return err
				}
			}

			err = grafana.RemoveDashboard(namespace, name)
			if err != nil {
				return err
			}

			err = r.Client.Update(ctx, &grafana)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *GrafanaDashboardReconciler) onDashboardCreated(ctx context.Context, grafana *v1beta1.Grafana, cr *v1beta1.GrafanaDashboard) error {
	if cr.Spec.Json == "" {
		return nil
	}

	var dashboardFromJson map[string]interface{}
	err := json.Unmarshal([]byte(cr.Spec.Json), &dashboardFromJson)
	if err != nil {
		return err
	}

	grafanaClient, err := client2.NewGrafanaClient(ctx, r.Client, grafana)
	if err != nil {
		return err
	}

	// check if dashboard already exists
	instanceDashboards, err := grafanaClient.Dashboards() // dashboards in Grafana instance
	for _, instanceDashboard := range instanceDashboards {
		if instanceDashboard.UID == string(cr.UID) && instanceDashboard.Title == dashboardFromJson["title"] {
			return nil
		}
	}

	dashboardFromJson["uid"] = string(cr.UID)

	resp, err := grafanaClient.NewDashboard(grapi.Dashboard{
		Meta: grapi.DashboardMeta{
			IsStarred: false,
			Slug:      cr.Name,
			//Folder:    ,
			//URL:       "",
		},
		Model: dashboardFromJson,
		//Folder:    0,
		Overwrite: true,
		Message:   "",
	})

	if err != nil {
		return err
	}

	if resp.Status != "success" {
		return errors.NewBadRequest(fmt.Sprintf("error creating dashboard, status was %v", resp.Status))
	}

	grafana.AddDashboard(cr.Namespace, cr.Name, resp.UID)
	if err != nil {
		return err
	}
	return r.Client.Update(ctx, grafana)
}

func (r *GrafanaDashboardReconciler) reconcilePlugins(ctx context.Context, grafana *v1beta1.Grafana, dashboard *v1beta1.GrafanaDashboard) error {
	if dashboard.Spec.Plugins == nil || len(dashboard.Spec.Plugins) == 0 {
		return nil
	}

	pluginsConfigMap := model.GetPluginsConfigMap(grafana, r.Scheme)
	selector := client.ObjectKey{
		Namespace: pluginsConfigMap.Namespace,
		Name:      pluginsConfigMap.Name,
	}

	err := r.Client.Get(ctx, selector, pluginsConfigMap)
	if err != nil {
		return err
	}

	val, err := json.Marshal(dashboard.Spec.Plugins.Sanitize())
	if err != nil {
		return err
	}

	if pluginsConfigMap.BinaryData == nil {
		pluginsConfigMap.BinaryData = make(map[string][]byte)
	}

	if bytes.Compare(val, pluginsConfigMap.BinaryData[dashboard.Name]) != 0 {
		pluginsConfigMap.BinaryData[dashboard.Name] = val
		return r.Client.Update(ctx, pluginsConfigMap)
	}

	return nil
}

func (r *GrafanaDashboardReconciler) getMatchingInstances(ctx context.Context, labelSelector *v1.LabelSelector) (v1beta1.GrafanaList, error) {
	var list v1beta1.GrafanaList
	opts := []client.ListOption{
		client.MatchingLabels(labelSelector.MatchLabels),
	}

	err := r.Client.List(ctx, &list, opts...)
	return list, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *GrafanaDashboardReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.GrafanaDashboard{}).
		Complete(r)
}
