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
	"github.com/grafana/grafana-openapi-client-go/models"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaAlertRuleGroupSpec defines the desired state of GrafanaAlertRuleGroup
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID)))", message="Only one of FolderUID or FolderRef can be set"
type GrafanaAlertRuleGroupSpec struct {
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:default="10m"
	ResyncPeriod metav1.Duration `json:"resyncPeriod,omitempty"`

	// selects Grafanas for import
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector"`

	// UID of the folder containing this rule group
	// Overrides the FolderSelector
	FolderUID string `json:"folderUID,omitempty"`

	// Match GrafanaFolders CRs to infer the uid
	FolderRef string `json:"folderRef,omitempty"`

	Rules []AlertRule `json:"rules"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:validation:Required
	Interval metav1.Duration `json:"interval"`

	// +optional
	AllowCrossNamespaceImport *bool `json:"allowCrossNamespaceImport,omitempty"`
}

// AlertRule defines a specific rule to be evaluated. It is based on the upstream model with some k8s specific type mappings
type AlertRule struct {
	Annotations map[string]string `json:"annotations,omitempty"`

	Condition string `json:"condition"`

	// +kubebuilder:validation:Required
	Data []*AlertQuery `json:"data"`

	// +kubebuilder:validation:Enum=OK;Alerting;Error;KeepLast
	ExecErrState string `json:"execErrState"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:validation:Required
	For *metav1.Duration `json:"for"`

	IsPaused bool `json:"isPaused,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	// +kubebuilder:validation:Enum=Alerting;NoData;OK;KeepLast
	NoDataState *string `json:"noDataState"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=190
	// +kubebuilder:example="Always firing"
	Title string `json:"title"`

	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	UID string `json:"uid"`
}

type AlertQuery struct {
	// Grafana data source unique identifier; it should be '__expr__' for a Server Side Expression operation.
	DatasourceUID string `json:"datasourceUid,omitempty"`

	// JSON is the raw JSON query and includes the above properties as well as custom properties.
	Model *apiextensions.JSON `json:"model,omitempty"`

	// QueryType is an optional identifier for the type of query.
	// It can be used to distinguish different types of queries.
	QueryType string `json:"queryType,omitempty"`

	// RefID is the unique identifier of the query, set by the frontend call.
	RefID string `json:"refId,omitempty"`

	// relative time range
	RelativeTimeRange *models.RelativeTimeRange `json:"relativeTimeRange,omitempty"`
}

// GrafanaAlertRuleGroupStatus defines the observed state of GrafanaAlertRuleGroup
type GrafanaAlertRuleGroupStatus struct {
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaAlertRuleGroup is the Schema for the grafanaalertrulegroups API
type GrafanaAlertRuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaAlertRuleGroupSpec   `json:"spec,omitempty"`
	Status GrafanaAlertRuleGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaAlertRuleGroupList contains a list of GrafanaAlertRuleGroup
type GrafanaAlertRuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaAlertRuleGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaAlertRuleGroup{}, &GrafanaAlertRuleGroupList{})
}
