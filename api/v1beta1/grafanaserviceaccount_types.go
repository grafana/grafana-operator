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

// GrafanaServiceAccountTokenSpec defines a token for a service account
type GrafanaServiceAccountTokenSpec struct {
	// Name of the token
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Expiration date of the token. If not set, the token never expires
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	Expires *metav1.Time `json:"expires,omitempty"`

	// Name of the secret to store the token. If not set, a name will be generated
	// +optional
	// +kubebuilder:validation:MinLength=1
	SecretName string `json:"secretName,omitempty"`
}

// GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.
type GrafanaServiceAccountSpec struct {
	// How often the resource is synced, defaults to 10m0s if not set
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"
	// +kubebuilder:validation:XValidation:rule="duration(self) > duration('0s')",message="spec.resyncPeriod must be greater than 0"
	ResyncPeriod metav1.Duration `json:"resyncPeriod,omitempty"`

	// Suspend pauses reconciliation of the service account
	// +optional
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

	// Name of the Grafana instance to create the service account for
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.instanceName is immutable"
	InstanceName string `json:"instanceName"`

	// Name of the service account in Grafana
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.name is immutable"
	Name string `json:"name"`

	// Role of the service account (Viewer, Editor, Admin)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Viewer;Editor;Admin
	Role string `json:"role"`

	// Whether the service account is disabled
	// +optional
	// +kubebuilder:default=false
	IsDisabled bool `json:"isDisabled,omitempty"`

	// Tokens to create for the service account
	// +optional
	// +listType=map
	// +listMapKey=name
	Tokens []GrafanaServiceAccountTokenSpec `json:"tokens,omitempty"`
}

// GrafanaServiceAccountSecretStatus describes a Secret created in Kubernetes to store the service account token.
type GrafanaServiceAccountSecretStatus struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
}

// GrafanaServiceAccountTokenStatus describes a token created in Grafana.
type GrafanaServiceAccountTokenStatus struct {
	Name string `json:"name"`

	// Expiration time of the token
	// N.B. There's possible discrepancy with the expiration time in spec
	// It happens because Grafana API accepts TTL in seconds then calculates the expiration time against the current time
	Expires *metav1.Time `json:"expires,omitempty"`

	// ID of the token in Grafana
	ID int64 `json:"id"`

	// Name of the secret containing the token
	Secret *GrafanaServiceAccountSecretStatus `json:"secret,omitempty"`
}

// GrafanaServiceAccountInfo describes the Grafana service account information.
type GrafanaServiceAccountInfo struct {
	Name  string `json:"name"`
	Login string `json:"login"`

	// ID of the service account in Grafana
	ID int64 `json:"id"`

	// Role is the Grafana role for the service account (Viewer, Editor, Admin)
	Role string `json:"role"`

	// IsDisabled indicates if the service account is disabled
	IsDisabled bool `json:"isDisabled"`

	// Information about tokens
	// +optional
	Tokens []GrafanaServiceAccountTokenStatus `json:"tokens,omitempty"`
}

// GrafanaServiceAccountStatus defines the observed state of a GrafanaServiceAccount
type GrafanaServiceAccountStatus struct {
	GrafanaCommonStatus `json:",inline"`

	// Info contains the Grafana service account information
	Account *GrafanaServiceAccountInfo `json:"account,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaServiceAccount is the Schema for the grafanaserviceaccounts API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaServiceAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaServiceAccountSpec   `json:"spec,omitempty"`
	Status GrafanaServiceAccountStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaServiceAccount)(nil)

//+kubebuilder:object:root=true

// GrafanaServiceAccountList contains a list of GrafanaServiceAccount
type GrafanaServiceAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaServiceAccount `json:"items"`
}

// Find searches for a GrafanaServiceAccount by namespace/name in the list.
func (in *GrafanaServiceAccountList) Find(namespace, name string) *GrafanaServiceAccount {
	for i := range in.Items {
		if in.Items[i].Namespace == namespace && in.Items[i].Name == name {
			return &in.Items[i]
		}
	}

	return nil
}

// MatchLabels is a no-op for GrafanaServiceAccount
func (in *GrafanaServiceAccount) MatchLabels() *metav1.LabelSelector {
	labels := &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"non-existent-set-of-labels": "no-op",
		},
	}

	return labels
}

func (in *GrafanaServiceAccount) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaServiceAccount) Metadata() metav1.ObjectMeta {
	return in.ObjectMeta
}

// AllowCrossNamespace is a no-op for GrafanaServiceAccount
func (in *GrafanaServiceAccount) AllowCrossNamespace() bool {
	return false
}

func (in *GrafanaServiceAccount) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

func init() {
	SchemeBuilder.Register(&GrafanaServiceAccount{}, &GrafanaServiceAccountList{})
}
