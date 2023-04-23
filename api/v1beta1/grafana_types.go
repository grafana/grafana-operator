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
	"strings"

	"github.com/grafana-operator/grafana-operator/v5/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OperatorStageName string

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
	// Service sets how the service object should look like with your grafana instance, contains a number of defaults.
	Service *ServiceV1 `json:"service,omitempty"`
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
}

type External struct {
	// URL of the external grafana instance you want to manage.
	URL string `json:"url"`
	// The API key to talk to the external grafana instance, you need to define ether apiKey or adminUser/adminPassword.
	ApiKey *v1.SecretKeySelector `json:"apiKey,omitempty"`
	// AdminUser key to talk to the external grafana instance.
	AdminUser *v1.SecretKeySelector `json:"adminUser,omitempty"`
	// AdminPassword key to talk to the external grafana instance.
	AdminPassword *v1.SecretKeySelector `json:"adminPassword,omitempty"`
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
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	AdminUrl   string             `json:"adminUrl,omitempty"`
	Plugins    PluginList         `json:"plugins,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description=""
//+kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""

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

func (in *Grafana) IsInternal() bool {
	return in.Spec.External == nil
}

func (in *Grafana) IsExternal() bool {
	return in.Spec.External != nil
}

func (in *Grafana) Ready() bool {
	cond := in.GetReadyCondition()
	return cond != nil && cond.Status == metav1.ConditionTrue
}

func (in *Grafana) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

func (in *Grafana) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *Grafana) GetReadyCondition() *metav1.Condition {
	return api.GetReadyCondition(in)
}

func (in *Grafana) SetCondition(condition metav1.Condition) bool {
	return api.SetCondition(in, condition)
}

func (in *Grafana) SetReadyCondition(status metav1.ConditionStatus, reason string, message string) bool {
	return api.SetReadyCondition(in, status, reason, message)
}

func InstanceKeyFor(grafana *Grafana) string {
	return client.ObjectKeyFromObject(grafana).String()
}

func NamespacedNameFor(instanceKey string) client.ObjectKey {
	split := strings.Split(instanceKey, string(types.Separator))
	if len(split) != 2 {
		panic(fmt.Sprintf("instanceKey %s provided didn't contain exactly 1 %v", instanceKey, types.Separator))
	}

	return client.ObjectKey{
		Namespace: split[0],
		Name:      split[1],
	}
}
