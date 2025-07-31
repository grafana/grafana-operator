package grafana

import (
	"context"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/model"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type PvcReconciler struct {
	client client.Client
}

func NewPvcReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PvcReconciler{
		client: client,
	}
}

func (r *PvcReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("PvcReconciler")

	if cr.Spec.PersistentVolumeClaim == nil {
		log.Info("skip creating persistent volume claim")
		return v1beta1.OperatorStageResultSuccess, nil
	}

	pvc := model.GetGrafanaDataPVC(cr, scheme)

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, pvc, func() error {
		err := v1beta1.Merge(pvc, cr.Spec.PersistentVolumeClaim)
		if err != nil {
			setInvalidMergeCondition(cr, "PersistentVolumeClaim", err)
			return err
		}

		removeInvalidMergeCondition(cr, "PersistentVolumeClaim")

		if scheme != nil {
			err = controllerutil.SetControllerReference(cr, pvc, scheme)
			if err != nil {
				return err
			}
		}

		model.SetInheritedLabels(pvc, cr.Labels)

		return nil
	})
	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}
