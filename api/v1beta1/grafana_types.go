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
	OperatorStageComplete       OperatorStageName = "complete"
)

const (
	OperatorStageResultSuccess    OperatorStageStatus = "success"
	OperatorStageResultFailed     OperatorStageStatus = "failed"
	OperatorStageResultInProgress OperatorStageStatus = "in progress"
)

const (
	AnnotationDashboards  = "grafana-operator/managed-dashboards"
	AnnotationDatasources = "grafana-operator/managed-datasources"
	AnnotationFolders     = "grafana-operator/managed-folders"
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
	// +kubebuilder:pruning:PreserveUnknownFields
	Config                map[string]map[string]string `json:"config"`
	Ingress               *IngressNetworkingV1         `json:"ingress,omitempty"`
	Route                 *RouteOpenshiftV1            `json:"route,omitempty"`
	Service               *ServiceV1                   `json:"service,omitempty"`
	Deployment            *DeploymentV1                `json:"deployment,omitempty"`
	PersistentVolumeClaim *PersistentVolumeClaimV1     `json:"persistentVolumeClaim,omitempty"`
	ServiceAccount        *ServiceAccountV1            `json:"serviceAccount,omitempty"`
	Client                *GrafanaClient               `json:"client,omitempty"`
	Jsonnet               *JsonnetConfig               `json:"jsonnet,omitempty"`
	GrafanaContainer      *GrafanaContainer            `json:"grafanaContainer,omitempty"`
}

type GrafanaContainer struct {
	BaseImage string `json:"baseImage,omitempty"`
	InitImage string `json:"initImage,omitempty"`
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

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Stage       OperatorStageName      `json:"stage,omitempty"`
	StageStatus OperatorStageStatus    `json:"stageStatus,omitempty"`
	LastMessage string                 `json:"lastMessage,omitempty"`
	AdminUrl    string                 `json:"adminUrl,omitempty"`
	Dashboards  NamespacedResourceList `json:"dashboards,omitempty"`
	Datasources NamespacedResourceList `json:"datasources,omitempty"`
	Folders     NamespacedResourceList `json:"folders,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Grafana is the Schema for the grafanas API
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GrafanaSpec   `json:"spec,omitempty"`
	Status            GrafanaStatus `json:"status,omitempty"`
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

func (in *Grafana) PreferIngress() bool {
	return in.Spec.Client != nil && in.Spec.Client.PreferIngress != nil && *in.Spec.Client.PreferIngress
}
