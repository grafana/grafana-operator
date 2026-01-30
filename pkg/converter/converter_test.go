/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package converter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewConverter(t *testing.T) {
	opts := ConverterOptions{
		Namespace: "test-namespace",
		AdditionalLabels: map[string]string{
			"app": "test",
		},
	}

	got := NewConverter(opts)
	require.NotNil(t, got)
}

func TestConvertFile(t *testing.T) {
	inputYAML := `
groups:
- name: test-alerts
  rules:
  - alert: HighCPUUsage
    expr: cpu_usage > 80
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: High CPU usage detected
  - alert: HighMemory
    expr: memory_usage > 90
    for: 10m
    labels:
      severity: warning
    annotations:
      summary: High memory usage detected
`

	tmpFile, err := os.CreateTemp("", "test-rules-*.yaml")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(inputYAML)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	conv := NewConverter(ConverterOptions{
		Namespace: "monitoring",
		AdditionalLabels: map[string]string{
			"managed_by": "grafana-operator",
		},
		AdditionalAnnotations: map[string]string{
			"imported_by": "grafana-alert-importer",
		},
	})

	groups, err := conv.ConvertFile(tmpFile.Name())
	require.NoError(t, err)
	require.Len(t, groups, 1)

	got := groups[0]
	assert.Equal(t, "test-alerts", got.Name)
	assert.Equal(t, "monitoring", got.Namespace)
	assert.Equal(t, "grafana-operator", got.Labels["managed_by"])
	require.Len(t, got.Spec.Rules, 2)

	rule := got.Spec.Rules[0]
	assert.Equal(t, "HighCPUUsage", rule.Title)
	assert.Equal(t, "cpu_usage > 80", rule.Condition)
	require.NotNil(t, rule.For)
	assert.Equal(t, "5m", *rule.For)
	assert.Equal(t, "critical", rule.Labels["severity"])
	assert.Equal(t, "High CPU usage detected", rule.Annotations["summary"])
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple-name",
			input: "simple-name",
			want:  "simple-name",
		},
		{
			name:  "uppercase",
			input: "UPPERCASE",
			want:  "uppercase",
		},
		{
			name:  "with underscore",
			input: "with_underscore",
			want:  "with-underscore",
		},
		{
			name:  "with dot",
			input: "with.dot",
			want:  "with-dot",
		},
		{
			name:  "with spaces",
			input: "name with spaces",
			want:  "name-with-spaces",
		},
		{
			name:  "numeric prefix",
			input: "123test",
			want:  "123test",
		},
		{
			name:  "leading hyphen",
			input: "-leading-hyphen",
			want:  "leading-hyphen",
		},
		{
			name:  "trailing hyphen",
			input: "trailing-hyphen-",
			want:  "trailing-hyphen",
		},
		{
			name:  "mixed case underscore dot",
			input: "mixed-CASE_123.dot",
			want:  "mixed-case-123-dot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "one minute",
			input: "1m",
			want:  "1m0s",
		},
		{
			name:  "five minutes thirty seconds",
			input: "5m30s",
			want:  "5m30s",
		},
		{
			name:  "one hour",
			input: "1h",
			want:  "1h0m0s",
		},
		{
			name:  "empty defaults to 1m",
			input: "",
			want:  "1m0s",
		},
		{
			name:  "invalid defaults to 1m",
			input: "invalid",
			want:  "1m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.input)
			assert.Equal(t, tt.want, got.Duration.String())
		})
	}
}

func TestConvertDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "converter-test")
	require.NoError(t, err)

	defer os.RemoveAll(tmpDir)

	files := map[string]string{
		"alerts1.yaml": `
groups:
- name: group1
  rules:
  - alert: Alert1
    expr: up == 0
`,
		"alerts2.yaml": `
groups:
- name: group2
  rules:
  - alert: Alert2
    expr: up == 1
`,
		"readme.txt": `
This is not a YAML file and should be skipped
`,
	}

	for filename, content := range files {
		path := filepath.Join(tmpDir, filename)
		err := os.WriteFile(path, []byte(content), 0o600)
		require.NoError(t, err)
	}

	conv := NewConverter(ConverterOptions{
		Namespace: "test-ns",
	})

	groups, err := conv.ConvertDirectory(tmpDir)
	require.NoError(t, err)
	assert.Len(t, groups, 2)
}

func TestConvertFileWithInvalidYAML(t *testing.T) {
	inputYAML := `
groups:
- name: test
  invalid: [[[
`

	tmpFile, err := os.CreateTemp("", "invalid-*.yaml")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(inputYAML)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	conv := NewConverter(ConverterOptions{})

	_, err = conv.ConvertFile(tmpFile.Name())
	assert.Error(t, err)
}

func TestConvertFileWithInvalidPrometheusRules(t *testing.T) {
	inputYAML := `
groups:
- name: test
  rules:
  - alert: MissingExpr
    labels:
      severity: warning
`

	tmpFile, err := os.CreateTemp("", "invalid-rule-*.yaml")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(inputYAML)
	require.NoError(t, err)
	require.NoError(t, tmpFile.Close())

	conv := NewConverter(ConverterOptions{})

	_, err = conv.ConvertFile(tmpFile.Name())
	assert.Error(t, err)
}

func TestConverterOptions(t *testing.T) {
	selector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"grafana": "instance1",
		},
	}

	opts := ConverterOptions{
		Namespace:        "monitoring",
		InstanceSelector: selector,
		AdditionalLabels: map[string]string{
			"team": "sre",
		},
		AdditionalAnnotations: map[string]string{
			"imported": "true",
		},
		FolderRef:    "my-folder",
		FolderUID:    "abc123",
		ResyncPeriod: "5m",
	}

	conv := NewConverter(opts)

	assert.Equal(t, "monitoring", conv.opts.Namespace)
	assert.Equal(t, selector, conv.opts.InstanceSelector)
	assert.Equal(t, "sre", conv.opts.AdditionalLabels["team"])
	assert.Equal(t, "true", conv.opts.AdditionalAnnotations["imported"])
}
