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
	"time"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DashboardSourceType string

const (
	DashboardSourceTypeRawJson    DashboardSourceType = "json"
	DashboardSourceTypeGzipJson   DashboardSourceType = "gzipJson"
	DashboardSourceJsonnetProject DashboardSourceType = "jsonnetProjectWithRuntimeRaw"
	DashboardSourceTypeUrl        DashboardSourceType = "url"
	DashboardSourceTypeJsonnet    DashboardSourceType = "jsonnet"
	DashboardSourceTypeGrafanaCom DashboardSourceType = "grafana"
	DashboardSourceConfigMap      DashboardSourceType = "configmap"
)

type GrafanaDashboardDatasource struct {
	InputName      string `json:"inputName"`
	DatasourceName string `json:"datasourceName"`
}

type GrafanaDashboardUrlBasicAuth struct {
	Username *v1.SecretKeySelector `json:"username,omitempty"`
	Password *v1.SecretKeySelector `json:"password,omitempty"`
}

type GrafanaDashboardUrlAuthorization struct {
	BasicAuth *GrafanaDashboardUrlBasicAuth `json:"basicAuth,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
// +kubebuilder:validation:XValidation:rule="(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID)))", message="Only one of folderUID or folderRef can be declared at the same time"
// +kubebuilder:validation:XValidation:rule="(has(self.folder) && !(has(self.folderRef) || has(self.folderUID))) || !(has(self.folder))", message="folder field cannot be set when folderUID or folderRef is already declared"
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid)))", message="spec.uid is immutable"
type GrafanaDashboardSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Manually specify the uid for the dashboard, overwrites uids already present
	// in the json model. Can be any string consisting of alphanumeric characters,
	// - and _ with a maximum length of 40
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	CustomUID string `json:"uid,omitempty"`

	// dashboard json
	// +optional
	Json string `json:"json,omitempty"`

	// GzipJson the dashboard's JSON compressed with Gzip. Base64-encoded when in YAML.
	// +optional
	GzipJson []byte `json:"gzipJson,omitempty"`

	// dashboard url
	// +optional
	Url string `json:"url,omitempty"`

	// authorization options for dashboard from url
	// +optional
	UrlAuthorization *GrafanaDashboardUrlAuthorization `json:"urlAuthorization,omitempty"`

	// Jsonnet
	// +optional
	Jsonnet string `json:"jsonnet,omitempty"`

	// Jsonnet project build
	JsonnetProjectBuild *JsonnetProjectBuild `json:"jsonnetLib,omitempty"`

	// grafana.com/dashboards
	// +optional
	GrafanaCom *GrafanaComDashboardReference `json:"grafanaCom,omitempty"`

	// dashboard from configmap
	// +optional
	ConfigMapRef *v1.ConfigMapKeySelector `json:"configMapRef,omitempty"`

	// folder assignment for dashboard
	// +optional
	FolderTitle string `json:"folder,omitempty"`

	// UID of the target folder for this dashboard
	// +optional
	FolderUID string `json:"folderUID,omitempty"`

	// Name of a `GrafanaFolder` resource in the same namespace
	// +optional
	FolderRef string `json:"folderRef,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`

	// Cache duration for dashboards fetched from URLs
	// +optional
	ContentCacheDuration metav1.Duration `json:"contentCacheDuration,omitempty"`

	// maps required data sources to existing ones
	// +optional
	Datasources []GrafanaDashboardDatasource `json:"datasources,omitempty"`

	// environments variables as a map
	// +optional
	Envs []GrafanaDashboardEnv `json:"envs,omitempty"`

	// environments variables from secrets or config maps
	// +optional
	EnvsFrom []GrafanaDashboardEnvFromSource `json:"envFrom,omitempty"`
}

type GrafanaDashboardEnv struct {
	Name string `json:"name"`
	// Inline env value
	// +optional
	Value string `json:"value,omitempty"`
	// Reference on value source, might be the reference on a secret or config map
	// +optional
	ValueFrom GrafanaDashboardEnvFromSource `json:"valueFrom,omitempty"`
}

type GrafanaDashboardEnvFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *v1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *v1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type JsonnetProjectBuild struct {
	JPath              []string `json:"jPath,omitempty"`
	FileName           string   `json:"fileName"`
	GzipJsonnetProject []byte   `json:"gzipJsonnetProject"`
}

// GrafanaComDashbooardReference is a reference to a dashboard on grafana.com/dashboards
type GrafanaComDashboardReference struct {
	Id       int  `json:"id"`
	Revision *int `json:"revision,omitempty"`
}

// GrafanaDashboardStatus defines the observed state of GrafanaDashboard
type GrafanaDashboardStatus struct {
	GrafanaCommonStatus `json:",inline"`

	ContentCache     []byte      `json:"contentCache,omitempty"`
	ContentTimestamp metav1.Time `json:"contentTimestamp,omitempty"`
	ContentUrl       string      `json:"contentUrl,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	// The dashboard instanceSelector can't find matching grafana instances
	NoMatchingInstances bool   `json:"NoMatchingInstances,omitempty"`
	UID                 string `json:"uid,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaDashboard is the Schema for the grafanadashboards API
// +kubebuilder:printcolumn:name="No matching instances",type="boolean",JSONPath=".status.NoMatchingInstances",description=""
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
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

// FolderRef implements FolderReferencer.
func (in *GrafanaDashboard) FolderRef() string {
	return in.Spec.FolderRef
}

// FolderUID implements FolderReferencer.
func (in *GrafanaDashboard) FolderUID() string {
	return in.Spec.FolderUID
}

// Wrapper around CustomUID, dashboardModelUID or default metadata.uid
func (in *GrafanaDashboard) CustomUIDOrUID(dashboardUID string) string {
	if in.Spec.CustomUID != "" {
		return in.Spec.CustomUID
	}

	if dashboardUID != "" {
		return dashboardUID
	}

	return string(in.ObjectMeta.UID)
}

// FolderNamespace implements FolderReferencer.
func (in *GrafanaDashboard) FolderNamespace() string {
	return in.Namespace
}

// Conditions implements FolderReferencer.
func (in *GrafanaDashboard) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// CurrentGeneration implements FolderReferencer.
func (in *GrafanaDashboard) CurrentGeneration() int64 {
	return in.Generation
}

func (in *GrafanaDashboard) Unchanged(hash string) bool {
	return in.Status.Hash == hash
}

func (in *GrafanaDashboard) ResyncPeriodHasElapsed() bool {
	deadline := in.Status.LastResync.Add(in.Spec.ResyncPeriod.Duration)
	return time.Now().After(deadline)
}

func (in *GrafanaDashboard) GetSourceTypes() []DashboardSourceType {
	var sourceTypes []DashboardSourceType

	if in.Spec.Json != "" {
		sourceTypes = append(sourceTypes, DashboardSourceTypeRawJson)
	}

	if in.Spec.GzipJson != nil {
		sourceTypes = append(sourceTypes, DashboardSourceTypeGzipJson)
	}

	if in.Spec.Url != "" {
		sourceTypes = append(sourceTypes, DashboardSourceTypeUrl)
	}

	if in.Spec.Jsonnet != "" {
		sourceTypes = append(sourceTypes, DashboardSourceTypeJsonnet)
	}

	if in.Spec.GrafanaCom != nil {
		sourceTypes = append(sourceTypes, DashboardSourceTypeGrafanaCom)
	}

	if in.Spec.ConfigMapRef != nil {
		sourceTypes = append(sourceTypes, DashboardSourceConfigMap)
	}

	if in.Spec.JsonnetProjectBuild != nil {
		sourceTypes = append(sourceTypes, DashboardSourceJsonnetProject)
	}

	return sourceTypes
}

func (in *GrafanaDashboard) GetContentCache() []byte {
	return in.Status.getContentCache(in.Spec.Url, in.Spec.ContentCacheDuration.Duration)
}

// getContentCache returns content cache when the following conditions are met: url is the same, data is not expired, gzipped data is not corrupted
func (in *GrafanaDashboardStatus) getContentCache(url string, cacheDuration time.Duration) []byte {
	if in.ContentUrl != url {
		return []byte{}
	}

	notExpired := cacheDuration <= 0 || in.ContentTimestamp.Add(cacheDuration).After(time.Now())
	if !notExpired {
		return []byte{}
	}

	cache, err := Gunzip(in.ContentCache)
	if err != nil {
		return []byte{}
	}

	return cache
}

func (in *GrafanaDashboard) IsAllowCrossNamespaceImport() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaDashboard) IsUpdatedUID(uid string) bool {
	// Dashboard has just been created, status is not yet updated
	if in.Status.UID == "" {
		return false
	}

	uid = in.CustomUIDOrUID(uid)

	return in.Status.UID != uid
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
