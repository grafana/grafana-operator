package autodetect

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = Describe("AutoDetect", func() {
	t := GinkgoT()

	Context("HasAPIGroup correctly discovers apiGroup presence when", func() {
		tests := []struct {
			name     string
			apiGroup string
			want     bool
		}{
			{
				name:     "apiVersion exists",
				apiGroup: "gateway.networking.k8s.io",
				want:     true,
			},
			{
				name:     "apiVersion does not exist",
				apiGroup: "non.existent.api.io",
				want:     false,
			},
		}

		for _, tt := range tests {
			It(tt.name, func() {
				autoDetect, err := NewAutoDetect(cfg)
				require.NoError(t, err)
				require.NotNil(t, autoDetect)

				got, err := autoDetect.HasAPIGroup(tt.apiGroup)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})

	Context("HasKind correctly discovers CRD presence when", func() {
		tests := []struct {
			name       string
			apiVersion string
			kind       string
			want       bool
		}{
			{
				name:       "apiVersion and kind exist",
				apiVersion: "gateway.networking.k8s.io/v1",
				kind:       "HTTPRoute",
				want:       true,
			},
			{
				name:       "kind does not exist",
				apiVersion: "gateway.networking.k8s.io/v1",
				kind:       "NonExistentKind",
				want:       false,
			},
			{
				name:       "apiVersion does not exist",
				apiVersion: "non.existent.api.io/v1",
				kind:       "HTTPRoute",
				want:       false,
			},
		}

		for _, tt := range tests {
			It(tt.name, func() {
				autoDetect, err := NewAutoDetect(cfg)
				require.NoError(t, err)
				require.NotNil(t, autoDetect)

				got, err := autoDetect.HasKind(tt.apiVersion, tt.kind)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
})
