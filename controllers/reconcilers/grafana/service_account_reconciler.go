package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ServiceAccountReconciler struct {
	client client.Client
}

func NewServiceAccountReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &ServiceAccountReconciler{
		client: client,
	}
}

func (r *ServiceAccountReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	sa := model.GetGrafanaServiceAccount(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, sa, func() error {
		err := v1beta1.Merge(sa, cr.Spec.ServiceAccount)
		if err != nil {
			setInvalidMergeCondition(cr, "ServiceAccount", err)
			return err
		}

		removeInvalidMergeCondition(cr, "ServiceAccount")

		if scheme != nil {
			err = controllerutil.SetControllerReference(cr, sa, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(sa, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}
