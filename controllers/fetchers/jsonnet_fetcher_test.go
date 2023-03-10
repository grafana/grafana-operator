package fetchers

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/grafana-operator/grafana-operator-experimental/api/v1beta1"
	"github.com/grafana-operator/grafana-operator-experimental/embeds"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func normalizeAndCompareJson(json1, json2 []byte) bool {
	var data1, data2 map[string]interface{}

	if err := json.Unmarshal(json1, &data1); err != nil {
		return false
	}
	if err := json.Unmarshal(json2, &data2); err != nil {
		return false
	}

	normalized1, err := json.Marshal(data1)
	if err != nil {
		return false
	}
	normalized2, err := json.Marshal(data2)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(normalized1, normalized2)
}

func TestFetchJsonnet(t *testing.T) {
	tests := []struct {
		name          string
		dashboard     *v1beta1.GrafanaDashboard
		libsonnet     embed.FS
		expected      []byte
		expectedError error
	}{
		{
			name: "Successful Jsonnet Evaluation",
			dashboard: &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grafanadashboard-jsonnet",
					Namespace: "grafana",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					Jsonnet: string(embeds.TestDashboardEmbed),
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			},
			libsonnet:     embeds.GrafonnetEmbed,
			expected:      embeds.TestDashboardEmbedExpectedJSON,
			expectedError: nil,
		},
		{
			name: "Empty Jsonnet Content",
			dashboard: &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{Name: "dashboard-1"},
				Spec: v1beta1.GrafanaDashboardSpec{
					Jsonnet: "",
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			},
			libsonnet:     embeds.GrafonnetEmbed,
			expected:      nil,
			expectedError: errors.New("no jsonnet Content Found, nil or empty string"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := FetchJsonnet(test.dashboard, test.libsonnet)

			if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.expectedError) {
				t.Errorf("expected error %v, but got %v", test.expectedError, err)
			}

			if test.expected != nil && !normalizeAndCompareJson(test.expected, result) {
				t.Errorf("expected string %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}
