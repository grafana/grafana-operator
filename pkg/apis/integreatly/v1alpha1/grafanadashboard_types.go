package v1alpha1

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GrafanaDashboardKind = "GrafanaDashboard"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDashboardSpec defines the desired state of GrafanaDashboard
type GrafanaDashboardSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	Json             string                       `json:"json"`
	Jsonnet          string                       `json:"jsonnet"`
	Plugins          PluginList                   `json:"plugins,omitempty"`
	Url              string                       `json:"url,omitempty"`
	ConfigMapRef     *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`
	Datasources      []GrafanaDashboardDatasource `json:"datasources,omitempty"`
	CustomFolderName string                       `json:"customFolderName,omitempty"`
}

type GrafanaDashboardDatasource struct {
	InputName      string `json:"inputName"`
	DatasourceName string `json:"datasourceName"`
}

// Used to keep a dashboard reference without having access to the dashboard
// struct itself
type GrafanaDashboardRef struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	UID        string `json:"uid"`
	Hash       string `json:"hash"`
	FolderId   *int64 `json:"folderId"`
	FolderName string `json:"folderName"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDashboard is the Schema for the grafanadashboards API
// +k8s:openapi-gen=true
type GrafanaDashboard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GrafanaDashboardSpec `json:"spec,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDashboardList contains a list of GrafanaDashboard
type GrafanaDashboardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDashboard `json:"items"`
}

type GrafanaDashboardStatusMessage struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDashboard{}, &GrafanaDashboardList{})
}

func (d *GrafanaDashboard) Hash() string {
	hash := sha256.New()

	for _, input := range d.Spec.Datasources {
		io.WriteString(hash, input.DatasourceName)
		io.WriteString(hash, input.InputName)
	}

	io.WriteString(hash, d.Spec.Json)
	io.WriteString(hash, d.Spec.Url)
	io.WriteString(hash, d.Spec.Jsonnet)
	io.WriteString(hash, d.Namespace)
	io.WriteString(hash, d.Spec.CustomFolderName)

	if d.Spec.ConfigMapRef != nil {
		io.WriteString(hash, d.Spec.ConfigMapRef.Name)
		io.WriteString(hash, d.Spec.ConfigMapRef.Key)
	}

	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (d *GrafanaDashboard) Parse(optional string) (map[string]interface{}, error) {
	var dashboardBytes = []byte(d.Spec.Json)
	if optional != "" {
		dashboardBytes = []byte(optional)
	}

	var parsed = make(map[string]interface{})
	err := json.Unmarshal(dashboardBytes, &parsed)
	return parsed, err
}

func (d *GrafanaDashboard) UID() string {
	content, err := d.Parse("")
	if err == nil {
		// Check if the user has defined an uid and if that's the
		// case, use that
		if content["uid"] != nil && content["uid"] != "" {
			return content["uid"].(string)
		}
	}

	// Use sha1 to keep the hash limit at 40 bytes which is what
	// Grafana allows for UIDs
	return fmt.Sprintf("%x", sha1.Sum([]byte(d.Namespace+d.Name)))
}
