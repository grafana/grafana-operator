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

func (in *Grafana) GetDashboards() NamespacedResources {
	dashboards := NamespacedResources{}
	dashboards.Deserialize(in.Annotations[AnnotationDashboards])
	return dashboards
}

func (in *Grafana) FindDashboardByNamespaceAndName(namespace string, name string) (bool, string) {
	managedDashboards := in.GetDashboards()
	for ns, dashboards := range managedDashboards {
		if ns == namespace {
			for _, dashboard := range dashboards {
				if dashboard.Name == name {
					return true, dashboard.UID
				}
			}
		}
	}
	return false, ""
}

func (in *Grafana) GetFolders() NamespacedResources {
	folders := NamespacedResources{}
	folders.Deserialize(in.Annotations[AnnotationFolders])
	return folders
}

func (in *Grafana) FindFolderByNamespaceAndName(namespace, name string) (bool, string) {
	managedFolders := in.GetFolders()
	for ns, folders := range managedFolders {
		if ns == namespace {
			for _, folder := range folders {
				if folder.Name == name {
					return true, folder.UID
				}
			}
		}
	}
	return false, ""
}

func (in *Grafana) FindDashboardByUID(uid string) bool {
	managedDashboards := in.GetDashboards()
	for _, dashboards := range managedDashboards {
		for _, dashboard := range dashboards {
			if dashboard.UID == uid {
				return true
			}
		}
	}
	return false
}

func (in *Grafana) AddDashboard(namespace string, name string, uid string) error {
	managedDashboards := in.GetDashboards()
	newDashboards := managedDashboards.AddResource(namespace, name, uid)
	bytes, err := newDashboards.Serialize()
	if err != nil {
		return err
	}
	in.Annotations[AnnotationDashboards] = string(bytes)
	return nil
}

func (in *Grafana) RemoveDashboard(namespace string, name string) error {
	managedDashboards := in.GetDashboards()
	newDashboards := managedDashboards.RemoveResource(namespace, name)
	bytes, err := newDashboards.Serialize()
	if err != nil {
		return err
	}
	in.Annotations[AnnotationDashboards] = string(bytes)
	return nil
}

func (in *Grafana) GetDatasources() NamespacedResources {
	datasources := NamespacedResources{}
	datasources.Deserialize(in.Annotations[AnnotationDatasources])
	return datasources
}

func (in *Grafana) FindDatasourceByNamespaceAndName(namespace string, name string) (bool, string) {
	managedDatasources := in.GetDatasources()
	for ns, datasources := range managedDatasources {
		if ns == namespace {
			for _, ds := range datasources {
				if ds.Name == name {
					return true, ds.UID
				}
			}
		}
	}
	return false, ""
}

func (in *Grafana) FindDatasourceByUID(uid string) bool {
	managedDatasources := in.GetDashboards()
	for _, datasources := range managedDatasources {
		for _, ds := range datasources {
			if ds.UID == uid {
				return true
			}
		}
	}
	return false
}

func (in *Grafana) AddDatasource(namespace string, name string, uid string) error {
	managedDatasources := in.GetDatasources()
	newDatasources := managedDatasources.AddResource(namespace, name, uid)
	bytes, err := newDatasources.Serialize()
	if err != nil {
		return err
	}
	in.Annotations[AnnotationDatasources] = string(bytes)
	return nil
}

func (in *Grafana) RemoveDatasource(namespace string, name string) error {
	managedDatasources := in.GetDatasources()
	newDatasources := managedDatasources.RemoveResource(namespace, name)
	bytes, err := newDatasources.Serialize()
	if err != nil {
		return err
	}
	in.Annotations[AnnotationDatasources] = string(bytes)
	return nil
}
