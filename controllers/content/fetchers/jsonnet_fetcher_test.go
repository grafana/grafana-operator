package fetchers

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/grafana/grafana-operator/v5/controllers/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/embeds"
)

func setup(t *testing.T) {
	t.Helper()
	require.NoError(t, os.Mkdir(config.GrafanaDashboardsRuntimeBuild, os.ModePerm))
}

func teardown(t *testing.T) {
	t.Helper()
	require.NoError(t, os.RemoveAll(config.GrafanaDashboardsRuntimeBuild))
}

func normalizeAndCompareJSON(json1, json2 []byte) bool {
	var data1, data2 map[string]any

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
		name    string
		jsonnet string
		envs    map[string]string
		want    []byte
	}{
		{
			name:    "Successful Jsonnet Evaluation",
			jsonnet: string(embeds.TestDashboardEmbed),
			envs:    map[string]string{},
			want:    embeds.TestDashboardEmbedExpectedJSON,
		},
		{
			name:    "Successful Jsonnet Evaluation with non-empty envs",
			jsonnet: string(embeds.TestDashboardEmbedWithEnv),
			want:    embeds.TestDashboardEmbedWithEnvExpectedJSON,
			envs: map[string]string{
				"TEST_ENV": "123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := &v1beta1.GrafanaDashboard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "grafanadashboard-jsonnet",
					Namespace: "grafana",
				},
				Spec: v1beta1.GrafanaDashboardSpec{
					GrafanaContentSpec: v1beta1.GrafanaContentSpec{
						Jsonnet: tt.jsonnet,
					},
				},
			}

			got, err := FetchJsonnet(cr, tt.envs, embeds.GrafonnetEmbed)
			require.NoError(t, err)

			assert.JSONEq(t, string(tt.want), string(got))
		})
	}

	t.Run("Empty Jsonnet Content", func(t *testing.T) {
		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grafanadashboard-jsonnet",
				Namespace: "grafana",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					Jsonnet: "",
				},
			},
		}

		got, err := FetchJsonnet(cr, map[string]string{}, embeds.GrafonnetEmbed)
		assert.Nil(t, got)
		require.ErrorIs(t, err, errJsonnetNoContent)
	})
}

func TestBuildProjectAndFetchJsonnetFrom(t *testing.T) {
	setup(t)
	defer teardown(t)

	t.Run("Successful Jsonnet Evaluation with jsonnet build", func(t *testing.T) {
		cr := &v1beta1.GrafanaDashboard{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "grafanadashboard-jsonnet",
				Namespace: "grafana",
			},
			Spec: v1beta1.GrafanaDashboardSpec{
				GrafanaContentSpec: v1beta1.GrafanaContentSpec{
					JsonnetProjectBuild: &v1beta1.JsonnetProjectBuild{
						JPath:              []string{"/testing/jsonnetProjectWithRuntimeRaw"},
						FileName:           "testing/jsonnetProjectWithRuntimeRaw/dashboard_with_envs.jsonnet",
						GzipJsonnetProject: embeds.TestJsonnetProjectBuildFolderGzip,
					},
				},
			},
		}

		envs := map[string]string{
			"TEST_ENV": "123",
		}

		want := []byte("{\n    \"env\" : \"123\"   \n}")

		got, err := BuildProjectAndFetchJsonnetFrom(cr, envs)
		require.NoError(t, err)

		assert.JSONEq(t, string(want), string(got))
	})
}

func TestGetJsonProjectBuildRoundName(t *testing.T) {
	roundName, err := getJSONProjectBuildRoundName("test")
	require.NoError(t, err)

	roundNameParts := strings.Split(roundName, "-")
	require.Len(t, roundNameParts, 3)
	require.Equal(t, "test", roundNameParts[0])
}

func TestGetGzipArchiveFileNameWithExtension(t *testing.T) {
	archiveName := getGzipArchiveFileNameWithExtension("test")
	require.Equal(t, "test.tar.gz", archiveName)
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
