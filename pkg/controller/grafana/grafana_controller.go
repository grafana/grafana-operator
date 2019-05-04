package grafana

import (
	"context"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_grafana")

const (
	PhaseConfigFiles int = iota
	PhaseInstallGrafana
	PhaseDone
	PhasePlugins
)

const ReconcilePauseSeconds = 5
const OpenShiftOAuthRedirect = "serviceaccounts.openshift.io/oauth-redirectreference.primary"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Grafana Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGrafana{
		client:  mgr.GetClient(),
		scheme:  mgr.GetScheme(),
		helper:  NewKubeHelper(),
		plugins: newPluginsHelper(),
		config:  common.GetControllerConfig(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafana-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Grafana
	err = c.Watch(&source.Kind{Type: &integreatly.Grafana{}}, &handler.EnqueueRequestForObject{})
	return err
}

var _ reconcile.Reconciler = &ReconcileGrafana{}

// ReconcileGrafana reconciles a Grafana object
type ReconcileGrafana struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	helper  *KubeHelperImpl
	plugins *PluginsHelperImpl
	config  *common.ControllerConfig
}

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
func (r *ReconcileGrafana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &integreatly.Grafana{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	instanceCopy := instance.DeepCopy()

	switch instanceCopy.Status.Phase {
	case PhaseConfigFiles:
		return r.CreateConfigFiles(instanceCopy)
	case PhaseInstallGrafana:
		return r.InstallGrafana(instanceCopy)
	case PhaseDone:
		r.config.AddConfigItem(common.ConfigDashboardLabelSelector, instanceCopy.Spec.DashboardLabelSelector)
		return reconcile.Result{Requeue: true}, r.UpdatePhase(instanceCopy, PhasePlugins)
	case PhasePlugins:
		return r.ReconcileDashboardPlugins(instanceCopy)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafana) ReconcileDashboardPlugins(cr *integreatly.Grafana) (reconcile.Result, error) {
	// Waited long enough for dashboards to be ready?
	if r.plugins.CanUpdatePlugins() == false {
		return reconcile.Result{RequeueAfter: time.Second * ReconcilePauseSeconds}, nil
	}

	// Fetch all plugins of all dashboards
	var requestedPlugins integreatly.PluginList
	for _, v := range common.GetControllerConfig().Plugins {
		requestedPlugins = append(requestedPlugins, v...)
	}

	filteredPlugins, updated := r.plugins.FilterPlugins(cr, requestedPlugins)
	if updated {
		r.ReconcilePlugins(cr, filteredPlugins)

		// Update the dashboards that had their plugins modified
		// to let the owners know about the status
		err := r.UpdateDashboardMessages(filteredPlugins)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{RequeueAfter: time.Second * ReconcilePauseSeconds}, nil
}

func (r *ReconcileGrafana) ReconcilePlugins(cr *integreatly.Grafana, plugins []integreatly.GrafanaPlugin) error {
	var validPlugins []integreatly.GrafanaPlugin
	for _, plugin := range plugins {
		if r.plugins.PluginExists(plugin) == false {
			continue
		}

		log.Info(fmt.Sprintf("Installing plugin: %s@%s", plugin.Name, plugin.Version))
		validPlugins = append(validPlugins, plugin)
	}

	cr.Status.InstalledPlugins = validPlugins
	err := r.client.Update(context.TODO(), cr)
	if err != nil {
		return err
	}

	newEnv := r.plugins.BuildEnv(cr)
	err = r.helper.updateGrafanaDeployment(cr.Namespace, newEnv)
	return err
}

func (r *ReconcileGrafana) UpdateDashboardMessages(plugins integreatly.PluginList) error {
	for _, plugin := range plugins {
		err := r.client.Update(context.TODO(), plugin.Origin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileGrafana) CreateConfigFiles(cr *integreatly.Grafana) (reconcile.Result, error) {
	log.Info("Phase: Create Config Files")

	ingressType := GrafanaIngressName
	if common.GetControllerConfig().GetConfigBool(common.ConfigOpenshift, false) == true {
		ingressType = GrafanaRouteName
	}

	err := r.CreateServiceAccount(cr, GrafanaServiceAccountName)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	for _, resourceName := range []string{GrafanaConfigMapName, GrafanaDashboardsConfigMapName, GrafanaProvidersConfigMapName, GrafanaDatasourcesConfigMapName, GrafanaServiceName, ingressType} {
		if err := r.CreateResource(cr, resourceName); err != nil {
			log.Info(fmt.Sprintf("Error in CreateConfigFiles, resourceName=%s : err=%s", resourceName, err))
			// Requeue so it can be attempted again
			return reconcile.Result{Requeue: true}, err
		}
	}

	log.Info("Config files created")

	err = r.UpdatePhase(cr, PhaseInstallGrafana)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{RequeueAfter: time.Second * ReconcilePauseSeconds}, nil
}

func (r *ReconcileGrafana) InstallGrafana(cr *integreatly.Grafana) (reconcile.Result, error) {
	log.Info("Phase: Install Grafana")

	err := r.CreateDeployment(cr, GrafanaDeploymentName)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	err = r.UpdatePhase(cr, PhaseDone)
	if err != nil {
		return reconcile.Result{}, err
	}

	log.Info("Grafana installation complete")
	return reconcile.Result{Requeue: true}, nil
}

// Creates the deployment from the template and injects any specified extra containers before
// submitting it
func (r *ReconcileGrafana) CreateDeployment(cr *integreatly.Grafana, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)

	if err != nil {
		return err
	}

	// Deploy the unmodified resource if no extra containers are specified
	if len(cr.Spec.Containers) == 0 {
		return r.DeployResource(cr, resource, resourceName)
	}

	// Otherwise append extra containers before submitting the resource
	rawResource := newUnstructuredResourceMap(resource.(*unstructured.Unstructured))
	containers := rawResource.access("spec").access("template").access("spec").get("containers").([]interface{})

	for _, container := range cr.Spec.Containers {
		containers = append(containers, container)
		log.Info(fmt.Sprintf("adding extra container '%v' to '%v'", container.Name, GrafanaDeploymentName))
	}

	rawResource.access("spec").access("template").access("spec").set("containers", containers)
	return r.DeployResource(cr, resource, resourceName)
}

func (r *ReconcileGrafana) CreateServiceAccount(cr *integreatly.Grafana, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)

	if err != nil {
		return err
	}

	// Deploy the unmodified resource if not on OpenShift
	if common.GetControllerConfig().GetConfigBool(common.ConfigOpenshift, false) == false {
		return r.DeployResource(cr, resource, resourceName)
	}

	// Otherwise add an annotation that allows using the OAuthProxy (and will have no
	// effect otherwise)
	annotations := make(map[string]string)
	annotations[OpenShiftOAuthRedirect] = fmt.Sprintf(`{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"%s"}}`, GrafanaRouteName)

	rawResource := newUnstructuredResourceMap(resource.(*unstructured.Unstructured))
	rawResource.access("metadata").set("annotations", annotations)

	return r.DeployResource(cr, resource, resourceName)
}

// Creates a generic kubernetes resource from a template
func (r *ReconcileGrafana) CreateResource(cr *integreatly.Grafana, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)

	if err != nil {
		return err
	}

	return r.DeployResource(cr, resource, resourceName)
}

// Deploys a resource given by a runtime object
func (r *ReconcileGrafana) DeployResource(cr *integreatly.Grafana, resource runtime.Object, resourceName string) error {
	// Try to find the resource, it may already exist
	selector := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      resourceName,
	}
	err := r.client.Get(context.TODO(), selector, resource)

	// The resource exists, do nothing
	if err == nil {
		return nil
	}

	// Resource does not exist or something went wrong
	if errors.IsNotFound(err) {
		log.Info(fmt.Sprintf("Creating %s", resourceName))
	} else {
		return err
	}

	// Set the CR as the owner of this resource so that when
	// the CR is deleted this resource also gets removed
	err = controllerutil.SetControllerReference(cr, resource.(v1.Object), r.scheme)
	if err != nil {
		return err
	}

	err = r.client.Create(context.TODO(), resource)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileGrafana) UpdatePhase(cr *integreatly.Grafana, phase int) error {
	key := types.NamespacedName{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}

	// Refresh the resource before updating the status
	err := r.client.Get(context.TODO(), key, cr)
	if err != nil {
		return err
	}

	if cr.Status.Phase == phase {
		return nil
	}

	cr.Status.Phase = phase
	return r.client.Update(context.TODO(), cr)
}
