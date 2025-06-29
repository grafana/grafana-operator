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

// GrafanaMuteTimingSpec defines the desired state of GrafanaMuteTiming
type GrafanaMuteTimingSpec struct {
	GrafanaCommonSpec `json:",inline"`

	// A unique name for the mute timing
	Name string `json:"name"`

	// Time intervals for muting
	// +kubebuilder:validation:MinItems=1
	TimeIntervals []*TimeInterval `json:"time_intervals"`

	// Whether to enable or disable editing of the mute timing in Grafana UI
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.editable is immutable"
	// +optional
	// +kubebuilder:default=true
	Editable bool `json:"editable"`
}

type TimeInterval struct {
	// The date 1-31 of a month. Negative values can also be used to represent days that begin at the end of the month.
	// For example: -1 for the last day of the month.
	// +optional
	DaysOfMonth []string `json:"days_of_month,omitempty"`

	// Depending on the location, the time range is displayed in local time.
	// +optional
	Location string `json:"location,omitempty"`

	// The months of the year in either numerical or the full calendar month.
	// For example: 1, may.
	// +optional
	Months []string `json:"months,omitempty"`

	// The time inclusive of the start and exclusive of the end time (in UTC if no location has been selected, otherwise local time).
	// +optional
	Times []*TimeRange `json:"times,omitempty"`

	// The day or range of days of the week.
	// For example: monday, thursday
	// +optional
	Weekdays []string `json:"weekdays,omitempty"`

	// The year or years for the interval.
	// For example: 2021
	// +optional
	Years []string `json:"years,omitempty"`
}

type TimeRange struct {
	// start time
	StartTime string `json:"start_time"`

	// end time
	EndTime string `json:"end_time"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaMuteTiming is the Schema for the GrafanaMuteTiming API
// +kubebuilder:printcolumn:name="Last resync",type="date",format="date-time",JSONPath=".status.lastResync",description=""
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
// +kubebuilder:resource:categories={grafana-operator}
type GrafanaMuteTiming struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaMuteTimingSpec `json:"spec"`
	Status GrafanaCommonStatus   `json:"status,omitempty"`
}

var _ CommonResource = (*GrafanaMuteTiming)(nil)

func (in *GrafanaMuteTiming) MatchLabels() *metav1.LabelSelector {
	return in.Spec.InstanceSelector
}

func (in *GrafanaMuteTiming) MatchNamespace() string {
	return in.Namespace
}

func (in *GrafanaMuteTiming) AllowCrossNamespace() bool {
	return in.Spec.AllowCrossNamespaceImport
}

func (in *GrafanaMuteTiming) NamespacedResource() NamespacedResource {
	return NewNamespacedResource(in.Namespace, in.Name, in.Spec.Name)
}

func (in *GrafanaMuteTiming) CommonStatus() *GrafanaCommonStatus {
	return &in.Status
}

//+kubebuilder:object:root=true

// GrafanaMuteTimingList contains a list of GrafanaMuteTiming
type GrafanaMuteTimingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaMuteTiming `json:"items"`
}

func (in *GrafanaMuteTimingList) Exists(namespace, name string) bool {
	for _, muteTiming := range in.Items {
		if muteTiming.Namespace == namespace && muteTiming.Name == name {
			return true
		}
	}
	return false
}

func init() {
	SchemeBuilder.Register(&GrafanaMuteTiming{}, &GrafanaMuteTimingList{})
}
