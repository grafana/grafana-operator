package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGrafanaStatusListSilence(t *testing.T) {
	t.Run("&GrafanaSilence{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaSilence{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "GrafanaSilence does not have a case in Grafana.Status.StatusList")
	})
}
