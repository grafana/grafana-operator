/*
Copyright 2025.

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

// GrafanaServiceAccountToken describes a token to create.
type GrafanaServiceAccountToken struct {
	// Name is the name of the Kubernetes Secret (and token identifier in Grafana). The secret will contain the token value.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Expires is the optional expiration time for the token. After this time, the operator may rotate the token.
	// +kubebuilder:validation:Optional
	Expires *metav1.Time `json:"expires,omitempty"`
}

// GrafanaServiceAccountTokenStatus describes a token created in Grafana.
type GrafanaServiceAccountTokenStatus struct {
	// Name of the token (same as Secret name).
	Name string `json:"name"`

	// TokenID is the Grafana-assigned ID of the token.
	TokenID int64 `json:"tokenId"`

	// SecretName is the name of the Kubernetes Secret that stores the actual token value.
	// This may seem redundant if the Secret name usually matches the token's Name,
	// but it's stored explicitly in Status for clarity and future flexibility.
	SecretName string `json:"secretName"`
}

// GrafanaServiceAccountPermission defines a permission grant for a user or team.
// +kubebuilder:validation:XValidation:rule="(has(self.user) && self.user != '') || (has(self.team) && self.team != '')",message="one of user or team must be set"
// +kubebuilder:validation:XValidation:rule="!((has(self.user) && self.user != '') && (has(self.team) && self.team != ''))",message="user and team cannot both be set"
type GrafanaServiceAccountPermission struct {
	// User login or email to grant permissions to (optional).
	// +kubebuilder:validation:Optional
	User string `json:"user,omitempty"`

	// Team name to grant permissions to (optional).
	// +kubebuilder:validation:Optional
	Team string `json:"team,omitempty"`

	// Permission level: Edit or Admin.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Edit;Admin
	Permission string `json:"permission"`
}

// GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.
type GrafanaServiceAccountSpec struct {
	GrafanaCommonSpec `json:",inline"`

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
	Tokens []GrafanaServiceAccountToken `json:"tokens,omitempty"`

	// Permissions specifies additional Grafana permission grants for existing users or teams on this service account.
	// +kubebuilder:validation:Optional
	Permissions []GrafanaServiceAccountPermission `json:"permissions,omitempty"`

	// GenerateTokenSecret, if true, will create one default API token in a Secret if no Tokens are specified.
	// If false, no token is created unless explicitly listed in Tokens.
	// +kubebuilder:default=true
	GenerateTokenSecret bool `json:"generateTokenSecret,omitempty"`
}

// GrafanaServiceAccountInstanceStatus holds status for one Grafana instance.
type GrafanaServiceAccountInstanceStatus struct {
	// GrafanaNamespace and GrafanaName specify which Grafana resource this status record belongs to.
	GrafanaNamespace string `json:"grafanaNamespace"`
	GrafanaName      string `json:"grafanaName"`

	// ServiceAccountID is the numeric ID of the service account in this Grafana.
	ServiceAccountID int64 `json:"serviceAccountID"`

	// Tokens is the status of tokens for this service account in Grafana.
	Tokens []GrafanaServiceAccountTokenStatus `json:"tokens,omitempty"`
}

// GrafanaServiceAccountStatus defines the observed state of a GrafanaServiceAccount.
type GrafanaServiceAccountStatus struct {
	GrafanaCommonStatus `json:",inline"`

	// Instances lists Grafana instances where this service account is applied.
	// +optional
	Instances []GrafanaServiceAccountInstanceStatus `json:"instances,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaServiceAccount is the Schema for the grafanaserviceaccounts API.
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaServiceAccountSpec   `json:"spec,omitempty"`
	Status GrafanaServiceAccountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaServiceAccountList contains a list of GrafanaServiceAccount objects.
type GrafanaServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaServiceAccount `json:"items"`
}

// Find searches for a GrafanaServiceAccount by namespace/name in the list.
func (in *GrafanaServiceAccountList) Find(namespace, name string) *GrafanaServiceAccount {
	for _, serviceAccount := range in.Items {
		if serviceAccount.Namespace == namespace && serviceAccount.Name == name {
			return &serviceAccount
		}
	}
	return nil
}

// MatchLabels returns the LabelSelector (from GrafanaCommonSpec) to find matching Grafana instances.
func (in *GrafanaServiceAccount) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

// MatchNamespace returns the namespace where this service account is defined.
func (in *GrafanaServiceAccount) MatchNamespace() string {
	return in.ObjectMeta.Namespace
}

// AllowCrossNamespace indicates whether cross-namespace import is allowed for this resource.
func (in *GrafanaServiceAccount) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func init() {
	SchemeBuilder.Register(&GrafanaServiceAccount{}, &GrafanaServiceAccountList{})
}
