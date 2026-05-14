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

// Package v1beta1 contains API Schema definitions for the grafana v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=grafana.integreatly.org
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "grafana.integreatly.org", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&GrafanaAlertRuleGroup{}, &GrafanaAlertRuleGroupList{},
		&GrafanaContactPoint{}, &GrafanaContactPointList{},
		&GrafanaDashboard{}, &GrafanaDashboardList{},
		&GrafanaDatasource{}, &GrafanaDatasourceList{},
		&GrafanaFolder{}, &GrafanaFolderList{},
		&GrafanaLibraryPanel{}, &GrafanaLibraryPanelList{},
		&GrafanaManifest{}, &GrafanaManifestList{},
		&GrafanaMuteTiming{}, &GrafanaMuteTimingList{},
		&GrafanaNotificationPolicyRoute{}, &GrafanaNotificationPolicyRouteList{},
		&GrafanaNotificationPolicy{}, &GrafanaNotificationPolicyList{},
		&GrafanaNotificationTemplate{}, &GrafanaNotificationTemplateList{},
		&GrafanaPrometheusRule{}, &GrafanaPrometheusRuleList{},
		&GrafanaServiceAccount{}, &GrafanaServiceAccountList{},
		&Grafana{}, &GrafanaList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)

	return nil
}
