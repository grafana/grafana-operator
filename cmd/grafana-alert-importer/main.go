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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/converter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var cli struct {
	Input            string   `arg:"" type:"path" help:"Path to Prometheus rule file or directory containing rule files"`
	Output           string   `short:"o" type:"path" help:"Output file path (default: stdout)"`
	Namespace        string   `short:"n" default:"default" help:"Target Kubernetes namespace for the CRs"`
	Labels           []string `short:"l" help:"Extra labels to add to the CR metadata (e.g., 'team=sre')"`
	Annotations      []string `short:"a" help:"Extra annotations to add to the CR metadata (e.g., 'imported=true')"`
	InstanceSelector string   `help:"Label selector to match Grafana instances (e.g., 'grafana=main')"`
	ResyncPeriod     string   `default:"5m" help:"Resync period for alert rules"`
	FolderUID        string   `default:"prometheus-alerts" help:"UID of the folder in Grafana to store alerts"`
}

func main() {
	ctx := kong.Parse(&cli,
		kong.Name("grafana-alert-importer"),
		kong.Description("Import Prometheus alert rules and convert them to GrafanaAlertRuleGroup CRs"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
		}),
		kong.Vars{
			"version": fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		},
	)

	if err := run(); err != nil {
		ctx.Errorf("%v", err)
		os.Exit(1)
	}
}

func run() error {
	if cli.Input == "" {
		return fmt.Errorf("input path is required")
	}

	// Parse labels
	additionalLabels := make(map[string]string)

	for _, label := range cli.Labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format: %s (expected key=value)", label)
		}

		additionalLabels[parts[0]] = parts[1]
	}

	// Parse annotations
	additionalAnnotations := make(map[string]string)

	for _, annotation := range cli.Annotations {
		parts := strings.SplitN(annotation, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid annotation format: %s (expected key=value)", annotation)
		}

		additionalAnnotations[parts[0]] = parts[1]
	}

	// Parse instance selector if provided
	var instanceSelector *metav1.LabelSelector

	if cli.InstanceSelector != "" {
		parts := strings.SplitN(cli.InstanceSelector, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid instance-selector format: %s (expected key=value)", cli.InstanceSelector)
		}

		instanceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				parts[0]: parts[1],
			},
		}
	}

	opts := converter.ConverterOptions{
		Namespace:             cli.Namespace,
		InstanceSelector:      instanceSelector,
		AdditionalLabels:      additionalLabels,
		AdditionalAnnotations: additionalAnnotations,
		FolderUID:             cli.FolderUID,
		ResyncPeriod:          cli.ResyncPeriod,
	}

	conv := converter.NewConverter(opts)

	var groups []v1beta1.GrafanaAlertRuleGroup

	var err error

	info, err := os.Stat(cli.Input)
	if err != nil {
		return fmt.Errorf("stat input path: %w", err)
	}

	if info.IsDir() {
		groups, err = conv.ConvertDirectory(cli.Input)
	} else {
		groups, err = conv.ConvertFile(cli.Input)
	}

	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Generate output
	output, err := generateYAML(groups)
	if err != nil {
		return fmt.Errorf("generating YAML: %w", err)
	}

	if cli.Output == "" {
		// Write to stdout
		_, err := os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	} else {
		// Write to file
		if err := os.WriteFile(cli.Output, output, 0o600); err != nil {
			return fmt.Errorf("writing to output file: %w", err)
		}

		fmt.Fprintf(os.Stdout, "Successfully converted %d rule group(s) to %s\n", len(groups), cli.Output)
	}

	return nil
}

func generateYAML(groups []v1beta1.GrafanaAlertRuleGroup) ([]byte, error) {
	var sb strings.Builder

	for i, group := range groups {
		yamlBytes, err := yaml.Marshal(group)
		if err != nil {
			return nil, fmt.Errorf("marshaling group %d: %w", i, err)
		}

		sb.Write(yamlBytes)

		if i < len(groups)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return []byte(sb.String()), nil
}
