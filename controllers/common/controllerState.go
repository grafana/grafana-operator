package common

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

var ControllerEvents = make(chan ControllerState, 1)
var DashboardControllerEvents = make(chan ControllerState, 1)
var DashboardFolderControllerEvents = make(chan ControllerState, 1)
var DatasourceControllerEvents = make(chan ControllerState, 1)
var NotificationChannelControllerEvents = make(chan ControllerState, 1)

type ControllerState struct {
	DashboardSelectors            []*v1.LabelSelector
	DashboardNamespaceSelector    *v1.LabelSelector
	DashboardContentCacheDuration *v1.Duration
	AdminUrl                      string
	GrafanaReady                  bool
	ClientTimeout                 int
}
