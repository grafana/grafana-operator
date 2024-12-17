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
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaNotificationPolicyRouteSpec defines the desired state of GrafanaNotificationPolicyRoute
type GrafanaNotificationPolicyRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	// +optional
	Priority *int8 `json:"priority,omitempty"`

	// Route for alerts to match against
	Route *Route `json:"route"`
}

// GrafanaNotificationPolicyRouteStatus defines the observed state of GrafanaNotificationPolicyRoute
type GrafanaNotificationPolicyRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaNotificationPolicyRoute is the Schema for the grafananotificationpolicyroutes API
type GrafanaNotificationPolicyRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationPolicyRouteSpec   `json:"spec,omitempty"`
	Status GrafanaNotificationPolicyRouteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaNotificationPolicyRouteList contains a list of GrafanaNotificationPolicyRoute
type GrafanaNotificationPolicyRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationPolicyRoute `json:"items"`
}

// Implement sort.Interface for GrafanaNotificationPolicyRouteList

func (l GrafanaNotificationPolicyRouteList) Len() int {
	return len(l.Items)
}

func (l GrafanaNotificationPolicyRouteList) Less(i, j int) bool {
	iPriority := l.Items[i].Spec.Priority
	jPriority := l.Items[j].Spec.Priority

	// If both priorities are nil, maintain original order
	if iPriority == nil && jPriority == nil {
		return i < j
	}

	// Nil priorities are considered lower (come later)
	if iPriority == nil {
		return false
	}
	if jPriority == nil {
		return true
	}

	// Compare non-nil priorities
	return *iPriority < *jPriority
}

func (l GrafanaNotificationPolicyRouteList) Swap(i, j int) {
	l.Items[i], l.Items[j] = l.Items[j], l.Items[i]
}

// SortByPriority sorts the list by Priority
// Priority can be 1-100 or nil, with nil being the lowest priority 100
func (l *GrafanaNotificationPolicyRouteList) SortByPriority() {
	sort.Sort(l)
}

// StatusDiscoveredRoutes returns the list of discovered routes using the namespace and name
// Used to display all discovered routes in the GrafanaNotificationPolicy status
func (l *GrafanaNotificationPolicyRouteList) StatusDiscoveredRoutes() []string {
	sort.Sort(l)

	discoveredRoutes := make([]string, len(l.Items))
	for i, route := range l.Items {
		priority := "nil"
		if route.Spec.Priority != nil {
			priority = fmt.Sprintf("%d", *route.Spec.Priority)
		}
		discoveredRoutes[i] = fmt.Sprintf("%s/%s (priority: %s)", route.Namespace, route.Name, priority)
	}

	return discoveredRoutes
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationPolicyRoute{}, &GrafanaNotificationPolicyRouteList{})
}
