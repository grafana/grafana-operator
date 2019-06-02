package grafana

import (
	"context"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"

	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
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
		helper:  common.NewKubeHelper(),
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
	err = c.Watch(&source.Kind{Type: &i8ly.Grafana{}}, &handler.EnqueueRequestForObject{})
	return err
}

var _ reconcile.Reconciler = &ReconcileGrafana{}

// ReconcileGrafana reconciles a Grafana object
type ReconcileGrafana struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client  client.Client
	scheme  *runtime.Scheme
	helper  *common.KubeHelperImpl
	plugins *PluginsHelperImpl
	config  *common.ControllerConfig
}

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
func (r *ReconcileGrafana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &i8ly.Grafana{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	cr := instance.DeepCopy()

	switch cr.Status.Phase {
	case PhaseConfigFiles:
		return r.CreateConfigFiles(cr)
	case PhaseInstallGrafana:
		return r.InstallGrafana(cr)
	case PhaseDone:
		log.Info("Grafana installation complete")
		return r.UpdatePhase(cr, PhasePlugins)
	case PhasePlugins:
		// Make the label selector available to other controllers
		r.config.AddConfigItem(common.ConfigDashboardLabelSelector, cr.Spec.DashboardLabelSelectors)
		return r.ReconcileDashboardPlugins(cr)
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileGrafana) ReconcileDashboardPlugins(cr *i8ly.Grafana) (reconcile.Result, error) {
	// Waited long enough for dashboards to be ready?
	if r.plugins.CanUpdatePlugins() == false {
		return reconcile.Result{RequeueAfter: time.Second * ReconcilePauseSeconds}, nil
	}

	// Fetch all plugins of all dashboards
	var requestedPlugins i8ly.PluginList
	for _, v := range common.GetControllerConfig().Plugins {
		requestedPlugins = append(requestedPlugins, v...)
	}

	// Consolidate plugins and create the new list of plugin requirements
	// If 'updated' is false then no changes have to be applied
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

func (r *ReconcileGrafana) ReconcilePlugins(cr *i8ly.Grafana, plugins []i8ly.GrafanaPlugin) error {
	var validPlugins []i8ly.GrafanaPlugin
	var failedPlugins []i8ly.GrafanaPlugin

	for _, plugin := range plugins {
		if r.plugins.PluginExists(plugin) == false {
			log.Info(fmt.Sprintf("Invalid plugin: %s@%s", plugin.Name, plugin.Version))
			failedPlugins = append(failedPlugins, plugin)
			continue
		}

		log.Info(fmt.Sprintf("Installing plugin: %s@%s", plugin.Name, plugin.Version))
		validPlugins = append(validPlugins, plugin)
	}

	cr.Status.InstalledPlugins = validPlugins
	cr.Status.FailedPlugins = failedPlugins

	err := r.client.Update(context.TODO(), cr)
	if err != nil {
		return err
	}

	newEnv := r.plugins.BuildEnv(cr)
	err = r.helper.UpdateGrafanaDeployment(cr.Namespace, newEnv)
	return err
}

func (r *ReconcileGrafana) UpdateDashboardMessages(plugins i8ly.PluginList) error {
	for _, plugin := range plugins {
		err := r.client.Update(context.TODO(), plugin.Origin)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileGrafana) CreateConfigFiles(cr *i8ly.Grafana) (reconcile.Result, error) {
	log.Info("Phase: Create Config Files")

	ingressType := common.GrafanaIngressName
	if common.GetControllerConfig().GetConfigBool(common.ConfigOpenshift, false) == true {
		ingressType = common.GrafanaRouteName
	}

	err := r.CreateServiceAccount(cr, common.GrafanaServiceAccountName)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, resourceName := range []string{common.GrafanaConfigMapName, common.GrafanaDashboardsConfigMapName, common.GrafanaProvidersConfigMapName, common.GrafanaDatasourcesConfigMapName, common.GrafanaServiceName, ingressType} {
		if err := r.CreateResource(cr, resourceName); err != nil {
			log.Info(fmt.Sprintf("Error in CreateConfigFiles, resourceName=%s : err=%s", resourceName, err))
			// Requeue so it can be attempted again
			return reconcile.Result{}, err
		}
	}

	log.Info("Config files created")
	return r.UpdatePhase(cr, PhaseInstallGrafana)
}

func (r *ReconcileGrafana) InstallGrafana(cr *i8ly.Grafana) (reconcile.Result, error) {
	log.Info("Phase: Install Grafana")

	err := r.CreateDeployment(cr, common.GrafanaDeploymentName)
	if err != nil {
		return reconcile.Result{}, err
	}

	return r.UpdatePhase(cr, PhaseDone)
}

// Creates the deployment from the template and injects any specified extra containers before
// submitting it
func (r *ReconcileGrafana) CreateDeployment(cr *i8ly.Grafana, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)
	if err != nil {
		return err
	}

	rawResource := newUnstructuredResourceMap(resource.(*unstructured.Unstructured))

	// Extra secrets to be added as volumes?
	if len(cr.Spec.Secrets) > 0 {
		volumes := rawResource.access("spec").access("template").access("spec").get("volumes").([]interface{})

		for _, secret := range cr.Spec.Secrets {
			volumeName := fmt.Sprintf("secret-%s", secret)
			log.Info(fmt.Sprintf("adding volume for secret '%s' as '%s'", secret, volumeName))
			volumes = append(volumes, core.Volume{
				Name: volumeName,
				VolumeSource: core.VolumeSource{
					Secret: &core.SecretVolumeSource{
						SecretName: secret,
					},
				},
			})
		}

		rawResource.access("spec").access("template").access("spec").set("volumes", volumes)
	}

	// Extra containers to add to the deployment?
	if len(cr.Spec.Containers) > 0 {
		// Otherwise append extra containers before submitting the resource
		containers := rawResource.access("spec").access("template").access("spec").get("containers").([]interface{})

		for _, container := range cr.Spec.Containers {
			containers = append(containers, container)
			log.Info(fmt.Sprintf("adding extra container '%v' to '%v'", container.Name, common.GrafanaDeploymentName))
		}

		rawResource.access("spec").access("template").access("spec").set("containers", containers)
	}

	return r.DeployResource(cr, resource, resourceName)

}

func (r *ReconcileGrafana) CreateServiceAccount(cr *i8ly.Grafana, resourceName string) error {
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
	annotations[OpenShiftOAuthRedirect] = fmt.Sprintf(`{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"%s"}}`, common.GrafanaRouteName)

	rawResource := newUnstructuredResourceMap(resource.(*unstructured.Unstructured))
	rawResource.access("metadata").set("annotations", annotations)

	return r.DeployResource(cr, resource, resourceName)
}

// Creates a generic kubernetes resource from a template
func (r *ReconcileGrafana) CreateResource(cr *i8ly.Grafana, resourceName string) error {
	resourceHelper := newResourceHelper(cr)
	resource, err := resourceHelper.createResource(resourceName)

	if err != nil {
		return err
	}

	return r.DeployResource(cr, resource, resourceName)
}

// Deploys a resource given by a runtime object
func (r *ReconcileGrafana) DeployResource(cr *i8ly.Grafana, resource runtime.Object, resourceName string) error {
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

func (r *ReconcileGrafana) UpdatePhase(cr *i8ly.Grafana, phase int) (reconcile.Result, error) {
	cr.Status.Phase = phase
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}
