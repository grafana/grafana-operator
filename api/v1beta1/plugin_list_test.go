package v1beta1

import (
	"strings"
	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
)

func TestPluginString(t *testing.T) {
	err := quick.Check(func(a string, b string, c string) bool {
		if strings.Contains(a, ",") || strings.Contains(b, ",") || strings.Contains(c, ",") {
			return true // skip plugins with ,
		}

		pl := PluginList{
			{
				Name:    a,
				Version: "7.2",
			},
			{
				Name:    b,
				Version: "2.2",
			},
			{
				Name:    c,
				Version: "6.7",
			},
		}
		out := pl.String()

		split := strings.Split(out, ",")
		if len(split) != 3 {
			return false
		}

		if split[0] > split[1] {
			return false
		}

		if split[1] > split[2] {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Errorf("plugin list was not sorted: %s", err.Error())
	}
}

func TestPluginSanitize(t *testing.T) {
	pl := PluginList{
		{
			Name:    "plugin-a",
			Version: "1.0.0",
		},
		{
			Name:    "plugin-b",
			Version: "2.0.0",
		},
		{
			Name:    "plugin-a",
			Version: "3.0.0",
		},
	}
	sanitized := pl.Sanitize()
	assert.Len(t, sanitized, 2)
}
