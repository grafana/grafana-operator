package common

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

var ControllerEvents = make(chan ControllerState, 1)

type ControllerState struct {
	DashboardSelectors         []*v1.LabelSelector
	DashboardNamespaceSelector *v1.LabelSelector
	AdminUsername              string
	AdminPassword              string
	AdminUrl                   string
	GrafanaReady               bool
	ClientTimeout              int
	FixAnnotations             bool
	FixHeights                 bool
}
