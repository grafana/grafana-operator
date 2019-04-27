package grafanadashboard

import (
	"context"
	defaultErrors "errors"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/grafana"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"time"

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

var log = logf.Log.WithName("controller_grafanadashboard")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDashboard Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGrafanaDashboard{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: common.GetControllerConfig(),
		helper: grafana.NewKubeHelper(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafanadashboard-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDashboard
	err = c.Watch(&source.Kind{Type: &v1alpha1.GrafanaDashboard{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Info("Starting dashboard controller")
	}

	return err
}

var _ reconcile.Reconciler = &ReconcileGrafanaDashboard{}

// ReconcileGrafanaDashboard reconciles a GrafanaDashboard object
type ReconcileGrafanaDashboard struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	config *common.ControllerConfig
	helper *grafana.KubeHelperImpl
}

func (r *ReconcileGrafanaDashboard) matchesSelector(d *v1alpha1.GrafanaDashboard, s *v1.LabelSelector) (bool, error) {
	selector, err := v1.LabelSelectorAsSelector(s)
	if err != nil {
		return false, err
	}

	return selector.Empty() || selector.Matches(labels.Set(d.Labels)), nil
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDashboard) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	dashboardLabelSelector := r.config.GetConfigItem(common.ConfigDashboardLabelSelector, nil)
	if dashboardLabelSelector == nil {
		return reconcile.Result{RequeueAfter: time.Second * 10}, nil
	}

	// Fetch the GrafanaDashboard instance
	instance := &v1alpha1.GrafanaDashboard{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	instanceCopy := instance.DeepCopy()
	if match, err := r.matchesSelector(instanceCopy, dashboardLabelSelector.(*v1.LabelSelector)); err != nil {
		return reconcile.Result{}, err
	} else if !match {
		log.Info(fmt.Sprintf("Ignoring dashboard '%s' in '%s' because the labels do not match", instanceCopy.Name, instanceCopy.Namespace))
	}

	if instanceCopy.DeletionTimestamp != nil {
		err = r.DeleteDashboard(instanceCopy)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.config.EmptyPluginsFor(instanceCopy)
	} else {
		err = r.ImportDashboard(instanceCopy)
		if err != nil {
			return reconcile.Result{}, err
		}
		r.config.SetPluginsFor(instanceCopy)
	}

	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDashboard) ImportDashboard(d *v1alpha1.GrafanaDashboard) error {
	operatorNamespace := r.config.GetConfigString(common.ConfigOperatorNamespace, "")
	if operatorNamespace == "" {
		return defaultErrors.New("no monitoring namespace set")
	}

	return r.helper.UpdateDashboard(operatorNamespace, d.Namespace, d)
}

func (r *ReconcileGrafanaDashboard) DeleteDashboard(d *v1alpha1.GrafanaDashboard) error {
	operatorNamespace := r.config.GetConfigString(common.ConfigOperatorNamespace, "")
	if operatorNamespace == "" {
		return defaultErrors.New("no monitoring namespace set")
	}

	return r.helper.DeleteDashboard(operatorNamespace, d.Namespace, d)
}
