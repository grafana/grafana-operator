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
	"time"

	"github.com/grafana-operator/grafana-operator/v5/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GrafanaDatasourceDataSource struct {
	ID     int64  `json:"id,omitempty"`
	UID    string `json:"uid,omitempty"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	Access string `json:"access"`

	Database string `json:"database,omitempty"`
	User     string `json:"user,omitempty"`

	OrgID     int64 `json:"orgId,omitempty"`
	IsDefault bool  `json:"isDefault"`

	BasicAuth     bool   `json:"basicAuth"`
	BasicAuthUser string `json:"basicAuthUser,omitempty"`

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
type GrafanaDatasourceSpec struct {
	DataSource GrafanaDatasourceDataSource `json:"datasource,omitempty"`

	// selects Grafana instances
	// +optional
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`

	// +optional
	ValuesFrom []GrafanaDatasourceValueFrom `json:"valuesFrom,omitempty"`

	// how often the datasource is refreshed
	Interval metav1.Duration `json:"interval"`

	// allow to import this resources from an operator in a different namespace
	// +optional
	AllowCrossNamespaceReferences *bool `json:"allowCrossNamespaceImport,omitempty"`
}

type GrafanaDatasourceValueFrom struct {
	TargetPath string                           `json:"targetPath"`
	ValueFrom  GrafanaDatasourceValueFromSource `json:"valueFrom"`
}

type GrafanaDatasourceValueFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

// GrafanaDatasourceStatus defines the observed state of GrafanaDatasource
type GrafanaDatasourceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Instances stores UID, version, and folder info for each instance the datasource has been created in
	// +optional
	Instances map[string]GrafanaDatasourceInstanceStatus `json:"instances,omitempty"`
}

type GrafanaDatasourceInstanceStatus struct {
	ID  int64  `json:"ID,omitempty"`
	UID string `json:"UID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// GrafanaDatasource is the Schema for the grafanadatasources API
type GrafanaDatasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDatasourceSpec   `json:"spec,omitempty"`
	Status GrafanaDatasourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaDatasourceList contains a list of GrafanaDatasource
type GrafanaDatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDatasource `json:"items"`
}

func (in *GrafanaDatasource) GetResyncPeriod() time.Duration {
	return in.Spec.Interval.Duration
}

func (in *GrafanaDatasource) IsAllowCrossNamespaceImport() bool {
	if in.Spec.AllowCrossNamespaceReferences != nil {
		return *in.Spec.AllowCrossNamespaceReferences
	}
	return false
}

func (in *GrafanaDatasource) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *GrafanaDatasource) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *GrafanaDatasource) GetReadyCondition() *metav1.Condition {
	return api.GetReadyCondition(in)
}

func (in *GrafanaDatasource) SetCondition(condition metav1.Condition) bool {
	return api.SetCondition(in, condition)
}

func (in *GrafanaDatasource) SetReadyCondition(status metav1.ConditionStatus, reason string, message string) bool {
	return api.SetReadyCondition(in, status, reason, message)
}

func (in *GrafanaDatasourceList) Find(namespace string, name string) *GrafanaDatasource {
	for _, datasource := range in.Items {
		if datasource.Namespace == namespace && datasource.Name == name {
			return &datasource
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&GrafanaDatasource{}, &GrafanaDatasourceList{})
}
