package v1beta1

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestGrafanaDatasources_expandVariables(t *testing.T) {
	type testcase struct {
		name      string
		variables map[string][]byte
		in        GrafanaDatasource
		out       []byte
	}

	testcases := []testcase{
		{
			name: "basic replacement",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			in: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						User: "${PROMETHEUS_USERNAME}",
					},
				},
			},
			out: []byte("{\"name\":\"prometheus\",\"user\":\"root\"}"),
		},
		{
			name: "replacement without curly braces",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			in: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					Datasource: &GrafanaDatasourceInternal{
						Name: "prometheus",
						User: "$PROMETHEUS_USERNAME",
					},
				},
			},
			out: []byte("{\"name\":\"prometheus\",\"user\":\"root\"}"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := tc.in.ExpandVariables(tc.variables)
			if err != nil {
				t.Fatal(err)
			}

			if bytes.Compare(b, tc.out) != 0 {
				t.Fatal(errors.New(fmt.Sprintf("expected %v, but got %v", string(tc.out), string(b))))
			}
		})
	}
}
