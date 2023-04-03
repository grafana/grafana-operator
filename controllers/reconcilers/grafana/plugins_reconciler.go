package grafana

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type PluginsReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func GetPluginsConfigMapMeta(cr *v1beta1.Grafana) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-plugins", cr.Name),
			Namespace: cr.Namespace,
		},
	}
}

func (r *PluginsReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) error {
	logger := r.Log.WithValues("grafana", client.ObjectKeyFromObject(cr))

	plugins := GetPluginsConfigMapMeta(cr)
	if err := controllerutil.SetOwnerReference(cr, plugins, r.Scheme); err != nil {
		return err // todo: error condition
	}

	err := r.Client.Get(ctx, client.ObjectKeyFromObject(plugins), plugins)

	// plugins config map not found, we need to create it
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(ctx, plugins)
		if err != nil {
			return fmt.Errorf("failed to create plugins configmap: %w", err)
		}

		// no plugins yet, assign plugins to empty string
		// vars.Plugins = "" // TODO

		return nil
	} else if err != nil {
		return fmt.Errorf("faile to get plugins configmap: %w", err)
	}

	// plugins config map found, but may be empty
	if plugins.BinaryData == nil || len(plugins.BinaryData) == 0 {
		// vars.Plugins = "" //TDOD
		return nil
	}

	var consolidatedPlugins v1beta1.PluginList
	for dashboard, plugins := range plugins.BinaryData {
		var dashboardPlugins v1beta1.PluginList
		err = json.Unmarshal(plugins, &dashboardPlugins)
		if err != nil {
			return fmt.Errorf("failed to consolidate plugins: %w", err)
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
				return fmt.Errorf("error checking existing plugins: %w", err)
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

	// vars.Plugins = consolidatedPlugins.String() // TODO
	return nil // todo: successcondition
}
