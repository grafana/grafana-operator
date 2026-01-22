/*
Copyright 2022.

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

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewConverter(t *testing.T) {
	opts := ConverterOptions{
		Namespace: "test-namespace",
		AdditionalLabels: map[string]string{
			"app": "test",
		},
	}

	converter := NewConverter(opts)
	if converter == nil {
		t.Fatal("expected non-nil converter")
	}
}

func TestConvertFile(t *testing.T) {
	// Create a temporary directory with test data
	tmpDir, err := os.MkdirTemp("", "converter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test input file
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
	inputFile := filepath.Join(tmpDir, "test-rules.yaml")
	if err := os.WriteFile(inputFile, []byte(inputYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	converter := NewConverter(ConverterOptions{
		Namespace: "monitoring",
		AdditionalLabels: map[string]string{
			"managed_by": "grafana-operator",
		},
		AdditionalAnnotations: map[string]string{
			"imported_by": "grafana-alert-importer",
		},
	})

	groups, err := converter.ConvertFile(inputFile)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	group := groups[0]

	// Check metadata
	if group.Name != "test-alerts" {
		t.Errorf("expected name 'test-alerts', got '%s'", group.Name)
	}
	if group.Namespace != "monitoring" {
		t.Errorf("expected namespace 'monitoring', got '%s'", group.Namespace)
	}
	if group.Labels["managed_by"] != "grafana-operator" {
		t.Errorf("expected label 'managed_by' to be 'grafana-operator', got '%s'", group.Labels["managed_by"])
	}

	// Check rules
	if len(group.Spec.Rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(group.Spec.Rules))
	}

	// Check first rule
	rule := group.Spec.Rules[0]
	if rule.Title != "HighCPUUsage" {
		t.Errorf("expected rule title 'HighCPUUsage', got '%s'", rule.Title)
	}
	if rule.Condition != "cpu_usage > 80" {
		t.Errorf("expected condition 'cpu_usage > 80', got '%s'", rule.Condition)
	}
	if rule.For == nil || *rule.For != "5m" {
		t.Errorf("expected 'for' to be '5m', got '%v'", rule.For)
	}
	if rule.Labels["severity"] != "critical" {
		t.Errorf("expected severity label 'critical', got '%s'", rule.Labels["severity"])
	}
	if rule.Annotations["summary"] != "High CPU usage detected" {
		t.Errorf("expected summary annotation 'High CPU usage detected', got '%s'", rule.Annotations["summary"])
	}
}

func TestSanitizeName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"simple-name", "simple-name"},
		{"UPPERCASE", "uppercase"},
		{"with_underscore", "with-underscore"},
		{"with.dot", "with-dot"},
		{"name with spaces", "name-with-spaces"},
		{"123test", "123test"},
		{"-leading-hyphen", "leading-hyphen"},
		{"trailing-hyphen-", "trailing-hyphen"},
		{"mixed-CASE_123.dot", "mixed-case-123-dot"},
	}

	for _, tc := range testCases {
		result := sanitizeName(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeName(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"1m", "1m0s"},
		{"5m30s", "5m30s"},
		{"1h", "1h0m0s"},
		{"", "1m0s"},      // Default
		{"invalid", "1m0s"}, // Default on error
	}

	for _, tc := range testCases {
		result := parseDuration(tc.input)
		if result.Duration.String() != tc.expected {
			t.Errorf("parseDuration(%q) = %s, want %s", tc.input, result.Duration.String(), tc.expected)
		}
	}
}

func TestConvertDirectory(t *testing.T) {
	// Create a temporary directory with test data
	tmpDir, err := os.MkdirTemp("", "converter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple test files
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
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", filename, err)
		}
	}

	converter := NewConverter(ConverterOptions{
		Namespace: "test-ns",
	})

	groups, err := converter.ConvertDirectory(tmpDir)
	if err != nil {
		t.Fatalf("conversion failed: %v", err)
	}

	if len(groups) != 2 {
		t.Errorf("expected 2 groups (one from each YAML file), got %d", len(groups))
	}
}

func TestConvertFileWithInvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "converter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid YAML
	inputYAML := `
groups:
- name: test
  invalid: [[[
`
	inputFile := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(inputFile, []byte(inputYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	converter := NewConverter(ConverterOptions{})

	_, err = converter.ConvertFile(inputFile)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestConvertFileWithInvalidPrometheusRules(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "converter-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create YAML with invalid Prometheus rule (missing 'expr')
	inputYAML := `
groups:
- name: test
  rules:
  - alert: MissingExpr
    labels:
      severity: warning
`
	inputFile := filepath.Join(tmpDir, "invalid-rule.yaml")
	if err := os.WriteFile(inputFile, []byte(inputYAML), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	converter := NewConverter(ConverterOptions{})

	_, err = converter.ConvertFile(inputFile)
	if err == nil {
		t.Error("expected error for invalid Prometheus rule, got nil")
	}
}

func TestConverterOptions(t *testing.T) {
	selector := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"grafana": "instance1",
		},
	}

	opts := ConverterOptions{
		Namespace: "monitoring",
		InstanceSelector: selector,
		AdditionalLabels: map[string]string{
			"team": "sre",
		},
		AdditionalAnnotations: map[string]string{
			"imported": "true",
		},
		FolderRef: "my-folder",
		FolderUID: "abc123",
		ResyncPeriod: "5m",
	}

	converter := NewConverter(opts)

	if converter.opts.Namespace != "monitoring" {
		t.Error("namespace not set correctly")
	}
	if converter.opts.InstanceSelector != selector {
		t.Error("instance selector not set correctly")
	}
	if converter.opts.AdditionalLabels["team"] != "sre" {
		t.Error("additional labels not set correctly")
	}
	if converter.opts.AdditionalAnnotations["imported"] != "true" {
		t.Error("additional annotations not set correctly")
	}
}
