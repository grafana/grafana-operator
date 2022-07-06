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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type OperatorStageName string

type OperatorStageStatus string

const (
	OperatorStageGrafanaConfig  OperatorStageName = "config"
	OperatorStageAdminUser      OperatorStageName = "admin user"
	OperatorStagePvc            OperatorStageName = "pvc"
	OperatorStageServiceAccount OperatorStageName = "service account"
	OperatorStageService        OperatorStageName = "service"
	OperatorStageIngress        OperatorStageName = "ingress"
	OperatorStagePlugins        OperatorStageName = "plugins"
	OperatorStageDeployment     OperatorStageName = "deployment"
)

const (
	OperatorStageResultSuccess    OperatorStageStatus = "success"
	OperatorStageResultFailed     OperatorStageStatus = "failed"
	OperatorStageResultInProgress OperatorStageStatus = "in progress"
)

// temporary values passed between reconciler stages
type OperatorReconcileVars struct {
	// used to restart the Grafana container when the config changes
	ConfigHash string

	// env var value for installed plugins
	Plugins string
}

// GrafanaSpec defines the desired state of Grafana
type GrafanaSpec struct {
	Config                GrafanaConfig            `json:"config"`
	Containers            []v1.Container           `json:"containers,omitempty"`
	Ingress               *IngressNetworkingV1     `json:"ingress,omitempty"`
	Route                 *RouteOpenshiftV1        `json:"route,omitempty"`
	Service               *ServiceV1               `json:"service,omitempty"`
	Deployment            *DeploymentV1            `json:"deployment,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimV1 `json:"persistentVolumeClaim,omitempty"`
	ServiceAccount        *ServiceAccountV1        `json:"serviceAccount,omitempty"`
	Client                *GrafanaClient           `json:"client,omitempty"`
	Jsonnet               *JsonnetConfig           `json:"jsonnet,omitempty"`
}

type JsonnetConfig struct {
	LibraryLabelSelector *metav1.LabelSelector `json:"libraryLabelSelector,omitempty"`
}

// GrafanaClient contains the Grafana API client settings
type GrafanaClient struct {
	// +nullable
	TimeoutSeconds *int `json:"timeout,omitempty"`
	// +nullable
	PreferIngress *bool `json:"preferIngress,omitempty"`
}

type GrafanaConfig struct {
	// +kubebuilder:pruning:PreserveUnknownFields
	Paths map[string]string `json:"paths,omitempty" ini:"paths,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Server map[string]string `json:"server,omitempty" ini:"server,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Database map[string]string `json:"database,omitempty" ini:"database,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	RemoteCache map[string]string `json:"remote_cache,omitempty" ini:"remote_cache,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Security map[string]string `json:"security,omitempty" ini:"security,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Users map[string]string `json:"users,omitempty" ini:"users,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Auth map[string]string `json:"auth,omitempty" ini:"auth,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthBasic map[string]string `json:"auth.basic,omitempty" ini:"auth.basic,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthAnonymous map[string]string `json:"auth.anonymous,omitempty" ini:"auth.anonymous,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthAzureAD map[string]string `json:"auth.azuread,omitempty" ini:"auth.azuread,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthGoogle map[string]string `json:"auth.google,omitempty" ini:"auth.google,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthGithub map[string]string `json:"auth.github,omitempty" ini:"auth.github,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthGitlab map[string]string `json:"auth.gitlab,omitempty" ini:"auth.gitlab,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthGenericOauth map[string]string `json:"auth.generic_oauth,omitempty" ini:"auth.generic_oauth,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthOkta map[string]string `json:"auth.okta,omitempty" ini:"auth.okta,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthLdap map[string]string `json:"auth.ldap,omitempty" ini:"auth.ldap,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthProxy map[string]string `json:"auth.proxy,omitempty" ini:"auth.proxy,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	AuthSaml map[string]string `json:"auth.saml,omitempty" ini:"auth.saml,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	DataProxy map[string]string `json:"dataproxy,omitempty" ini:"dataproxy,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Analytics map[string]string `json:"analytics,omitempty" ini:"analytics,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Dashboards map[string]string `json:"dashboards,omitempty" ini:"dashboards,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Smtp map[string]string `json:"smtp,omitempty" ini:"smtp,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Live map[string]string `json:"live,omitempty" ini:"live,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Log map[string]string `json:"log,omitempty" ini:"log,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	LogConsole map[string]string `json:"log.console,omitempty" ini:"log.console,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	LogFrontend map[string]string `json:"log.frontend,omitempty" ini:"log.frontend,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Metrics map[string]string `json:"metrics,omitempty" ini:"metrics,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	MetricsGraphite map[string]string `json:"metrics.graphite,omitempty" ini:"metrics.graphite,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Snapshots map[string]string `json:"snapshots,omitempty" ini:"snapshots,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ExternalImageStorage map[string]string `json:"external_image_storage,omitempty" ini:"external_image_storage,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ExternalImageStorageS3 map[string]string `json:"external_image_storage.s3,omitempty" ini:"external_image_storage.s3,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ExternalImageStorageWebdav map[string]string `json:"external_image_storage.webdav,omitempty" ini:"external_image_storage.webdav,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ExternalImageStorageGcs map[string]string `json:"external_image_storage.gcs,omitempty" ini:"external_image_storage.gcs,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	ExternalImageStorageAzureBlob map[string]string `json:"external_image_storage.azure_blob,omitempty" ini:"external_image_storage.azure_blob,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Alerting map[string]string `json:"alerting,omitempty" ini:"alerting,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	UnifiedAlerting map[string]string `json:"unified_alerting,omitempty" ini:"unified_alerting,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Panels map[string]string `json:"panels,omitempty" ini:"panels,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Plugins map[string]string `json:"plugins,omitempty" ini:"plugins,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	Rendering map[string]string `json:"rendering,omitempty" ini:"rendering,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	FeatureToggles map[string]string `json:"feature_toggles,omitempty" ini:"feature_toggles,omitempty"`
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Stage       OperatorStageName   `json:"stage,omitempty"`
	StageStatus OperatorStageStatus `json:"stageStatus,omitempty"`
	LastMessage string              `json:"lastMessage,omitempty"`
	AdminUrl    string              `json:"adminUrl,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Grafana is the Schema for the grafanas API
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaSpec   `json:"spec,omitempty"`
	Status GrafanaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaList contains a list of Grafana
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Grafana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Grafana{}, &GrafanaList{})
}

func (r *Grafana) PreferIngress() bool {
	return r.Spec.Client != nil && r.Spec.Client.PreferIngress != nil && *r.Spec.Client.PreferIngress
}
