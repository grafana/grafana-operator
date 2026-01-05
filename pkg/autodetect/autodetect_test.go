package autodetect

import (
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = Describe("HasAPIGroup", func() {
	t := GinkgoT()

	Context("correctly tests for apiGroup presence when", func() {
		tests := []struct {
			name     string
			apiGroup string
			want     bool
		}{
			{
				name:     "apiGroup exists",
				apiGroup: "gateway.networking.k8s.io",
				want:     true,
			},
			{
				name:     "apiGroup does NOT exist",
				apiGroup: "non.existent.api.io",
				want:     false,
			},
		}

		for _, tt := range tests {
			It(tt.name, func() {
				autoDetect, err := NewAutoDetect(cfgWithCRDs)
				require.NoError(t, err)
				require.NotNil(t, autoDetect)

				got, err := autoDetect.HasAPIGroup(tt.apiGroup)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
})

var _ = Describe("HasKind", func() {
	t := GinkgoT()

	Context("correctly tests for CRD presence when", func() {
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
				name:       "kind does NOT exist",
				apiVersion: "gateway.networking.k8s.io/v1",
				kind:       "NonExistentKind",
				want:       false,
			},
			{
				name:       "apiVersion does NOT exist",
				apiVersion: "non.existent.api.io/v1",
				kind:       "HTTPRoute",
				want:       false,
			},
		}

		for _, tt := range tests {
			It(tt.name, func() {
				autoDetect, err := NewAutoDetect(cfgWithCRDs)
				require.NoError(t, err)
				require.NotNil(t, autoDetect)

				got, err := autoDetect.HasKind(tt.apiVersion, tt.kind)
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)
			})
		}
	})
})

var _ = Describe("IsOpenshift", func() {
	t := GinkgoT()

	Context("correctly tests for CRD presence when", func() {
		It("Route CRD exists", func() {
			autoDetect, err := NewAutoDetect(cfgWithCRDs)
			require.NoError(t, err)
			require.NotNil(t, autoDetect)

			got, err := autoDetect.IsOpenshift()
			require.NoError(t, err)

			assert.True(t, got)
		})

		It("Route CRD does NOT exist", func() {
			autoDetect, err := NewAutoDetect(cfgNoCRDs)
			require.NoError(t, err)
			require.NotNil(t, autoDetect)

			got, err := autoDetect.IsOpenshift()
			require.NoError(t, err)

			assert.False(t, got)
		})
	})
})

var _ = Describe("HasHTTPRouteCRD", func() {
	t := GinkgoT()

	Context("correctly tests for CRD presence when", func() {
		It("HTTPRoute CRD exists", func() {
			autoDetect, err := NewAutoDetect(cfgWithCRDs)
			require.NoError(t, err)
			require.NotNil(t, autoDetect)

			got, err := autoDetect.HasHTTPRouteCRD()
			require.NoError(t, err)

			assert.True(t, got)
		})

		It("HTTPRoute CRD does NOT exist", func() {
			autoDetect, err := NewAutoDetect(cfgNoCRDs)
			require.NoError(t, err)
			require.NotNil(t, autoDetect)

			got, err := autoDetect.HasHTTPRouteCRD()
			require.NoError(t, err)

			assert.False(t, got)
		})
	})
})
