package client

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestAlertmanagerStatusReceivesAlerts(t *testing.T) {
	tests := []struct {
		name   string
		status AlertmanagerStatus
		want   bool
	}{
		{
			name:   "no external alertmanager",
			status: AlertmanagerStatus{AlertmanagersChoice: v1beta1.AlertmanagerExternal, NumExternalAlertmanagers: 0},
			want:   false,
		},
		{
			name:   "external configured but choice internal",
			status: AlertmanagerStatus{AlertmanagersChoice: v1beta1.AlertmanagerInternal, NumExternalAlertmanagers: 1},
			want:   false,
		},
		{
			name:   "external configured and choice external",
			status: AlertmanagerStatus{AlertmanagersChoice: v1beta1.AlertmanagerExternal, NumExternalAlertmanagers: 1},
			want:   true,
		},
		{
			name:   "external configured and choice all",
			status: AlertmanagerStatus{AlertmanagersChoice: v1beta1.AlertmanagerAll, NumExternalAlertmanagers: 2},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.ReceivesAlerts())
		})
	}
}
