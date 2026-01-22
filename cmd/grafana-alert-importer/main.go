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

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/pkg/converter"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var (
	inputDir     string
	outputFile   string
	namespace    string
	labels       []string
	annotations  []string
	instanceRef  string
	resyncPeriod string
	folderUID    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "grafana-alert-importer",
		Short:   "Import Prometheus alert rules and convert them to GrafanaAlertRuleGroup CRs",
		Long:    `A tool that converts Prometheus alert rule files to GrafanaAlertRuleGroup Custom Resources for use with the Grafana Operator.`,
		Version: fmt.Sprintf("%s (commit: %s, date: %s)", version, commit, date),
		RunE:    runImport,
	}

	rootCmd.Flags().StringVarP(&inputDir, "input", "i", "", "Path to Prometheus rule file or directory containing rule files (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Target Kubernetes namespace for the CRs")
	rootCmd.Flags().StringArrayVarP(&labels, "label", "l", []string{}, "Extra labels to add to the CR metadata (e.g., 'team=sre')")
	rootCmd.Flags().StringArrayVarP(&annotations, "annotation", "a", []string{}, "Extra annotations to add to the CR metadata (e.g., 'imported=true')")
	rootCmd.Flags().StringVar(&instanceRef, "instance-selector", "", "Label selector to match Grafana instances (e.g., 'grafana=main')")
	rootCmd.Flags().StringVar(&resyncPeriod, "resync-period", "5m", "Resync period for alert rules")
	rootCmd.Flags().StringVar(&folderUID, "folder-uid", "prometheus-alerts", "UID of the folder in Grafana to store alerts")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runImport(cmd *cobra.Command, args []string) error {
	if inputDir == "" {
		return fmt.Errorf("--input/-i is required")
	}

	// Parse labels
	additionalLabels := make(map[string]string)
	for _, label := range labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid label format: %s (expected key=value)", label)
		}
		additionalLabels[parts[0]] = parts[1]
	}

	// Parse annotations
	additionalAnnotations := make(map[string]string)
	for _, annotation := range annotations {
		parts := strings.SplitN(annotation, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid annotation format: %s (expected key=value)", annotation)
		}
		additionalAnnotations[parts[0]] = parts[1]
	}

	// Parse instance selector if provided
	var instanceSelector *metav1.LabelSelector
	if instanceRef != "" {
		parts := strings.SplitN(instanceRef, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid instance-selector format: %s (expected key=value)", instanceRef)
		}
		instanceSelector = &metav1.LabelSelector{
			MatchLabels: map[string]string{
				parts[0]: parts[1],
			},
		}
	}

	opts := converter.ConverterOptions{
		Namespace:             namespace,
		InstanceSelector:      instanceSelector,
		AdditionalLabels:      additionalLabels,
		AdditionalAnnotations: additionalAnnotations,
		FolderUID:             folderUID,
		ResyncPeriod:          resyncPeriod,
	}

	conv := converter.NewConverter(opts)

	var groups []v1beta1.GrafanaAlertRuleGroup
	var err error

	info, err := os.Stat(inputDir)
	if err != nil {
		return fmt.Errorf("stat input path: %w", err)
	}

	if info.IsDir() {
		groups, err = conv.ConvertDirectory(inputDir)
	} else {
		groups, err = conv.ConvertFile(inputDir)
	}

	if err != nil {
		return fmt.Errorf("conversion failed: %w", err)
	}

	// Generate output
	output, err := generateYAML(groups)
	if err != nil {
		return fmt.Errorf("generating YAML: %w", err)
	}

	if outputFile == "" {
		// Write to stdout
		_, err := os.Stdout.Write(output)
		if err != nil {
			return fmt.Errorf("writing to stdout: %w", err)
		}
	} else {
		// Write to file
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			return fmt.Errorf("writing to output file: %w", err)
		}
		fmt.Printf("Successfully converted %d rule group(s) to %s\n", len(groups), outputFile)
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
