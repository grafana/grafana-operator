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
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/rulefmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Converter handles the conversion of Prometheus alert rules to GrafanaAlertRuleGroup CRs
type Converter struct {
	opts ConverterOptions
}

// NewConverter creates a new Converter with the given options
func NewConverter(opts ConverterOptions) *Converter {
	return &Converter{opts: opts}
}

// ConvertFile converts a single Prometheus rules file to GrafanaAlertRuleGroup CRs
func (c *Converter) ConvertFile(inputPath string) ([]v1beta1.GrafanaAlertRuleGroup, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("reading input file: %w", err)
	}

	// Parse using Prometheus's own parser for better validation
	ruleGroups, errs := rulefmt.Parse(data, false, model.LegacyValidation)
	if len(errs) > 0 {
		return nil, fmt.Errorf("parsing Prometheus rules: %v", errs)
	}

	return c.convertRuleGroups(ruleGroups.Groups, filepath.Base(inputPath)), nil
}

// ConvertDirectory converts all Prometheus rule files in a directory
func (c *Converter) ConvertDirectory(inputDir string) ([]v1beta1.GrafanaAlertRuleGroup, error) {
	var allGroups []v1beta1.GrafanaAlertRuleGroup

	entries, err := os.ReadDir(inputDir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip non-YAML files
		if !strings.HasSuffix(entry.Name(), ".yaml") && !strings.HasSuffix(entry.Name(), ".yml") {
			continue
		}

		path := filepath.Join(inputDir, entry.Name())

		groups, err := c.ConvertFile(path)
		if err != nil {
			return nil, fmt.Errorf("converting %s: %w", entry.Name(), err)
		}

		allGroups = append(allGroups, groups...)
	}

	return allGroups, nil
}

// convertRuleGroups converts Prometheus rule groups to GrafanaAlertRuleGroup CRs
func (c *Converter) convertRuleGroups(groups []rulefmt.RuleGroup, sourceFile string) []v1beta1.GrafanaAlertRuleGroup {
	result := make([]v1beta1.GrafanaAlertRuleGroup, 0, len(groups))

	for _, group := range groups {
		cr := c.convertRuleGroup(group, sourceFile)
		result = append(result, cr)
	}

	return result
}

// convertRuleGroup converts a single Prometheus rule group to a GrafanaAlertRuleGroup CR
func (c *Converter) convertRuleGroup(group rulefmt.RuleGroup, sourceFile string) v1beta1.GrafanaAlertRuleGroup {
	// Generate a unique name for the CR
	name := sanitizeName(group.Name)

	// Create labels merging additional labels with default ones
	labels := make(map[string]string)
	if c.opts.AdditionalLabels != nil {
		maps.Copy(labels, c.opts.AdditionalLabels)
	}

	labels["source"] = sourceFile

	// Create annotations merging additional annotations with default ones
	annotations := make(map[string]string)
	if c.opts.AdditionalAnnotations != nil {
		maps.Copy(annotations, c.opts.AdditionalAnnotations)
	}

	annotations["original-file"] = sourceFile

	// Convert rules
	rules := make([]v1beta1.AlertRule, 0, len(group.Rules))
	for _, rule := range group.Rules {
		alertRule := c.convertRule(rule)
		rules = append(rules, alertRule)
	}

	// Parse interval
	interval := parseDurationFromModel(group.Interval)

	// Default folderUID if not provided
	folderUID := c.opts.FolderUID
	if folderUID == "" {
		folderUID = "prometheus-alerts"
	}

	return v1beta1.GrafanaAlertRuleGroup{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "grafana.integreatly.org/v1beta1",
			Kind:       "GrafanaAlertRuleGroup",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   c.opts.Namespace,
			Labels:      labels,
			Annotations: annotations,
		},
		Spec: v1beta1.GrafanaAlertRuleGroupSpec{
			GrafanaCommonSpec: v1beta1.GrafanaCommonSpec{
				InstanceSelector: c.opts.InstanceSelector,
			},
			Name:      group.Name,
			FolderUID: folderUID,
			Rules:     rules,
			Interval:  interval,
		},
	}
}

// convertRule converts a single Prometheus rule to a Grafana AlertRule
func (c *Converter) convertRule(rule rulefmt.Rule) v1beta1.AlertRule {
	noDataState := "NoData"
	alertRule := v1beta1.AlertRule{
		Title:        rule.Alert,
		Condition:    rule.Expr,
		ExecErrState: "Error",
		NoDataState:  &noDataState,
		UID:          generateUID(rule.Alert),
		Data:         []*v1beta1.AlertQuery{},
		Labels:       make(map[string]string),
		Annotations:  make(map[string]string),
	}

	// Add for duration if specified (model.Duration is a duration string)
	if rule.For > 0 {
		forStr := rule.For.String()
		alertRule.For = &forStr
	}

	// Add labels (merge with any from the rule)
	if rule.Labels != nil {
		maps.Copy(alertRule.Labels, rule.Labels)
	}

	// Add annotations
	if rule.Annotations != nil {
		maps.Copy(alertRule.Annotations, rule.Annotations)
	}

	return alertRule
}

// sanitizeName converts a name to be DNS-1123 compliant
func sanitizeName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace underscores, dots and spaces with hyphens
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, " ", "-")

	// Remove any characters that are not alphanumeric or hyphens
	var result strings.Builder

	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result.WriteRune(c)
		}
	}

	// Remove leading/trailing hyphens
	name = strings.Trim(result.String(), "-")

	// If empty after sanitization, use a hash of the original name
	if name == "" {
		name = fmt.Sprintf("alert-%x", time.Now().UnixNano())
	}

	return name
}

// parseDurationFromModel parses a model.Duration to metav1.Duration
// model.Duration is int64 representing nanoseconds
func parseDurationFromModel(d model.Duration) metav1.Duration {
	if d == 0 {
		d = model.Duration(time.Minute) // Default to 1 minute
	}

	return metav1.Duration{Duration: time.Duration(d)}
}

// parseDuration parses a duration string to metav1.Duration
func parseDuration(durationStr string) metav1.Duration {
	if durationStr == "" {
		durationStr = "1m" // Default to 1 minute
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		duration = time.Minute // Default to 1 minute on error
	}

	return metav1.Duration{Duration: duration}
}

// generateUID generates a UID for an alert rule from its name
// UIDs must be alphanumeric, dash, or underscore, max 40 chars
func generateUID(name string) string {
	// Start with sanitized name
	uid := sanitizeName(name)

	// Replace hyphens with underscores for UID
	uid = strings.ReplaceAll(uid, "-", "_")

	// Truncate to 40 characters
	if len(uid) > 40 {
		uid = uid[:40]
	}

	return uid
}
