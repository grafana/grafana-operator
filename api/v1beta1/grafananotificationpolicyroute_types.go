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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaNotificationPolicyRouteSpec defines the desired state of GrafanaNotificationPolicyRoute
type GrafanaNotificationPolicyRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Route for alerts to match against
	Route `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaNotificationPolicyRoute is the Schema for the grafananotificationpolicyroutes API
type GrafanaNotificationPolicyRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationPolicyRouteSpec `json:"spec,omitempty"`
	Status GrafanaCommonStatus                `json:"status,omitempty"`
}

func (r *GrafanaNotificationPolicyRoute) NamespacedResource() string {
	return fmt.Sprintf("%v/%v/%v", r.ObjectMeta.Namespace, r.ObjectMeta.Name, r.ObjectMeta.UID)
}

//+kubebuilder:object:root=true

// GrafanaNotificationPolicyRouteList contains a list of GrafanaNotificationPolicyRoute
type GrafanaNotificationPolicyRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationPolicyRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationPolicyRoute{}, &GrafanaNotificationPolicyRouteList{})
}
