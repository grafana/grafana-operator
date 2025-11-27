package grafana

import (
	"context"
	"encoding/json"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/dependents"
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

	cm := dependents.GetPluginsConfigMap(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, cm, func() error {
		if scheme != nil {
			err := controllerutil.SetOwnerReference(cr, cm, scheme)
			if err != nil {
				return err
			}
		}

		dependents.SetInheritedLabels(cm, cr.Labels)

		return nil
	})
	if err != nil {
		log.Error(err, "error getting plugins ConfigMap", "name", cm.Name, "namespace", cm.Namespace)
		return v1beta1.OperatorStageResultFailed, err
	}

	// plugins config map found, but may be empty
	if len(cm.BinaryData) == 0 {
		vars.Plugins = ""
		return v1beta1.OperatorStageResultSuccess, nil
	}

	pm := v1beta1.NewPluginMap()

	for k, v := range cm.BinaryData {
		var plugins v1beta1.PluginList

		err = json.Unmarshal(v, &plugins)
		if err != nil {
			log.Error(err, "error consolidating plugins from ConfigMap", "name", cm.Name, "namespace", cm.Namespace, "key", k)
			return v1beta1.OperatorStageResultFailed, err
		}

		pm.Merge(plugins)
	}

	vars.Plugins = pm.GetPluginList().String()

	return v1beta1.OperatorStageResultSuccess, nil
}
