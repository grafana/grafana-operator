package api

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CommonResource interface {
	MatchConditions() (*metav1.LabelSelector, string, *bool)
}
