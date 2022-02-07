package grafana

import (
	"context"
	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/model"
	"github.com/grafana-operator/grafana-operator-experimental/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

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

type IngressAccountReconciler struct {
	client client.Client
}

func IngressAccountReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &ServiceAccountReconciler{
		client: client,
	}
}

func (r *IngressAccountReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, status *v1beta1.GrafanaStatus, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	logger := log.FromContext(ctx)

	sa := model.GetGrafanaServiceAccount(cr, scheme)

	if cr.SkipCreateAdminAccount() {
		logger.Info("skip creating grafana service account")
		return v1beta1.OperatorStageResultSuccess, nil
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.client, sa, func() error {
		sa.Labels = getServiceAccountLabels(cr)
		sa.Annotations = getServiceAccountAnnotations(cr, sa.Annotations)
		sa.ImagePullSecrets = getServiceAccountImagePullSecrets(cr, sa.ImagePullSecrets)
		return nil
	})

	if err != nil {
		return v1beta1.OperatorStageResultFailed, err
	}

	return v1beta1.OperatorStageResultSuccess, nil
}

func getServiceAccountLabels(cr *v1beta1.Grafana) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return nil
	}
	return cr.Spec.ServiceAccount.Labels
}

func getServiceAccountAnnotations(cr *v1beta1.Grafana, existing map[string]string) map[string]string {
	if cr.Spec.ServiceAccount == nil {
		return existing
	}
	return model.MergeAnnotations(cr.Spec.ServiceAccount.Annotations, existing)
}

func getServiceAccountImagePullSecrets(cr *v1beta1.Grafana, existing []v1.LocalObjectReference) []v1.LocalObjectReference {
	if cr.Spec.ServiceAccount == nil {
		return existing
	}
	return mergeImagePullSecrets(cr.Spec.ServiceAccount.ImagePullSecrets, existing)
}

func mergeImagePullSecrets(requested []v1.LocalObjectReference, existing []v1.LocalObjectReference) []v1.LocalObjectReference {
	appendIfAbsent := func(secrets []v1.LocalObjectReference, secret v1.LocalObjectReference) []v1.LocalObjectReference {
		for _, s := range secrets {
			if s.Name == secret.Name {
				return secrets
			}
		}
		return append(secrets, secret)
	}

	for _, s := range requested {
		existing = appendIfAbsent(existing, s)
	}

	return existing
}
