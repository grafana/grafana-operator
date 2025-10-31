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
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PluginVersionLatest string = "latest"
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
	OperatorStageHTTPRoute      OperatorStageName = "http route"
	OperatorStagePlugins        OperatorStageName = "plugins"
	OperatorStageDeployment     OperatorStageName = "deployment"
	OperatorStageComplete       OperatorStageName = "complete"
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
	// +kubebuilder:pruning:PreserveUnknownFields
	// Config defines how your grafana ini file should looks like.
	Config map[string]map[string]string `json:"config,omitempty"`
	// Ingress sets how the ingress object should look like with your grafana instance.
	Ingress *IngressNetworkingV1 `json:"ingress,omitempty"`
	// Route sets how the ingress object should look like with your grafana instance, this only works in Openshift.
	Route *RouteOpenshiftV1 `json:"route,omitempty"`
	// HTTPRoute sets how the ingress object should look like with your grafana instance, this only works use gateway api.
	HTTPRoute *HTTPRouteV1 `json:"httpRoute,omitempty"`
	// Service sets how the service object should look like with your grafana instance, contains a number of defaults.
	Service *ServiceV1 `json:"service,omitempty"`
	// Version sets the tag of the default image: docker.io/grafana/grafana.
	// Allows full image refs with/without sha256checksum: "registry/repo/image:tag@sha"
	Version string `json:"version,omitempty"`
	// Deployment sets how the deployment object should look like with your grafana instance, contains a number of defaults.
	Deployment *DeploymentV1 `json:"deployment,omitempty"`
	// PersistentVolumeClaim creates a PVC if you need to attach one to your grafana instance.
	PersistentVolumeClaim *PersistentVolumeClaimV1 `json:"persistentVolumeClaim,omitempty"`
	// ServiceAccount sets how the ServiceAccount object should look like with your grafana instance, contains a number of defaults.
	ServiceAccount *ServiceAccountV1 `json:"serviceAccount,omitempty"`
	// Client defines how the grafana-operator talks to the grafana instance.
	Client  *GrafanaClient `json:"client,omitempty"`
	Jsonnet *JsonnetConfig `json:"jsonnet,omitempty"`
	// External enables you to configure external grafana instances that is not managed by the operator.
	External *External `json:"external,omitempty"`
	// Preferences holds the Grafana Preferences settings
	Preferences *GrafanaPreferences `json:"preferences,omitempty"`
	// DisableDefaultAdminSecret prevents operator from creating default admin-credentials secret
	DisableDefaultAdminSecret bool `json:"disableDefaultAdminSecret,omitempty"`
	// Suspend pauses reconciliation of owned resources like deployments, Services, Etc. upon changes
	// +optional
	Suspend bool `json:"suspend,omitempty"`
	// DisableDefaultSecurityContext prevents the operator from populating securityContext on deployments
	// +kubebuilder:validation:Enum=Pod;Container;All
	DisableDefaultSecurityContext string `json:"disableDefaultSecurityContext,omitempty"`
}

type External struct {
	// URL of the external grafana instance you want to manage.
	URL string `json:"url"`
	// The API key to talk to the external grafana instance, you need to define ether apiKey or adminUser/adminPassword.
	APIKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// AdminUser key to talk to the external grafana instance.
	AdminUser *v1.SecretKeySelector `json:"adminUser,omitempty"`
	// AdminPassword key to talk to the external grafana instance.
	AdminPassword *v1.SecretKeySelector `json:"adminPassword,omitempty"`
	// DEPRECATED, use top level `tls` instead.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
}

// TLSConfig specifies options to use when communicating with the Grafana endpoint
// +kubebuilder:validation:XValidation:rule="(has(self.insecureSkipVerify) && !(has(self.certSecretRef))) || (has(self.certSecretRef) && !(has(self.insecureSkipVerify)))", message="insecureSkipVerify and certSecretRef cannot be set at the same time"
type TLSConfig struct {
	// Disable the CA check of the server
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
	// Use a secret as a reference to give TLS Certificate information
	// +optional
	CertSecretRef *v1.SecretReference `json:"certSecretRef,omitempty"`
}

type JsonnetConfig struct {
	LibraryLabelSelector *metav1.LabelSelector `json:"libraryLabelSelector,omitempty"`
}

// GrafanaClient contains the Grafana API client settings
type GrafanaClient struct {
	// Use Kubernetes Serviceaccount as authentication
	// Requires configuring [auth.jwt] in the instance
	// +optional
	UseKubeAuth bool `json:"useKubeAuth,omitempty"`
	// +nullable
	TimeoutSeconds *int `json:"timeout,omitempty"`
	// +nullable
	// If the operator should send it's request through the grafana instances ingress object instead of through the service.
	PreferIngress *bool `json:"preferIngress,omitempty"`
	// TLS Configuration used to talk with the grafana instance.
	// +optional
	TLS *TLSConfig `json:"tls,omitempty"`
	// Custom HTTP headers to use when interacting with this Grafana.
	// +optional
	Headers map[string]string `json:"headers,omitempty"`
}

// GrafanaPreferences holds Grafana preferences API settings
type GrafanaPreferences struct {
	HomeDashboardUID string `json:"homeDashboardUid,omitempty"`
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Stage                 OperatorStageName      `json:"stage,omitempty"`
	StageStatus           OperatorStageStatus    `json:"stageStatus,omitempty"`
	LastMessage           string                 `json:"lastMessage,omitempty"`
	AdminURL              string                 `json:"adminUrl,omitempty"`
	AlertRuleGroups       NamespacedResourceList `json:"alertRuleGroups,omitempty"`
	ContactPoints         NamespacedResourceList `json:"contactPoints,omitempty"`
	Dashboards            NamespacedResourceList `json:"dashboards,omitempty"`
	Datasources           NamespacedResourceList `json:"datasources,omitempty"`
	ServiceAccounts       NamespacedResourceList `json:"serviceaccounts,omitempty"`
	Folders               NamespacedResourceList `json:"folders,omitempty"`
	LibraryPanels         NamespacedResourceList `json:"libraryPanels,omitempty"`
	MuteTimings           NamespacedResourceList `json:"muteTimings,omitempty"`
	NotificationTemplates NamespacedResourceList `json:"notificationTemplates,omitempty"`
	Version               string                 `json:"version,omitempty"`
	Conditions            []metav1.Condition     `json:"conditions,omitempty"`
}

func (in *GrafanaStatus) StatusList(cr client.Object) (*NamespacedResourceList, string, error) {
	switch t := cr.(type) {
	case *GrafanaAlertRuleGroup:
		return &in.AlertRuleGroups, "alertRuleGroups", nil
	case *GrafanaContactPoint:
		return &in.ContactPoints, "contactPoints", nil
	case *GrafanaDashboard:
		return &in.Dashboards, "dashboards", nil
	case *GrafanaDatasource:
		return &in.Datasources, "datasources", nil
	case *GrafanaFolder:
		return &in.Folders, "folders", nil
	case *GrafanaLibraryPanel:
		return &in.LibraryPanels, "libraryPanels", nil
	case *GrafanaMuteTiming:
		return &in.MuteTimings, "muteTimings", nil
	case *GrafanaNotificationTemplate:
		return &in.NotificationTemplates, "notificationTemplates", nil
	default:
		return nil, "", fmt.Errorf("unknown struct %T, extend Grafana.StatusListName", t)
	}
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Grafana is the Schema for the grafanas API
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description=""
// +kubebuilder:printcolumn:name="Stage",type="string",JSONPath=".status.stage",description=""
// +kubebuilder:printcolumn:name="Stage status",type="string",JSONPath=".status.stageStatus",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GrafanaSpec   `json:"spec"`
	Status            GrafanaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GrafanaList contains a list of Grafana
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Grafana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Grafana{}, &GrafanaList{})
}

func (in *Grafana) GetConfigSection(name string) map[string]string {
	if in.Spec.Config == nil {
		return map[string]string{}
	}

	if in.Spec.Config[name] == nil {
		return map[string]string{}
	}

	return in.Spec.Config[name]
}

func (in *Grafana) GetConfigSectionValue(name, key string) string {
	section := in.GetConfigSection(name)

	return section[key]
}

func (in *Grafana) PreferIngress() bool {
	return in.Spec.Client != nil && in.Spec.Client.PreferIngress != nil && *in.Spec.Client.PreferIngress
}

func (in *Grafana) IsInternal() bool {
	return in.Spec.External == nil
}

func (in *Grafana) IsExternal() bool {
	return in.Spec.External != nil
}

// Adds a resource to the end of the Grafana status list matching 'kind'
func (in *Grafana) AddNamespacedResource(ctx context.Context, cl client.Client, cr client.Object, r NamespacedResource) error {
	list, kind, err := in.Status.StatusList(cr)
	if err != nil {
		return err
	}

	// Allows overwriting resources with old UIDs/Names below
	idx := list.IndexOf(cr.GetNamespace(), cr.GetName())

	var jsonPatch []any

	switch {
	case len(*list) == 0:
		// Create list if previously empty or append to the end
		jsonPatch = []any{map[string]any{
			"op":    "add",
			"path":  fmt.Sprintf("/status/%s", kind),
			"value": []string{string(r)},
		}}
	case idx == -1:
		// Add resource to the end of the status list
		jsonPatch = []any{map[string]any{
			"op":    "add",
			"path":  fmt.Sprintf("/status/%s/-", kind),
			"value": string(r),
		}}
	case r == (*list)[idx]:
		// Do nothing if resource already exists with the same UID/Name.
		return nil
	default:
		// Overwrite old entry, prevents old UIDs/Names to block status updates
		jsonPatch = []any{map[string]any{
			"op":    "replace",
			"path":  fmt.Sprintf("/status/%s/%d", kind, idx),
			"value": string(r),
		}}
	}

	patch, err := json.Marshal(jsonPatch)
	if err != nil {
		return err
	}

	return cl.Status().Patch(ctx, in, client.RawPatch(types.JSONPatchType, patch))
}

// Removes a resource at index from the Grafana status list matching 'kind'
func (in *Grafana) RemoveNamespacedResource(ctx context.Context, cl client.Client, cr client.Object) error {
	list, kind, err := in.Status.StatusList(cr)
	if err != nil {
		return err
	}

	idx := list.IndexOf(cr.GetNamespace(), cr.GetName())
	if idx == -1 {
		return nil
	}

	// Create patch removing entry
	jsonPatch := []any{map[string]any{
		"op":   "remove",
		"path": fmt.Sprintf("/status/%s/%d", kind, idx),
	}}

	patch, err := json.Marshal(jsonPatch)
	if err != nil {
		return err
	}

	return cl.Status().Patch(ctx, in, client.RawPatch(types.JSONPatchType, patch))
}
