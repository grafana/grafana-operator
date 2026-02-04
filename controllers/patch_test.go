package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePatches(t *testing.T) {
	cases := []struct {
		p      *v1beta1.Patch
		target any
	}{
		{
			p:      &v1beta1.Patch{},
			target: nil,
		},
		{
			p: &v1beta1.Patch{
				Scripts: []string{
					`.spec.title="foo"`,
				},
			},
			target: nil,
		},
		{
			p: &v1beta1.Patch{
				Scripts: []string{
					".spec.title=-",
				},
			},
			target: &gojq.ParseError{},
		},
	}
	for _, c := range cases {
		_, err := ParsePatches(c.p)
		if c.target == nil {
			assert.NoError(t, err)
		} else {
			assert.ErrorAs(t, err, &c.target)
		}
	}
}

func TestApplyPatches(t *testing.T) {
	cases := []struct {
		patches  []string
		env      []string
		original map[string]any
		expected map[string]any
	}{
		{
			patches:  []string{`.foo="bar"`},
			env:      nil,
			original: map[string]any{},
			expected: map[string]any{"foo": "bar"},
		},
		{
			patches:  []string{`.foo="baz"`},
			env:      nil,
			original: map[string]any{"foo": "bar"},
			expected: map[string]any{"foo": "baz"},
		},
		{
			patches:  []string{`.foo="bar"`, `.foo="baz"`},
			env:      nil,
			original: map[string]any{"foo": "start"},
			expected: map[string]any{"foo": "baz"},
		},
		{
			patches:  []string{`.foo=env.E`},
			env:      []string{"E=bar"},
			original: map[string]any{},
			expected: map[string]any{"foo": "bar"},
		},
		{
			patches:  []string{`.foo=env.E`},
			env:      nil,
			original: map[string]any{},
			expected: map[string]any{"foo": nil},
		},
	}
	for _, c := range cases {
		compiled, err := ParsePatches(&v1beta1.Patch{
			Scripts: c.patches,
		})
		require.NoError(t, err)
		out, err := ApplyPatch(compiled, c.original, c.env)
		require.NoError(t, err)
		assert.Equal(t, c.expected, out)
	}
}
