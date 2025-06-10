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

// GrafanaServiceAccountTokenSpec describes a token to create.
type GrafanaServiceAccountTokenSpec struct {
	// Name is the name of the Kubernetes Secret (and token identifier in Grafana). The secret will contain the token value.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Expires is the optional expiration time for the token. After this time, the operator may rotate the token.
	// +kubebuilder:validation:Optional
	Expires *metav1.Time `json:"expires,omitempty"`
}

// GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.
type GrafanaServiceAccountSpec struct {
	// ID is a kind of unique identifier to distinguish between service accounts if the name is changed.
	// +kubebuilder:validation:Required
	ID string `json:"id"`

	// Name is the desired name of the service account in Grafana.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Role is the Grafana role for the service account (Viewer, Editor, Admin).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Viewer;Editor;Admin
	Role string `json:"role"`

	// IsDisabled indicates if the service account should be disabled in Grafana.
	// +kubebuilder:validation:Optional
	IsDisabled bool `json:"isDisabled,omitempty"`

	// Tokens defines API tokens to create for this service account. Each token will be stored in a Kubernetes Secret with the given name.
	// +kubebuilder:validation:Optional
	Tokens []GrafanaServiceAccountTokenSpec `json:"tokens,omitempty"`
}

type GrafanaServiceAccounts struct {
	// Accounts lists Grafana service accounts to manage.
	// Each service account is uniquely identified by its ID.
	// +listType=map
	// +listMapKey=id
	Accounts []GrafanaServiceAccountSpec `json:"accounts,omitempty"`

	// GenerateTokenSecret, if true, will create one default API token in a Secret if no Tokens are specified.
	// If false, no token is created unless explicitly listed in Tokens.
	// +kubebuilder:default=true
	GenerateTokenSecret bool `json:"generateTokenSecret,omitempty"`
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
	// Service sets how the service object should look like with your grafana instance, contains a number of defaults.
	Service *ServiceV1 `json:"service,omitempty"`
	// Version specifies the version of Grafana to use for this deployment. It follows the same format as the docker.io/grafana/grafana tags
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
	// DisableDefaultSecurityContext prevents the operator from populating securityContext on deployments
	// +kubebuilder:validation:Enum=Pod;Container;All
	DisableDefaultSecurityContext string `json:"disableDefaultSecurityContext,omitempty"`
	// Grafana Service Accounts
	GrafanaServiceAccounts *GrafanaServiceAccounts `json:"grafanaServiceAccounts,omitempty"`
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

type GrafanaServiceAccountSecretStatus struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// GrafanaServiceAccountTokenStatus describes a token created in Grafana.
type GrafanaServiceAccountTokenStatus struct {
	// Name is the name of the Kubernetes Secret. The secret will contain the token value.
	Name string `json:"name"`

	// Expires is the expiration time for the token.
	// N.B. There's possible discrepancy with the expiration time in spec.
	// It happens because Grafana API accepts TTL in seconds then calculates the expiration time against the current time.
	Expires *metav1.Time `json:"expires,omitempty"`

	// ID is the Grafana-assigned ID of the token.
	ID int64 `json:"tokenId"`

	// Secret is the Kubernetes Secret that stores the actual token value.
	// This may seem redundant if the Secret name usually matches the token's Name,
	// but it's stored explicitly in Status for clarity and future flexibility.
	Secret *GrafanaServiceAccountSecretStatus `json:"secret,omitempty"`
}

// GrafanaServiceAccountStatus holds status for one Grafana instance.
type GrafanaServiceAccountStatus struct {
	// SpecID is a kind of unique identifier to distinguish between service accounts if the name is changed.
	SpecID string `json:"specId"`

	// Name is the name of the service account in Grafana.
	Name string `json:"name"`

	// ServiceAccountID is the numeric ID of the service account in this Grafana.
	ServiceAccountID int64 `json:"serviceAccountId"`

	// Role is the Grafana role for the service account (Viewer, Editor, Admin).
	Role string `json:"role"`

	// IsDisabled indicates if the service account is disabled.
	IsDisabled bool `json:"isDisabled,omitempty"`

	// Tokens is the status of tokens for this service account in Grafana.
	Tokens []GrafanaServiceAccountTokenStatus `json:"tokens,omitempty"`
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Stage                  OperatorStageName             `json:"stage,omitempty"`
	StageStatus            OperatorStageStatus           `json:"stageStatus,omitempty"`
	LastMessage            string                        `json:"lastMessage,omitempty"`
	AdminURL               string                        `json:"adminUrl,omitempty"`
	AlertRuleGroups        NamespacedResourceList        `json:"alertRuleGroups,omitempty"`
	ContactPoints          NamespacedResourceList        `json:"contactPoints,omitempty"`
	Dashboards             NamespacedResourceList        `json:"dashboards,omitempty"`
	Datasources            NamespacedResourceList        `json:"datasources,omitempty"`
	Folders                NamespacedResourceList        `json:"folders,omitempty"`
	LibraryPanels          NamespacedResourceList        `json:"libraryPanels,omitempty"`
	MuteTimings            NamespacedResourceList        `json:"muteTimings,omitempty"`
	NotificationTemplates  NamespacedResourceList        `json:"notificationTemplates,omitempty"`
	Version                string                        `json:"version,omitempty"`
	Conditions             []metav1.Condition            `json:"conditions,omitempty"`
	GrafanaServiceAccounts []GrafanaServiceAccountStatus `json:"serviceAccounts,omitempty"`
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

func (in *Grafana) PreferIngress() bool {
	return in.Spec.Client != nil && in.Spec.Client.PreferIngress != nil && *in.Spec.Client.PreferIngress
}

func (in *Grafana) IsInternal() bool {
	return in.Spec.External == nil
}

func (in *Grafana) IsExternal() bool {
	return in.Spec.External != nil
}
