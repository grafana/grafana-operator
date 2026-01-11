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
	"encoding/json"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaDatasourceCorrelation defines a correlation from this datasource to a target datasource
type GrafanaDatasourceCorrelation struct {
	// UID of the target datasource
	// +kubebuilder:validation:Required
	TargetUID string `json:"targetUID"`

	// Label for the correlation, used as part of the correlation key
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Label string `json:"label"`

	// Optional description of the correlation
	// +optional
	Description string `json:"description,omitempty"`

	// Type of correlation: "query" or "external"
	// +kubebuilder:validation:Enum=query;external
	// +kubebuilder:default=query
	// +optional
	Type string `json:"type,omitempty"`

	// Config for the correlation
	// +optional
	Config *GrafanaDatasourceCorrelationConfig `json:"config,omitempty"`
}

// GrafanaDatasourceCorrelationConfig defines configuration for a correlation
type GrafanaDatasourceCorrelationConfig struct {
	// Field used to attach the correlation link
	// +kubebuilder:validation:Required
	Field string `json:"field"`

	// Target query configuration
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Required
	Target *apiextensionsv1.JSON `json:"target"`

	// Type of config: "query" or "external"
	// +kubebuilder:validation:Enum=query;external
	// +optional
	Type string `json:"type,omitempty"`

	// Transformations to apply to the source data
	// +optional
	Transformations []GrafanaDatasourceCorrelationTransformation `json:"transformations,omitempty"`
}

// GrafanaDatasourceCorrelationTransformation defines a transformation for correlation
type GrafanaDatasourceCorrelationTransformation struct {
	// Type of transformation: "regex" or "logfmt"
	// +kubebuilder:validation:Enum=regex;logfmt
	Type string `json:"type,omitempty"`

	// Field to transform
	// +optional
	Field string `json:"field,omitempty"`

	// Expression for regex transformations
	// +optional
	Expression string `json:"expression,omitempty"`

	// MapValue for logfmt transformations
	// +optional
	MapValue string `json:"mapValue,omitempty"`
}

type GrafanaDatasourceInternal struct {
	// Deprecated field, use spec.uid instead
	// +optional
	UID           string `json:"uid,omitempty"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	URL           string `json:"url,omitempty"`
	Access        string `json:"access,omitempty"`
	Database      string `json:"database,omitempty"`
	User          string `json:"user,omitempty"`
	IsDefault     *bool  `json:"isDefault,omitempty"`
	BasicAuth     *bool  `json:"basicAuth,omitempty"`
	BasicAuthUser string `json:"basicAuthUser,omitempty"`

	// Deprecated field, it has no effect
	OrgID *int64 `json:"orgId,omitempty"`

	// Whether to enable/disable editing of the datasource in Grafana UI
	// +optional
	Editable *bool `json:"editable,omitempty"`

	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// +optional
	JSONData json.RawMessage `json:"jsonData,omitempty"`

	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// +optional
	SecureJSONData json.RawMessage `json:"secureJsonData,omitempty"`
}

// GrafanaDatasourceSpec defines the desired state of GrafanaDatasource
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaDatasourceSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// The UID, for the datasource, fallback to the deprecated spec.datasource.uid
	// and metadata.uid. Can be any string consisting of alphanumeric characters,
	// - and _ with a maximum length of 40 +optional
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	CustomUID string `json:"uid,omitempty"`

	Datasource *GrafanaDatasourceInternal `json:"datasource"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`

	// environments variables from secrets or config maps
	// +optional
	// +kubebuilder:validation:MaxItems=99
	ValuesFrom []ValueFrom `json:"valuesFrom,omitempty"`

	// Correlations to create for this datasource
	// +optional
	Correlations []GrafanaDatasourceCorrelation `json:"correlations,omitempty"`
}

// GrafanaDatasourceStatus defines the observed state of GrafanaDatasource
type GrafanaDatasourceStatus struct {
	GrafanaCommonStatus `json:",inline"`

	Hash string `json:"hash,omitempty"`
	// Deprecated: Check status.conditions or operator logs
	LastMessage string `json:"lastMessage,omitempty"`
	// The datasource instanceSelector can't find matching grafana instances
	NoMatchingInstances bool   `json:"NoMatchingInstances,omitempty"`
	UID                 string `json:"uid,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaDatasource is the Schema for the grafanadatasources API
// +kubebuilder:printcolumn:name="No matching instances",type="boolean",JSONPath=".status.NoMatchingInstances",description=""
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaDatasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDatasourceSpec   `json:"spec"`
	Status GrafanaDatasourceStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaDatasource)(nil)

//+kubebuilder:object:root=true

// GrafanaDatasourceList contains a list of GrafanaDatasource
type GrafanaDatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDatasource `json:"items"`
}

func (in *GrafanaDatasource) Unchanged(hash string) bool {
	return in.Status.Hash == hash
}

func (in *GrafanaDatasource) IsUpdatedUID() bool {
	// Datasource has just been created, status is not yet updated
	if in.Status.UID == "" {
		return false
	}

	return in.Status.UID != in.GetGrafanaUID()
}

// GetGrafanaUID selects a UID to be used for Grafana API requests (preference: spec.CustomUID -> Spec.Datasource.UID -> metadata.uid)
func (in *GrafanaDatasource) GetGrafanaUID() string {
	if in.Spec.CustomUID != "" {
		return in.Spec.CustomUID
	}

	if in.Spec.Datasource.UID != "" {
		return in.Spec.Datasource.UID
	}

	return string(in.UID)
}

func (in *GrafanaDatasourceList) Exists(namespace, name string) bool {
	for _, item := range in.Items {
		if item.Namespace == namespace && item.Name == name {
			return true
		}
	}

	return false
}

func (in *GrafanaDatasource) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaDatasource) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaDatasource) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

func (in *GrafanaDatasource) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaDatasource) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

func (in *GrafanaDatasource) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GrafanaDatasource) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.GetGrafanaUID())
}

func (in *GrafanaDatasource) GetPluginConfigMapKey() string {
	return GetPluginConfigMapKey("datasource", &in.ObjectMeta)
}

func (in *GrafanaDatasource) GetPluginConfigMapDeprecatedKey() string {
	return fmt.Sprintf("%v-datasource", in.Name)
}

func init() {
	SchemeBuilder.Register(&GrafanaDatasource{}, &GrafanaDatasourceList{})
}
