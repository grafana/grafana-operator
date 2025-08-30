package grafana

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type PluginsReconciler struct {
	client client.Client
}

func NewPluginsReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PluginsReconciler{
		client: client,
	}
}

func (r *PluginsReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("PluginsReconciler")

	vars.Plugins = ""

	plugins := model.GetPluginsConfigMap(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, plugins, func() error {
		if scheme != nil {
			err := controllerutil.SetOwnerReference(cr, plugins, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(plugins, cr.Labels)

		return nil
	})
	if err != nil {
		log.Error(err, "error getting plugins config map", "name", plugins.Name, "namespace", plugins.Namespace)
		return v1beta1.OperatorStageResultFailed, err
	}

	// plugins config map found, but may be empty
	if len(plugins.BinaryData) == 0 {
		vars.Plugins = ""
		return v1beta1.OperatorStageResultSuccess, nil
	}

	pm := v1beta1.NewPluginMap()

	for dashboard, plugins := range plugins.BinaryData {
		var dashboardPlugins v1beta1.PluginList

		err = json.Unmarshal(plugins, &dashboardPlugins)
		if err != nil {
			log.Error(err, "error consolidating plugins", "dashboard", dashboard)
			return v1beta1.OperatorStageResultFailed, err
		}

		pm.Merge(dashboardPlugins)
	}

	vars.Plugins = pm.GetPluginList().String()

	return v1beta1.OperatorStageResultSuccess, nil
}
