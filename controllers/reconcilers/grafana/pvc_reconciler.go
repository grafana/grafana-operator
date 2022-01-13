package grafana

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type PvcReconciler struct {
	client client.Client
}

func NewPvcReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &PvcReconciler{
		client: client,
	}
}

func (r *PvcReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx)

	if !cr.UsePersistentVolume() {
		logger.Info("skip creating persistent volume claim")
		return v1beta1.OperatorStageResultSuccess, nil
	}

	pvc := model.GetGrafanaDataPVC(cr, scheme)
	_, err := controllerutil.CreateOrUpdate(ctx, r.client, pvc, func() error {
		pvc.Labels = getPVCLabels(cr)
		pvc.Annotations = getPVCAnnotations(cr, pvc.Annotations)
		pvc.Spec = getPVCSpec(cr)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getPVCSpec(cr *v1beta1.Grafana) v1.PersistentVolumeClaimSpec {
	return v1.PersistentVolumeClaimSpec{
		AccessModes: cr.Spec.DataStorage.AccessModes,
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: cr.Spec.DataStorage.Size,
			},
		},
		StorageClassName: &cr.Spec.DataStorage.Class,
	}
}

func getPVCLabels(cr *v1beta1.Grafana) map[string]string {
	if cr.Spec.DataStorage == nil {
		return nil
	}
	return cr.Spec.DataStorage.Labels
}

func getPVCAnnotations(cr *v1beta1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.DataStorage == nil {
		return existing
	}

	return model.MergeAnnotations(cr.Spec.DataStorage.Annotations, existing)
}
