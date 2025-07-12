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
	operatorapi "github.com/grafana/grafana-operator/v5/api"
	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaAlertRuleGroupSpec defines the desired state of GrafanaAlertRuleGroup
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID)))", message="Only one of FolderUID or FolderRef can be set and one must be defined"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable)))", message="spec.editable is immutable"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderUID) && !has(self.folderUID)) || (has(oldSelf.folderUID) && has(self.folderUID)))", message="spec.folderUID is immutable"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.folderRef) && !has(self.folderRef)) || (has(oldSelf.folderRef) && has(self.folderRef)))", message="spec.folderRef is immutable"
type GrafanaAlertRuleGroupSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// +optional
	// Name of the alert rule group. If not specified, the resource name will be used.
	Name string `json:"name,omitempty"`

	// UID of the folder containing this rule group
	// Overrides the FolderSelector
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderUID string `json:"folderUID,omitempty"`

	// Match GrafanaFolders CRs to infer the uid
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	FolderRef string `json:"folderRef,omitempty"`

	// +kubebuilder:validation:MinItems=1
	Rules []AlertRule `json:"rules"`

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

	NotificationSettings *NotificationSettings `json:"notificationSettings,omitempty"`

	Labels map[string]string `json:"labels,omitempty"`

	// +kubebuilder:validation:Enum=Alerting;NoData;OK;KeepLast
	NoDataState *string `json:"noDataState"`

	Record *Record `json:"record,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=190
	// +kubebuilder:example="Always firing"
	Title string `json:"title"`

	// UID of the alert rule. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	UID string `json:"uid"`
}

type NotificationSettings struct {
	GroupBy           []string `json:"group_by,omitempty"`
	GroupInterval     string   `json:"group_interval,omitempty"`
	GroupWait         string   `json:"group_wait,omitempty"`
	Receiver          string   `json:"receiver"`
	MuteTimeIntervals []string `json:"mute_time_intervals,omitempty"`
	RepeatInterval    string   `json:"repeat_interval,omitempty"`
}

type Record struct {
	// +kubebuilder:validation:Required
	From string `json:"from"`

	// +kubebuilder:validation:Required
	Metric string `json:"metric"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaAlertRuleGroup is the Schema for the grafanaalertrulegroups API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaAlertRuleGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaAlertRuleGroupSpec `json:"spec"`
	Status GrafanaCommonStatus       `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaAlertRuleGroup)(nil)

// GroupName returns the name of alert rule group.
func (in *GrafanaAlertRuleGroup) GroupName() string {
	groupName := in.Spec.Name
	if groupName == "" {
		groupName = in.Name
	}
	return groupName
}

// CurrentGeneration implements FolderReferencer.
func (in *GrafanaAlertRuleGroup) CurrentGeneration() int64 {
	return in.Generation
}

// Conditions implements FolderReferencer.
func (in *GrafanaAlertRuleGroup) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaAlertRuleGroup) FolderNamespace() string {
	return in.Namespace
}

// FolderRef implements FolderReferencer.
func (in *GrafanaAlertRuleGroup) FolderRef() string {
	return in.Spec.FolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaAlertRuleGroup) FolderUID() string {
	return in.Spec.FolderUID
}

func (in *GrafanaAlertRuleGroup) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaAlertRuleGroup) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaAlertRuleGroup) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaAlertRuleGroup) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

func (in *GrafanaAlertRuleGroup) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.GroupName())
}

var _ operatorapi.FolderReferencer = (*GrafanaAlertRuleGroup)(nil)

//+kubebuilder:object:root=true

// GrafanaAlertRuleGroupList contains a list of GrafanaAlertRuleGroup
type GrafanaAlertRuleGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaAlertRuleGroup `json:"items"`
}

func (in *GrafanaAlertRuleGroupList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func init() {
	SchemeBuilder.Register(&GrafanaAlertRuleGroup{}, &GrafanaAlertRuleGroupList{})
}
