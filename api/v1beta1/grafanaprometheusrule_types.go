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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaPrometheusRuleSpec defines the desired state of GrafanaPrometheusRule.
//
// The body mirrors a Prometheus rules file: a list of named groups, each with
// alerting and/or recording rules. The operator forwards the body to the
// rules.alerting.grafana.app/v0alpha1 PrometheusRule kind, which converts the
// rules server-side via the existing /api/convert/prometheus pipeline.
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || (!has(self.folderUID) && !(has(self.folderRef)))", message="Only one of FolderUID or FolderRef can be set"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderUID) && !has(self.folderUID)) || (has(oldSelf.folderUID) && has(self.folderUID)))", message="spec.folderUID is immutable"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderRef) && !has(self.folderRef)) || (has(oldSelf.folderRef) && has(self.folderRef)))", message="spec.folderRef is immutable"
type GrafanaPrometheusRuleSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// UID of the folder containing this resource. Overrides FolderRef. When
	// neither is set, the server auto-creates a default folder on first write.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderUID string `json:"folderUID,omitempty"`

	// Match GrafanaFolders CRs to infer the UID.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderRef string `json:"folderRef,omitempty"`

	// UID of the Prometheus datasource the rules query. Optional; defaults
	// server-side when omitted.
	// +optional
	DatasourceUID string `json:"datasourceUID,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Groups []PrometheusRuleGroup `json:"groups"`
}

// PrometheusRuleGroup mirrors the Prometheus rule-group shape.
type PrometheusRuleGroup struct {
	// Name of the rule group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Group-level evaluation interval.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// Group-level query offset.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	QueryOffset *metav1.Duration `json:"queryOffset,omitempty"`

	// Group-level rule limit.
	// +optional
	Limit *int64 `json:"limit,omitempty"`

	// Group-level labels merged into every rule's labels at conversion time.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Rules []PrometheusRule `json:"rules"`
}

// PrometheusRule is a single alerting or recording rule. Exactly one of Alert
// or Record must be set.
type PrometheusRule struct {
	// Name of the alert. Mutually exclusive with Record.
	// +optional
	Alert string `json:"alert,omitempty"`

	// Name of the time series to output. Mutually exclusive with Alert.
	// +optional
	Record string `json:"record,omitempty"`

	// PromQL expression to evaluate.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Expr string `json:"expr"`

	// Alerts are pending for For long before transitioning to firing.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	For *metav1.Duration `json:"for,omitempty"`

	// How long an alert continues firing after the condition stops being true.
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	KeepFiringFor *metav1.Duration `json:"keepFiringFor,omitempty"`

	// Labels to add or overwrite for each alert/sample.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to add to each alert.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaPrometheusRule is the Schema for the grafanaprometheusrules API.
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaPrometheusRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaPrometheusRuleSpec `json:"spec"`
	Status GrafanaCommonStatus       `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaPrometheusRule)(nil)

// Conditions implements FolderReferencer.
func (in *GrafanaPrometheusRule) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaPrometheusRule) FolderNamespace() string { return in.Namespace }

// FolderRef implements FolderReferencer.
func (in *GrafanaPrometheusRule) FolderRef() string { return in.Spec.FolderRef }

// FolderUID implements FolderReferencer.
func (in *GrafanaPrometheusRule) FolderUID() string { return in.Spec.FolderUID }

func (in *GrafanaPrometheusRule) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaPrometheusRule) MatchNamespace() string      { return in.Namespace }
func (in *GrafanaPrometheusRule) Metadata() metav1.ObjectMeta { return in.ObjectMeta }
func (in *GrafanaPrometheusRule) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaPrometheusRule) CommonStatus() *GrafanaCommonStatus { return &in.Status }

func (in *GrafanaPrometheusRule) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.Name)
}

var _ FolderReferencer = (*GrafanaPrometheusRule)(nil)

//+kubebuilder:object:root=true

// GrafanaPrometheusRuleList contains a list of GrafanaPrometheusRule.
type GrafanaPrometheusRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaPrometheusRule `json:"items"`
}

func (in *GrafanaPrometheusRuleList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}
