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

// GrafanaContentVariable overrides the default (current) value of a named template
// variable (templating.list[]) in the resolved dashboard model.
type GrafanaContentVariable struct {
	// Name of the template variable to override, matching templating.list[].name in the model.
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Value is the new default value. For datasource variables this is the datasource UID or
	// name; for query, custom, constant and textbox variables it is the selected value.
	// Multi-value (multi-select) variables are collapsed to this single value.
	Value string `json:"value"`
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

// GrafanaContentOCI references a dashboard JSON file inside an OCI artifact in a container registry.
// Bytes are fetched at reconcile time and never persisted to etcd; recommended for dashboards that
// exceed the etcd object-size limit (~1 MiB).
type GrafanaContentOCI struct {
	// Reference is the full OCI artifact reference including a tag or digest,
	// e.g. "ghcr.io/team/dashboards:v1.4.7" or
	// "ghcr.io/team/dashboards@sha256:abc123...". Prefer a digest for
	// reproducible deployments.
	// +kubebuilder:validation:MinLength=3
	// +kubebuilder:validation:MaxLength=512
	// +kubebuilder:validation:Pattern=`^[^:@]+(:[^:@/]+|@sha256:[a-fA-F0-9]{64})$`
	Reference string `json:"reference"`

	// Path is the path of the file to extract from the artifact.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=512
	Path string `json:"path"`

	// PullSecretRef references a kubernetes.io/dockerconfigjson Secret in the same namespace as the CR.
	// If omitted, anonymous pull is attempted.
	// +optional
	PullSecretRef *corev1.LocalObjectReference `json:"pullSecretRef,omitempty"`

	// InsecurePlainHTTP switches the registry connection to plain HTTP (non-TLS) instead of HTTPS.
	// Intended for in-cluster or test registries; HTTPS registries with self-signed
	// certificates are not supported. Default false.
	// +optional
	InsecurePlainHTTP bool `json:"insecurePlainHTTP,omitempty"`
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
	// +kubebuilder:validation:Pattern=`^https?://.+$`
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

	// model from an OCI artifact (e.g. ghcr.io/team/dashboards:v1)
	// +optional
	OCI *GrafanaContentOCI `json:"oci,omitempty"`

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

// GrafanaContentVariableOverrider is implemented by content resources whose model carries template
// variables (templating.list[]), so that only those resources take part in variable overriding.
// +kubebuilder:object:generate=false
type GrafanaContentVariableOverrider interface {
	ContentVariables() []GrafanaContentVariable
}
