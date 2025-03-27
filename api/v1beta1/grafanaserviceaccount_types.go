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

// GrafanaServiceAccountToken defines a token to be created for a Grafana service account.
type GrafanaServiceAccountToken struct {
	// Name is the name of the Kubernetes Secret in which this token will be stored.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Expires specifies the expiration timestamp (TTL) for this token. If set, the operator
	// will rotate or replace the token once the specified expiration time is reached.
	// If not set, the token does not expire (defaults to never expire).
	// +kubebuilder:validation:Optional
	Expires *metav1.Time `json:"expires,omitempty"`
}

// GrafanaServiceAccountTokenStatus describes the current state of a token that was created in Grafana.
type GrafanaServiceAccountTokenStatus struct {
	// Name is the name of the token, matching the one specified in .spec.tokens or generated automatically.
	Name string `json:"name"`

	// TokenID is the numeric identifier of the token as returned by Grafana upon creation.
	TokenID int64 `json:"tokenId"`

	// SecretName is the name of the Kubernetes Secret that stores the actual token value (Key).
	SecretName string `json:"secretName"`
}

// GrafanaServiceAccountPermission defines a permission grant for a user or group related to this service account.
type GrafanaServiceAccountPermission struct {
	// +kubebuilder:validation:Optional
	User string `json:"user"`

	// +kubebuilder:validation:Optional
	Team string `json:"team"`

	// Permission is the level of access granted to that user or group for this service account
	// (e.g., "Edit" or "Admin"). Depending on the Grafana version, this might map to an RBAC role or other permissions.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Viewer;Editor;Admin
	Permission string `json:"permission"`
}

// GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.
type GrafanaServiceAccountSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Name is the name of the service account in Grafana.
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// Role is the service account role in Grafana (Viewer, Editor, Admin, etc.).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Viewer;Editor;Admin
	Role string `json:"role,omitempty"`

	// IsDisabled indicates whether the service account should be disabled in Grafana.
	// +kubebuilder:validation:Optional
	IsDisabled bool `json:"isDisabled,omitempty"`

	// Tokens is the list of tokens to create for this service account. For each token in the list,
	// the operator generates a Grafana access token and stores it in a Kubernetes Secret with the specified name.
	// If no tokens are specified and GenerateTokenSecret is true, the operator creates a default token
	// in a Secret with a default name.
	// +kubebuilder:validation:Optional
	Tokens []GrafanaServiceAccountToken `json:"tokens,omitempty"`

	// Permissions specifies additional access permissions for users or teams in Grafana
	// related to this service account. This aligns with the UI where you can grant specific
	// users or groups Edit/Admin permissions on the service account.
	// +kubebuilder:validation:Optional
	Permissions []GrafanaServiceAccountPermission `json:"permissions,omitempty"`

	// GenerateTokenSecret indicates whether the operator should automatically create a Kubernetes Secret
	// to store a token for this service account. If true (default), at least one Secret with a token will be created.
	// If false, no token is generated unless explicitly defined in Tokens.
	// +kubebuilder:default=true
	GenerateTokenSecret bool `json:"generateTokenSecret,omitempty"`
}

// GrafanaServiceAccountStatus defines the observed state of a GrafanaServiceAccount.
type GrafanaServiceAccountStatus struct {
	GrafanaCommonStatus `json:",inline"`

	// ID is the numeric identifier of the service account in Grafana.
	ID int64 `json:"id,omitempty"`

	// Tokens is a list of detailed information for each token created in Grafana.
	// +optional
	Tokens []GrafanaServiceAccountTokenStatus `json:"tokens,omitempty"`
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
