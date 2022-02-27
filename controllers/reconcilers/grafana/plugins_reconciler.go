package grafana

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	PLUGINS_HASH_KEY = "PLUGINS_HASH"
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
	logger := log.FromContext(ctx)

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

		// no plugins yet, assign plugins hash to empty string
		vars.PluginsHash = ""

		return v1beta1.OperatorStageResultSuccess, nil
	} else if err != nil {
		logger.Error(err, "error getting plugins config map", "name", plugins.Name, "namespace", plugins.Namespace)
		return v1beta1.OperatorStageResultFailed, err
	}

	// plugins config map found, but may be empty
	if plugins.Data == nil {
		vars.PluginsHash = ""
		return v1beta1.OperatorStageResultSuccess, nil
	}

	if val, ok := plugins.Data[PLUGINS_HASH_KEY]; ok {
		vars.PluginsHash = val
	} else {
		vars.PluginsHash = ""
	}

	return v1beta1.OperatorStageResultSuccess, nil
}
