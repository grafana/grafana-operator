package controllers

import (
	"testing"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/itchyny/gojq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePatches(t *testing.T) {
	tests := []struct {
		p    *v1beta1.Patch
		want any
	}{
		{
			p:    &v1beta1.Patch{},
			want: nil,
		},
		{
			p: &v1beta1.Patch{
				Scripts: []string{
					`.spec.title="foo"`,
				},
			},
			want: nil,
		},
		{
			p: &v1beta1.Patch{
				Scripts: []string{
					".spec.title=-",
				},
			},
			want: &gojq.ParseError{},
		},
	}
	for _, tt := range tests {
		_, err := ParsePatches(tt.p)
		if tt.want == nil {
			assert.NoError(t, err)
		} else {
			assert.ErrorAs(t, err, &tt.want)
		}
	}
}

func TestApplyPatches(t *testing.T) {
	tests := []struct {
		patches  []string
		env      []string
		original map[string]any
		want     map[string]any
	}{
		{
			patches:  []string{`.foo="bar"`},
			env:      nil,
			original: map[string]any{},
			want:     map[string]any{"foo": "bar"},
		},
		{
			patches:  []string{`.foo="baz"`},
			env:      nil,
			original: map[string]any{"foo": "bar"},
			want:     map[string]any{"foo": "baz"},
		},
		{
			patches:  []string{`.foo="bar"`, `.foo="baz"`},
			env:      nil,
			original: map[string]any{"foo": "start"},
			want:     map[string]any{"foo": "baz"},
		},
		{
			patches:  []string{`.foo=env.E`},
			env:      []string{"E=bar"},
			original: map[string]any{},
			want:     map[string]any{"foo": "bar"},
		},
		{
			patches:  []string{`.foo=env.E`},
			env:      nil,
			original: map[string]any{},
			want:     map[string]any{"foo": nil},
		},
	}
	for _, tt := range tests {
		compiled, err := ParsePatches(&v1beta1.Patch{
			Scripts: tt.patches,
		})
		require.NoError(t, err)
		got, err := ApplyPatch(compiled, tt.original, tt.env)
		require.NoError(t, err)
		assert.Equal(t, tt.want, got)
	}
}
