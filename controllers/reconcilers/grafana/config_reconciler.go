package grafana

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	"github.com/grafana-operator/grafana-operator/v5/controllers/config"
	v1 "k8s.io/api/core/v1"
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

func GetGrafanaIniMeta(cr *v1beta1.Grafana) *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-grafana-ini", cr.Name),
			Namespace: cr.Namespace,
		},
	}
}

func (r *ConfigReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana) error {
	_ = log.FromContext(ctx)

	config, _ := config.WriteIni(cr.Spec.Config)
	// vars.ConfigHash = hash // TODO

	configMap := GetGrafanaIniMeta(cr)
	if err := controllerutil.SetControllerReference(cr, configMap, r.Scheme); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["grafana.ini"] = config
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create or update: %w", err)
	}
	return nil
}
