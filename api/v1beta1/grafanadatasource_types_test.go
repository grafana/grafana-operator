package v1beta1

import (
	"testing"

	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
)

func TestGrafanaDatasources_expandVariables(t *testing.T) {
	type testcase struct {
		name      string
		variables map[string][]byte
		input     GrafanaDatasource
		match     types.GomegaMatcher
	}

	testcases := []testcase{
		{
			name: "basic replacement",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			input: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					DataSource: GrafanaDatasourceDataSource{
						Name: "prometheus",
						User: "${PROMETHEUS_USERNAME}",
					},
				},
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Name": Equal("prometheus"),
				"User": Equal("root"),
			}),
		},
		{
			name: "replacement without curly braces",
			variables: map[string][]byte{
				"PROMETHEUS_USERNAME": []byte("root"),
			},
			input: GrafanaDatasource{
				Spec: GrafanaDatasourceSpec{
					DataSource: GrafanaDatasourceDataSource{
						Name: "prometheus",
						User: "$PROMETHEUS_USERNAME",
					},
				},
			},
			match: MatchFields(IgnoreExtras, Fields{
				"Name": Equal("prometheus"),
				"User": Equal("root"),
			}),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			out, err := testcase.input.ExpandVariables(testcase.variables)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(out).NotTo(BeNil())
			g.Expect(*out).To(testcase.match)
		})
	}
}
