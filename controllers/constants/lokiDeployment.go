package model

import (
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v13 "k8s.io/api/core/v1"
)

import (
	v1 "k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getExternal(cr *v1alpha1.Loki) *v1alpha1.LokiExternal {
	if cr.Spec.External != nil {
		return cr.Spec.External
	}
	return nil
}




//
func getLokiDeploymentSpec(cr *v1alpha1.Loki) v1.DeploymentSpec {
	return v1.DeploymentSpec{
		Selector: &v12.LabelSelector{
			MatchLabels: map[string]string{
				"app": LokiPodLabel,
			},
		},
		Template: v13.PodTemplateSpec{
			ObjectMeta: v12.ObjectMeta{
				Name:        LokiDeploymentName,
			},
			Spec: v13.PodSpec{
			},
		},
		Strategy: v1.DeploymentStrategy{
			Type:          "RollingUpdate",
			RollingUpdate: getRollingUpdateStrategy(),
		},
	}
}

func LokiDeployment(cr *v1alpha1.Loki) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: v12.ObjectMeta{
			Name:      LokiDeploymentName,
			Namespace: cr.Namespace,
		},
		Spec: getLokiDeploymentSpec(cr),
	}
}

func LokiDeploymentSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      LokiDeploymentName,
	}
}

func LokiDeploymentReconciled(cr *v1alpha1.Loki, currentState *v1.Deployment) *v1.Deployment {
	reconciled := currentState.DeepCopy()
	reconciled.Spec = getLokiDeploymentSpec(cr)
	return reconciled
}
