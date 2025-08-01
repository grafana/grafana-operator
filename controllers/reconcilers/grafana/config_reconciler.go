package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ConfigReconciler struct {
	client client.Client
}

func NewConfigReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &ConfigReconciler{
		client: client,
	}
}

func (r *ConfigReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	_ = logf.FromContext(ctx)

	cfg := config.WriteIni(cr.Spec.Config)
	vars.ConfigHash = config.GetHash(cfg)

	configMap := model.GetGrafanaConfigMap(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}

		configMap.Data["grafana.ini"] = cfg

		if scheme != nil {
			err := controllerutil.SetControllerReference(cr, configMap, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(configMap, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}
