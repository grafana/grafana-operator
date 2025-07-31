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

	"github.com/grafana/grafana-openapi-client-go/models"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GrafanaNotificationPolicySpec defines the desired state of GrafanaNotificationPolicy
// +kubebuilder:validation:XValidation:rule="((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable)))", message="spec.editable is immutable"
type GrafanaNotificationPolicySpec struct {
	GrafanaCommonSpec `json:",inline"`

	// Routes for alerts to match against
	Route *Route `json:"route"`

	// Whether to enable or disable editing of the notification policy in Grafana UI
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="Value is immutable"
	// +optional
	Editable *bool `json:"editable,omitempty"`
}

type Route struct {
	// continue
	Continue bool `json:"continue,omitempty"`

	// group by
	GroupBy []string `json:"group_by,omitempty"`

	// group interval
	GroupInterval string `json:"group_interval,omitempty"`

	// group wait
	GroupWait string `json:"group_wait,omitempty"`

	// match re
	MatchRe models.MatchRegexps `json:"match_re,omitempty"`

	// matchers
	Matchers Matchers `json:"matchers,omitempty"`

	// mute time intervals
	MuteTimeIntervals []string `json:"mute_time_intervals,omitempty"`

	// object matchers
	ObjectMatchers models.ObjectMatchers `json:"object_matchers,omitempty"`

	// provenance
	Provenance models.Provenance `json:"provenance,omitempty"`

	// receiver
	// +kubebuilder:validation:MinLength=1
	Receiver string `json:"receiver"`

	// repeat interval
	RepeatInterval string `json:"repeat_interval,omitempty"`

	// selects GrafanaNotificationPolicyRoutes to merge in when specified
	// mutually exclusive with Routes
	RouteSelector *metav1.LabelSelector `json:"routeSelector,omitempty"`

	// routes, mutually exclusive with RouteSelector
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Schemaless
	Routes []*Route `json:"routes,omitempty"`
}

type Matcher struct {
	// is equal
	IsEqual bool `json:"isEqual,omitempty"`

	// is regex
	IsRegex bool `json:"isRegex"`

	// name
	Name *string `json:"name,omitempty"`

	// value
	Value string `json:"value"`
}
type Matchers []*Matcher

func (m Matchers) ToModelMatchers() models.Matchers {
	out := make(models.Matchers, len(m))
	for i, v := range m {
		out[i] = &models.Matcher{
			IsEqual: v.IsEqual,
			IsRegex: &v.IsRegex,
			Name:    v.Name,
			Value:   &v.Value,
		}
	}

	return out
}

func (r *Route) ToModelRoute() *models.Route {
	out := &models.Route{
		Continue:          r.Continue,
		GroupBy:           r.GroupBy,
		GroupInterval:     r.GroupInterval,
		GroupWait:         r.GroupWait,
		MatchRe:           r.MatchRe,
		Matchers:          r.Matchers.ToModelMatchers(),
		MuteTimeIntervals: r.MuteTimeIntervals,
		ObjectMatchers:    r.ObjectMatchers,
		Provenance:        r.Provenance,
		Receiver:          r.Receiver,
		RepeatInterval:    r.RepeatInterval,
		Routes:            make([]*models.Route, len(r.Routes)),
	}
	for i, v := range r.Routes {
		out.Routes[i] = v.ToModelRoute()
	}

	return out
}

// selectorMutuallyExclusive checks if a single route satisfies the mutual exclusivity constraint
// for checking the entire route including nested routes, use IsRouteSelectorMutuallyExclusive
func (r *Route) selectorMutuallyExclusive() bool {
	return !(r.RouteSelector != nil && len(r.Routes) > 0) // nolint:staticcheck
}

// IsRouteSelectorMutuallyExclusive returns true when the route and all its sub-routes
// satisfy the constraint of routes and routeSelector being mutually exclusive
func (r *Route) IsRouteSelectorMutuallyExclusive() bool {
	if !r.selectorMutuallyExclusive() {
		return false
	}

	// Recursively check all child routes
	for _, childRoute := range r.Routes {
		if !childRoute.IsRouteSelectorMutuallyExclusive() {
			return false
		}
	}

	return true
}

// HasRouteSelector checks if the given Route or any of its nested Routes has a RouteSelector
func (r *Route) HasRouteSelector() bool {
	if r.RouteSelector != nil {
		return true
	}

	for _, nestedRoute := range r.Routes {
		if nestedRoute.HasRouteSelector() {
			return true
		}
	}

	return false
}

// GrafanaNotificationPolicyStatus defines the observed state of GrafanaNotificationPolicy
type GrafanaNotificationPolicyStatus struct {
	GrafanaCommonStatus `json:",inline"`

	DiscoveredRoutes *[]string `json:"discoveredRoutes,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaNotificationPolicy is the Schema for the GrafanaNotificationPolicy API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaNotificationPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaNotificationPolicySpec   `json:"spec"`
	Status GrafanaNotificationPolicyStatus `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaNotificationPolicy)(nil)

func (in *GrafanaNotificationPolicy) NamespacedResource() string {
	return fmt.Sprintf("%v/%v/%v", in.Namespace, in.Name, in.UID)
}

//+kubebuilder:object:root=true

// GrafanaNotificationPolicyList contains a list of GrafanaNotificationPolicy
type GrafanaNotificationPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaNotificationPolicy `json:"items"`
}

func (in *GrafanaNotificationPolicy) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaNotificationPolicy) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaNotificationPolicy) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaNotificationPolicy) CommonStatus() *GrafanaCommonStatus {
	return &in.Status.GrafanaCommonStatus
}

func init() {
	SchemeBuilder.Register(&GrafanaNotificationPolicy{}, &GrafanaNotificationPolicyList{})
}
