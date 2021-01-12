package common

import (
	"errors"
	v1 "github.com/openshift/api/route/v1"
	v12 "k8s.io/api/apps/v1"
	"k8s.io/api/extensions/v1beta1"
)

const (
	ConditionStatusSuccess = "True"
)

func IsRouteReady(route *v1.Route) bool {
	if route == nil {
		return false
	}
	// A route has a an array of Ingress where each have an array of conditions
	for _, ingress := range route.Status.Ingress {
		for _, condition := range ingress.Conditions {
			// A successful route will have the admitted condition type as true
			if condition.Type == v1.RouteAdmitted && condition.Status != ConditionStatusSuccess {
				return false
			}
		}
	}
	return true
}

func IsIngressReady(ingress *v1beta1.Ingress) bool {
	if ingress == nil {
		return false
	}

	return len(ingress.Status.LoadBalancer.Ingress) > 0
}

func IsDeploymentReady(deployment *v12.Deployment) (bool, error) {
	if deployment == nil {
		return false, nil
	}
	// A deployment has an array of conditions
	for _, condition := range deployment.Status.Conditions {
		// One failure condition exists, if this exists, return the Reason
		if condition.Type == v12.DeploymentReplicaFailure {
			return false, errors.New(condition.Reason)
			// A successful deployment will have the progressing condition type as true
		} else if condition.Type == v12.DeploymentProgressing && condition.Status != ConditionStatusSuccess {
			return false, nil
		}
	}

	return deployment.Status.ReadyReplicas == deployment.Status.Replicas, nil
}
