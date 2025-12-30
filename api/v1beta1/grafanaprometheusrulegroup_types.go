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

// GrafanaPrometheusRuleGroupSpec defines the desired state of GrafanaPrometheusRuleGroup
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID)))", message="Only one of FolderUID or FolderRef can be set and one must be defined"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable)))", message="spec.editable is immutable"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderUID) && !has(self.folderUID)) || (has(oldSelf.folderUID) && has(self.folderUID)))", message="spec.folderUID is immutable"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderRef) && !has(self.folderRef)) || (has(oldSelf.folderRef) && has(self.folderRef)))", message="spec.folderRef is immutable"
type GrafanaPrometheusRuleGroupSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Name of the alert rule group. If not specified, the resource name will be used.
	// +optional
	Name string `json:"name,omitempty"`

	// UID of the folder containing this rule group
	// Overrides the FolderSelector
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderUID string `json:"folderUID,omitempty"`

	// Match GrafanaFolders CRs to infer the uid
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderRef string `json:"folderRef,omitempty"`

	// DatasourceUID is the UID of the Prometheus datasource in Grafana to use for queries.
	// This is required to convert PromQL expressions to Grafana alert queries.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	DatasourceUID string `json:"datasourceUID"`

	// Rules define the alerting and recording rules in Prometheus format.
	// Recording rules are ignored as they are not supported in Grafana alerting.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Rules []PrometheusRule `json:"rules"`

	// Interval is the time interval between evaluation of the rule group.
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:validation:Required
	Interval metav1.Duration `json:"interval"`

	// Whether to enable or disable editing of the alert rule group in Grafana UI
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	Editable *bool `json:"editable,omitempty"`
}

// PrometheusRule represents a Prometheus-style alerting or recording rule
type PrometheusRule struct {
	// Alert is the name of the alerting rule. Mutually exclusive with Record.
	// +optional
	Alert string `json:"alert,omitempty"`

	// Record is the name of the recording rule. Mutually exclusive with Alert.
	// Note: Recording rules are not supported in Grafana alerting and will be ignored.
	// +optional
	Record string `json:"record,omitempty"`

	// Expr is the PromQL expression to evaluate
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Expr string `json:"expr"`

	// For is the duration for which the condition must be true before firing
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(s|m|h|d|w))+$"
	// +optional
	For string `json:"for,omitempty"`

	// KeepFiringFor is the minimum duration an alert will continue firing after the condition clears
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(s|m|h|d|w))+$"
	// +optional
	KeepFiringFor string `json:"keep_firing_for,omitempty"`

	// Labels to add or overwrite for each alert
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to add to each alert
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaPrometheusRuleGroup is the Schema for the grafanaprometheusrulegroups API
// It allows defining alert rules using Prometheus-style syntax that will be converted
// to Grafana managed alerts.
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaPrometheusRuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaPrometheusRuleGroupSpec `json:"spec"`
	Status GrafanaCommonStatus            `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaPrometheusRuleGroup)(nil)

// GroupName returns the name of alert rule group.
func (in *GrafanaPrometheusRuleGroup) GroupName() string {
	groupName := in.Spec.Name
	if groupName == "" {
		groupName = in.Name
	}

	return groupName
}

// Conditions implements FolderReferencer.
func (in *GrafanaPrometheusRuleGroup) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaPrometheusRuleGroup) FolderNamespace() string {
	return in.Namespace
}

// FolderRef implements FolderReferencer.
func (in *GrafanaPrometheusRuleGroup) FolderRef() string {
	return in.Spec.FolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaPrometheusRuleGroup) FolderUID() string {
	return in.Spec.FolderUID
}

func (in *GrafanaPrometheusRuleGroup) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaPrometheusRuleGroup) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaPrometheusRuleGroup) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaPrometheusRuleGroup) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaPrometheusRuleGroup) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaPrometheusRuleGroup) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.GroupName())
}

var _ FolderReferencer = (*GrafanaPrometheusRuleGroup)(nil)

//+kubebuilder:object:root=true

// GrafanaPrometheusRuleGroupList contains a list of GrafanaPrometheusRuleGroup
type GrafanaPrometheusRuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaPrometheusRuleGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaPrometheusRuleGroup{}, &GrafanaPrometheusRuleGroupList{})
}
