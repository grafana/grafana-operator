package grafana

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	"github.com/grafana-operator/grafana-operator/v5/controllers/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ConfigReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ConfigReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) (*metav1.Condition, error) {
	_ = log.FromContext(ctx)

	config, _ := config.WriteIni(cr.Spec.Config)
	// vars.ConfigHash = hash // TODO

	configMap := model.GetGrafanaConfigMap(cr, r.Scheme)
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["grafana.ini"] = config
		return nil
	})
	if err != nil {
		return nil, err // todo: error condition
	}
	return nil, nil // todo: success condition
}
