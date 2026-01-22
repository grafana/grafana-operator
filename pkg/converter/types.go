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
	"github.com/prometheus/prometheus/model/rulefmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConverterOptions holds configuration for the conversion process
type ConverterOptions struct {
	// Namespace is the target Kubernetes namespace for the CR
	Namespace string

	// InstanceSelector is used to match Grafana instances
	InstanceSelector *metav1.LabelSelector

	// AdditionalLabels are extra labels to add to the CR metadata
	AdditionalLabels map[string]string

	// AdditionalAnnotations are extra annotations to add to the CR metadata
	AdditionalAnnotations map[string]string

	// FolderRef references the GrafanaFolder CR
	FolderRef string

	// FolderUID is the UID of the folder in Grafana
	FolderUID string

	// ResyncPeriod is how often to resync the alert rules
	ResyncPeriod string
}

// Result holds the conversion result
type Result struct {
	// Name is the name of the generated GrafanaAlertRuleGroup resource
	Name string

	// Namespace is the namespace of the generated resource
	Namespace string

	// RuleGroups is the converted GrafanaAlertRuleGroup
	RuleGroups []rulefmt.RuleGroup

	// Errors contains any validation errors encountered
	Errors []error
}

// PrometheusRuleGroup is an intermediate representation for parsing Prometheus rules
type PrometheusRuleGroup struct {
	Name     string         `yaml:"name" json:"name"`
	Interval string         `yaml:"interval,omitempty" json:"interval,omitempty"`
	Rules    []PrometheusRule `yaml:"rules" json:"rules"`
}

// PrometheusRule represents a single Prometheus alerting rule
type PrometheusRule struct {
	Alert       string            `yaml:"alert" json:"alert"`
	Expr        string            `yaml:"expr" json:"expr"`
	For         string            `yaml:"for,omitempty" json:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty" json:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty" json:"annotations,omitempty"`
	Record      string            `yaml:"record,omitempty" json:"record,omitempty"`
}
