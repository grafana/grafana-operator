package autodetect

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

func TestDetectPlatformBasedOnAvailableAPIGroups(t *testing.T) {
	for _, tt := range []struct {
		apiGroupList *metav1.APIGroupList
		expected     bool
	}{
		{
			&metav1.APIGroupList{},
			false,
		},
		{
			&metav1.APIGroupList{
				Groups: []metav1.APIGroup{
					{
						Name: "route.openshift.io",
					},
				},
			},
			true,
		},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			output, err := json.Marshal(tt.apiGroupList)
			assert.NoError(t, err)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, err = w.Write(output)
			assert.NoError(t, err)
		}))
		defer server.Close()

		autoDetect, err := NewAutoDetect(&rest.Config{Host: server.URL})
		require.NoError(t, err)

		plt, err := autoDetect.IsOpenshift()
		require.NoError(t, err)
		assert.Equal(t, tt.expected, plt)
	}
}

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
