package grafana

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PluginsReconciler struct {
	client client.Client
}

func NewPluginsReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PluginsReconciler{
		client: client,
	}
}

func (r *PluginsReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx).WithName("PluginsReconciler")

	plugins := model.GetPluginsConfigMap(cr, scheme)
	selector := client.ObjectKey{
		Namespace: plugins.Namespace,
		Name:      plugins.Name,
	}

	err := r.client.Get(ctx, selector, plugins)

	// plugins config map not found, we need to create it
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(ctx, plugins)
		if err != nil {
			logger.Error(err, "error creating plugins config map", "name", plugins.Name, "namespace", plugins.Namespace)
			return v1beta1.OperatorStageResultFailed, err
		}

		// no plugins yet, assign plugins to empty string
		vars.Plugins = ""

		return v1beta1.OperatorStageResultSuccess, nil
	} else if err != nil {
		logger.Error(err, "error getting plugins config map", "name", plugins.Name, "namespace", plugins.Namespace)
		return v1beta1.OperatorStageResultFailed, err
	}

	// plugins config map found, but may be empty
	if plugins.BinaryData == nil || len(plugins.BinaryData) == 0 {
		vars.Plugins = ""
		return v1beta1.OperatorStageResultSuccess, nil
	}

	var consolidatedPlugins v1beta1.PluginList
	for dashboard, plugins := range plugins.BinaryData {
		var dashboardPlugins v1beta1.PluginList
		err = json.Unmarshal(plugins, &dashboardPlugins)
		if err != nil {
			logger.Error(err, "error consolidating plugins", "dashboard", dashboard)
			return v1beta1.OperatorStageResultFailed, err
		}

		for _, plugin := range dashboardPlugins {
			// new plugin
			plugin := plugin
			if !consolidatedPlugins.HasSomeVersionOf(&plugin) {
				consolidatedPlugins = append(consolidatedPlugins, plugin)
				continue
			}

			// newer version of plugin already installed
			hasNewer, err := consolidatedPlugins.HasNewerVersionOf(&plugin)
			if err != nil {
				logger.Error(err, "error checking existing plugins", "dashboard", dashboard)
				return v1beta1.OperatorStageResultFailed, err
			}

			if hasNewer {
				logger.Info("skipping plugin", "dashboard", dashboard, "plugin",
					plugin.Name, "version", plugin.Version)
				continue
			}

			// duplicate plugin
			if consolidatedPlugins.HasExactVersionOf(&plugin) {
				continue
			}

			// some version is installed, but it is not newer and it is not the same: must be older
			consolidatedPlugins.Update(&plugin)
		}
	}

	vars.Plugins = consolidatedPlugins.String()
	return v1beta1.OperatorStageResultSuccess, nil
}
