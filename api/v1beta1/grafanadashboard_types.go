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
	"bytes"
	"compress/gzip"
	"io"
	"math"
	"time"

	"github.com/grafana-operator/grafana-operator/v5/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GrafanaDashboardDatasource struct {
	InputName     string             `json:"inputName"`
	DatasourceRef v1.ObjectReference `json:"datasourceRef"`
}

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	Source GrafanaDashboardSource `json:"source"`

	// selects Grafanas for import
	// +optional
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`

	// folder assignment for dashboard
	// +optional
	FolderTitle string `json:"folder,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`

	// how often the dashboard is endured to exist on the selected instances, defaults to 24h if not set
	Interval metav1.Duration `json:"interval"`

	// maps required data sources to existing ones
	// +optional
	Datasources []GrafanaDashboardDatasource `json:"datasources,omitempty"`

	// allow to import this resources from an operator in a different namespace
	// +optional
	AllowCrossNamespaceReferences *bool `json:"allowCrossNamespaceReferences,omitempty"`
}

// +kubebuilder:validation:MinProperties:=1
// +kubebuilder:validation:MaxProperties:=1
type GrafanaDashboardSource struct {
	Inline *GrafanaDashboardInlineSource `json:"inline,omitempty"`

	Remote *GrafanaDashboardRemoteSource `json:"remote,omitempty"`

	ConfigMap *v1.ConfigMapKeySelector `json:"configMap,omitempty"`
}

// +kubebuilder:validation:MinProperties:=1
// +kubebuilder:validation:MaxProperties:=1
type GrafanaDashboardInlineSource struct {
	// dashboard json
	// +optional
	Json *string `json:"json,omitempty"`

	// GzipJson the dashboard's JSON compressed with Gzip. Base64-encoded when in YAML.
	// +optional
	GzipJson []byte `json:"gzipJson,omitempty"`

	// Jsonnet
	// +optional
	Jsonnet *string `json:"jsonnet,omitempty"`
}

// +kubebuilder:validation:MinProperties:=2
// +kubebuilder:validation:MaxProperties:=2
type GrafanaDashboardRemoteSource struct {
	// Cache duration for dashboards fetched from URLs
	ContentCacheDuration metav1.Duration `json:"contentCacheDuration"`

	// dashboard url
	// +optional
	Url *string `json:"url,omitempty"`

	// grafana.com/dashboards
	// +optional
	GrafanaCom *GrafanaComDashboardReference `json:"grafanaCom,omitempty"`
}

// GrafanaComDashbooardReference is a reference to a dashboard on grafana.com/dashboards
type GrafanaComDashboardReference struct {
	Id int `json:"id"`

	// +optional
	Revision *int `json:"revision,omitempty"`
}

// GrafanaDashboardStatus defines the observed state of GrafanaDashboard
type GrafanaDashboardStatus struct {
	// Content contains information about fetched remote content
	// +optional
	Content      *GrafanaDashboardStatusContent      `json:"content,omitempty"`
	ContentError *GrafanaDashboardStatusContentError `json:"contentError,omitempty"`

	// Instances stores UID, version, and folder info for each instance the dashboard has been created in
	// +optional
	Instances map[string]GrafanaDashboardInstanceStatus `json:"instances,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type GrafanaDashboardStatusContent struct {
	Cache     []byte      `json:"cache,omitempty"`
	Timestamp metav1.Time `json:"timestamp,omitempty"`
	Url       string      `json:"url"`
}

type GrafanaDashboardStatusContentError struct {
	Message   string      `json:"message"`
	Timestamp metav1.Time `json:"timestamp"`
	Attempts  int         `json:"attempts"`
}

type GrafanaDashboardInstanceStatus struct {
	Version int64  `json:"Version,omitempty"`
	UID     string `json:"UID,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName=dash
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

// GrafanaDashboard is the Schema for the grafanadashboards API
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDashboardSpec   `json:"spec,omitempty"`
	Status GrafanaDashboardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaDashboardList contains a list of GrafanaDashboard
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDashboard `json:"items"`
}

func (in *GrafanaDashboard) GetResyncPeriod() time.Duration {
	return in.Spec.Interval.Duration
}

func (in *GrafanaDashboard) GetContentCache(url string) []byte {
	if in.Status.Content == nil {
		return nil
	}
	if in.Status.Content.Url != url {
		in.Status.Content = nil
		return nil
	}

	// expect that this is only called when a remote source actually exists
	cacheDuration := in.Spec.Source.Remote.ContentCacheDuration.Duration

	// TODO: is this boolean logic correct?
	expired := cacheDuration == 0 || time.Now().After(in.Status.Content.Timestamp.Add(cacheDuration))
	if expired {
		return nil
	}

	cache, err := Gunzip(in.Status.Content.Cache)
	if err != nil {
		in.Status.Content = nil
		return nil
	}

	return cache
}

func (in *GrafanaDashboard) SetStatusContentError(err error) {
	if err == nil {
		in.Status.ContentError = nil
		return
	}

	if in.Status.Content == nil {
		in.Status.Content = &GrafanaDashboardStatusContent{}
	}
	if in.Status.ContentError == nil {
		in.Status.ContentError = &GrafanaDashboardStatusContentError{}
	}

	in.Status.ContentError = &GrafanaDashboardStatusContentError{
		Message:   err.Error(),
		Timestamp: metav1.Now(),
		Attempts:  in.Status.ContentError.Attempts + 1,
	}
}

func (in *GrafanaDashboard) ContentErrorBackoff() *api.BackoffError {
	if in.Status.Content != nil && in.Status.ContentError != nil {
		e := in.Status.ContentError
		backoffDuration := time.Second * time.Duration(math.Exp(0.5*float64(e.Attempts)))
		retryTime := e.Timestamp.Add(backoffDuration)
		if retryTime.After(time.Now()) {
			return &api.BackoffError{Attempts: e.Attempts, RetryTime: retryTime}
		}
	}
	return nil
}

func (in *GrafanaDashboard) IsAllowCrossNamespaceImport() bool {
	if in.Spec.AllowCrossNamespaceReferences != nil {
		return *in.Spec.AllowCrossNamespaceReferences
	}
	return false
}

func (in *GrafanaDashboard) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *GrafanaDashboard) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *GrafanaDashboard) GetReadyCondition() *metav1.Condition {
	return api.GetReadyCondition(in)
}

func (in *GrafanaDashboard) SetCondition(condition metav1.Condition) bool {
	return api.SetCondition(in, condition)
}

func (in *GrafanaDashboard) UnSetCondition(conditionType string) {
	api.UnSetCondition(in, conditionType)
}

func (in *GrafanaDashboard) SetReadyCondition(status metav1.ConditionStatus, reason string, message string) bool {
	return api.SetReadyCondition(in, status, reason, message)
}

func Gunzip(compressed []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(gz)
}

func Gzip(content []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(buf)

	_, err := gz.Write(content)
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return io.ReadAll(buf)
}

func (in *GrafanaDashboardList) Find(namespace string, name string) *GrafanaDashboard {
	for _, dashboard := range in.Items {
		if dashboard.Namespace == namespace && dashboard.Name == name {
			return &dashboard
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}
