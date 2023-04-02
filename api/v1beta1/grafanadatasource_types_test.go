package v1beta1

import (
	"fmt"
	"reflect"
	"testing"

	gapi "github.com/grafana/grafana-api-golang-client"
)

func TestGrafanaDatasources_expandVariables(t *testing.T) {
	type testcase struct {
		name      string
		variables map[string][]byte
		in        GrafanaDatasource
		out       gapi.DataSource
	}

	testcases := []testcase{
		{
			name: "basic replacement",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			in: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					DataSource: GrafanaDatasourceDataSource{
						Name: "prometheus",
						User: "${PROMETHEUS_USERNAME}",
					},
				},
			},
			out: gapi.DataSource{
				Name: "prometheus",
				User: "root",
			},
		},
		{
			name: "replacement without curly braces",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			in: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					DataSource: GrafanaDatasourceDataSource{
						Name: "prometheus",
						User: "$PROMETHEUS_USERNAME",
					},
				},
			},
			out: gapi.DataSource{
				Name: "prometheus",
				User: "root",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.in.ExpandVariables(tc.variables)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(out, tc.out) {
				t.Error(fmt.Errorf("expected %v, but got %v", tc.out, out))
			}
		})
	}
}
