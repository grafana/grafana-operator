package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrafanaStatusListNotificationPolicyRoute(t *testing.T) {
	t.Run("&NotificationPolicyRoute{} does not map to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaNotificationPolicyRoute{}
		_, _, err := g.Status.StatusList(arg)
		assert.Error(t, err, "NotificationPolicyRoute should not have a case in Grafana.Status.StatusList")
	})
}
