package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GrafanaResourceDatasource is used to set the datasource name of any templated datasources in
// content definitions (e.g., dashboard JSON).
type GrafanaContentDatasource struct {
	InputName      string `json:"inputName"`
	DatasourceName string `json:"datasourceName"`
}

type GrafanaContentEnv struct {
	Name string `json:"name"`
	// Inline env value
	// +optional
	Value string `json:"value,omitempty"`
	// Reference on value source, might be the reference on a secret or config map
	// +optional
	ValueFrom GrafanaContentEnvFromSource `json:"valueFrom,omitempty"`
}

type GrafanaContentEnvFromSource struct {
	// Selects a key of a ConfigMap.
	// +optional
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef,omitempty"`
	// Selects a key of a Secret.
	// +optional
	SecretKeyRef *corev1.SecretKeySelector `json:"secretKeyRef,omitempty"`
}

type GrafanaContentURLBasicAuth struct {
	Username *corev1.SecretKeySelector `json:"username,omitempty"`
	Password *corev1.SecretKeySelector `json:"password,omitempty"`
}

type GrafanaContentURLAuthorization struct {
	BasicAuth *GrafanaContentURLBasicAuth `json:"basicAuth,omitempty"`
}

type JsonnetProjectBuild struct {
	JPath              []string `json:"jPath,omitempty"`
	FileName           string   `json:"fileName"`
	GzipJsonnetProject []byte   `json:"gzipJsonnetProject"`
}

// GrafanaComContentReference is a reference to content hosted on grafana.com
type GrafanaComContentReference struct {
	ID       int  `json:"id"`
	Revision *int `json:"revision,omitempty"`
}

type GrafanaContentSpec struct {
	// Manually specify the uid, overwrites uids already present in the json model.
	// Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40.
	// +optional
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="spec.uid is immutable"
	// +kubebuilder:validation:MaxLength=40
	// +kubebuilder:validation:Pattern="^[a-zA-Z0-9-_]+$"
	CustomUID string `json:"uid,omitempty"`

	// model json
	// +optional
	JSON string `json:"json,omitempty"`

	// GzipJson the model's JSON compressed with Gzip. Base64-encoded when in YAML.
	// +optional
	GzipJSON []byte `json:"gzipJson,omitempty"`

	// model url
	// +optional
	URL string `json:"url,omitempty"`

	// authorization options for model from url
	// +optional
	URLAuthorization *GrafanaContentURLAuthorization `json:"urlAuthorization,omitempty"`

	// Jsonnet
	// +optional
	Jsonnet string `json:"jsonnet,omitempty"`

	// Jsonnet project build
	JsonnetProjectBuild *JsonnetProjectBuild `json:"jsonnetLib,omitempty"`

	// model from configmap
	// +optional
	ConfigMapRef *corev1.ConfigMapKeySelector `json:"configMapRef,omitempty"`

	// grafana.com/dashboards
	// +optional
	GrafanaCom *GrafanaComContentReference `json:"grafanaCom,omitempty"`

	// Cache duration for models fetched from URLs
	// +optional
	ContentCacheDuration metav1.Duration `json:"contentCacheDuration,omitempty"`

	// maps required data sources to existing ones
	// +optional
	Datasources []GrafanaContentDatasource `json:"datasources,omitempty"`

	// environments variables as a map
	// +optional
	Envs []GrafanaContentEnv `json:"envs,omitempty"`

	// environments variables from secrets or config maps
	// +optional
	EnvsFrom []GrafanaContentEnvFromSource `json:"envFrom,omitempty"`
}

type GrafanaContentStatus struct {
	ContentCache     []byte      `json:"contentCache,omitempty"`
	ContentTimestamp metav1.Time `json:"contentTimestamp,omitempty"`
	ContentURL       string      `json:"contentUrl,omitempty"`
	Hash             string      `json:"hash,omitempty"`
	UID              string      `json:"uid,omitempty"`
}

// Common interface for any resource that embeds or references Grafana-native model content.
// +kubebuilder:object:generate=false
type GrafanaContentResource interface {
	client.Object
	GrafanaContentSpec() *GrafanaContentSpec
	GrafanaContentStatus() *GrafanaContentStatus
}
