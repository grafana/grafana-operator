package fetchers

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/stretchr/testify/require"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/embeds"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setup(t *testing.T) {
	t.Helper()
	require.NoError(t, os.Mkdir(config.GrafanaDashboardsRuntimeBuild, os.ModePerm))
}

func teardown(t *testing.T) {
	t.Helper()
	require.NoError(t, os.RemoveAll(config.GrafanaDashboardsRuntimeBuild))
}

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
		envs          map[string]string
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
			envs:          map[string]string{},
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
			envs:          map[string]string{},
			expectedError: errors.New("no jsonnet Content Found, nil or empty string"),
		},
		{
			name: "Successful Jsonnet Evaluation with non-empty envs",
			dashboard: &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grafanadashboard-jsonnet",
					Namespace: "grafana",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					Jsonnet: string(embeds.TestDashboardEmbedWithEnv),
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			},
			libsonnet: embeds.GrafonnetEmbed,
			expected:  embeds.TestDashboardEmbedWithEnvExpectedJSON,
			envs: map[string]string{
				"TEST_ENV": "123",
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := FetchJsonnet(test.dashboard, test.envs, test.libsonnet)

			if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.expectedError) {
				t.Errorf("expected error %v, but got %v", test.expectedError, err)
			}

			if test.expected != nil && !normalizeAndCompareJson(test.expected, result) {
				t.Errorf("expected string %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestBuildProjectAndFetchJsonnetFrom(t *testing.T) {
	setup(t)
	defer teardown(t)

	tests := []struct {
		name          string
		dashboard     *v1beta1.GrafanaDashboard
		libsonnet     embed.FS
		expected      []byte
		envs          map[string]string
		expectedError error
	}{
		{
			name: "Successful Jsonnet Evaluation with jsonnet build",
			dashboard: &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grafanadashboard-jsonnet",
					Namespace: "grafana",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					JsonnetProjectBuild: &v1beta1.JsonnetProjectBuild{
						JPath:              []string{"/testing/jsonnetProjectWithRuntimeRaw"},
						FileName:           "testing/jsonnetProjectWithRuntimeRaw/dashboard_with_envs.jsonnet",
						GzipJsonnetProject: embeds.TestJsonnetProjectBuildFolderGzip,
					},
				},
				Status: v1beta1.GrafanaDashboardStatus{},
			},
			libsonnet: embeds.GrafonnetEmbed,
			expected:  []byte("{\n    \"env\" : \"123\"   \n}"),
			envs: map[string]string{
				"TEST_ENV": "123",
			},
			expectedError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := BuildProjectAndFetchJsonnetFrom(test.dashboard, test.envs)

			if fmt.Sprintf("%v", err) != fmt.Sprintf("%v", test.expectedError) {
				t.Errorf("expected error %v, but got %v", test.expectedError, err)
			}

			if test.expected != nil && !normalizeAndCompareJson(test.expected, result) {
				t.Errorf("expected string %s, but got %s", string(test.expected), string(result))
			}
		})
	}
}

func TestGetJsonProjectBuildRoundName(t *testing.T) {
	roundName, err := getJsonProjectBuildRoundName("test")
	require.NoError(t, err)
	roundNameParts := strings.Split(roundName, "-")
	require.Equal(t, 3, len(roundNameParts))
	require.Equal(t, "test", roundNameParts[0])
}

func TestGetGzipArchiveFileNameWithExtension(t *testing.T) {
	archiveName := getGzipArchiveFileNameWithExtension("test")
	require.Equal(t, archiveName, "test.tar.gz")
}

func TestStoreByteArrayGzipOnDisk(t *testing.T) {
	setup(t)
	defer teardown(t)

	bytesString := "H4sIADxf5mQAA+2R3QqCMBiGPfYqPrwA3damZES3UEd1amtWYi50OiK690ZJYPR3UES05+B7Ybx8P3uzShaFUONSZoKrwPkECKGIMThpeFZE6FlbAFMaYcqiHsOAMGM4coB9ZJsr6kolpVlF5iITy1O56TO2NH3Qp73joj9C1s1fNqJs1kL77ftbZpj/CCm9nz8mpJs/MT7iAHrL9Cf8ef655EkObdowBNi7YPDqMvdi8FZKbas4CLTW/k7Wqp4Ln8tNoBPFV6NmuJhoqvvT5YxPPPcwcN394dsXWSwWi+UVjj9zQO4ACgAA"

	gzipFileName := "test"

	path, err := storeByteArrayGzipOnDisk(gzipFileName, []byte(bytesString))

	targetGzipFilePath := getGzipArchiveFilePath(gzipFileName)
	require.NoError(t, err)
	require.NotEmpty(t, path)
	require.Equal(t, targetGzipFilePath, path)
}
