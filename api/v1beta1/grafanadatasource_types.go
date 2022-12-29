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
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GrafanaDatasourceInternal struct {
	UID           string `json:"uid,omitempty"`
	Name          string `json:"name,omitempty"`
	Type          string `json:"type,omitempty"`
	URL           string `json:"url,omitempty"`
	Access        string `json:"access,omitempty"`
	Database      string `json:"database,omitempty"`
	User          string `json:"user,omitempty"`
	OrgID         *int64 `json:"orgId,omitempty"`
	IsDefault     *bool  `json:"isDefault,omitempty"`
	BasicAuth     *bool  `json:"basicAuth,omitempty"`
	BasicAuthUser string `json:"basicAuthUser,omitempty"`
	Editable      *bool  `json:"editable,omitempty"`

	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// +optional
	JSONData json.RawMessage `json:"jsonData,omitempty"`

	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	// +optional
	SecureJSONData json.RawMessage `json:"secureJsonData,omitempty"`
}

// GrafanaDatasourceSpec defines the desired state of GrafanaDatasource
type GrafanaDatasourceSpec struct {
	Datasource *GrafanaDatasourceInternal `json:"datasource,omitempty"`

	// selects Grafana instances for import
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`

	// plugins
	// +optional
	Plugins PluginList `json:"plugins,omitempty"`

	// secrets used for variable expansion
	// +optional
	Secrets []string `json:"secrets,omitempty"`

	// how often the datasource is refreshed, defaults to 24h if not set
	// +optional
	ResyncPeriod string `json:"resyncPeriod,omitempty"`
}

// GrafanaDatasourceStatus defines the observed state of GrafanaDatasource
type GrafanaDatasourceStatus struct {
	Hash string `json:"hash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GrafanaDatasource is the Schema for the grafanadatasources API
type GrafanaDatasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDatasourceSpec   `json:"spec,omitempty"`
	Status GrafanaDatasourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GrafanaDatasourceList contains a list of GrafanaDatasource
type GrafanaDatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDatasource `json:"items"`
}

func (in *GrafanaDatasource) Hash() string {
	hash := sha256.New()

	if in.Spec.Datasource != nil {
		hash.Write([]byte(in.Spec.Datasource.Name))
		hash.Write([]byte(in.Spec.Datasource.Access))
		hash.Write([]byte(in.Spec.Datasource.BasicAuthUser))
		hash.Write(in.Spec.Datasource.JSONData)
		hash.Write(in.Spec.Datasource.SecureJSONData)
		hash.Write([]byte(in.Spec.Datasource.Database))
		hash.Write([]byte(in.Spec.Datasource.Type))
		hash.Write([]byte(in.Spec.Datasource.User))
		hash.Write([]byte(in.Spec.Datasource.URL))

		if in.Spec.Datasource.BasicAuth != nil && *in.Spec.Datasource.BasicAuth {
			hash.Write([]byte("_"))
		}

		if in.Spec.Datasource.IsDefault != nil && *in.Spec.Datasource.IsDefault {
			hash.Write([]byte("_"))
		}

		if in.Spec.Datasource.OrgID != nil {
			hash.Write([]byte(fmt.Sprint(*in.Spec.Datasource.OrgID)))
		}

		if in.Spec.Datasource.Editable != nil && *in.Spec.Datasource.Editable {
			hash.Write([]byte("_"))
		}
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (in *GrafanaDatasource) GetResyncPeriod() time.Duration {
	if in.Spec.ResyncPeriod == "" {
		in.Spec.ResyncPeriod = DefaultResyncPeriod
		return in.GetResyncPeriod()
	}

	duration, err := time.ParseDuration(in.Spec.ResyncPeriod)
	if err != nil {
		in.Spec.ResyncPeriod = DefaultResyncPeriod
		return in.GetResyncPeriod()
	}

	return duration
}

func (in *GrafanaDatasource) Unchanged() bool {
	return in.Hash() == in.Status.Hash
}

func (in *GrafanaDatasourceList) Find(namespace string, name string) *GrafanaDatasource {
	for _, datasource := range in.Items {
		if datasource.Namespace == namespace && datasource.Name == name {
			return &datasource
		}
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&GrafanaDatasource{}, &GrafanaDatasourceList{})
}
