package v1alpha1

import (
	v12 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// LokiSpec defines the desired state of Loki
type LokiSpec struct {
	Config       LokiConfig        `json:"config"`
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Deployment   *LokiDeployment   `json"deployment,omitempty"`
	DataStorage  *LokiDataStorage  `json"dataStorage,omitempty"`
	// When set, refer to unmamnaged Loki instance and do not create a managed one
	External  *LokiExternal `json:"external,omitempty"`
	Service   *LokiService  `json:"service,omitempty"`
	Ingress   *LokiIngress  `json:"ingress,omitempty"`
	Route     *LokiRoute    `json:"route,omitempty"`
	BaseImage string        `json:"baseImage,omitempty"`
}

type LokiDeployment struct {
	Annotations                   map[string]string      `json:"annotations,omitempty"`
	Labels                        map[string]string      `json:"labels,omitempty"`
	Replicas                      int32                  `json:"replicas"`
	NodeSelector                  map[string]string      `json:"nodeSelector,omitempty"`
	Tolerations                   []v1.Toleration        `json:"tolerations,omitempty"`
	Affinity                      *v1.Affinity           `json:"affinity,omitempty"`
	SecurityContext               *v1.PodSecurityContext `json:"securityContext,omitempty"`
	ContainerSecurityContext      *v1.SecurityContext    `json:"containerSecurityContext,omitempty"`
	TerminationGracePeriodSeconds int64                  `json:"terminationGracePeriodSeconds"`
}

type LokiConfig struct {
	Server *LokiConfigServer `json:"server,omitempty" ini:"server,omitempty"`
}
type LokiConfigServer struct {
	HttpAddr   string `json:"http_addr,omitempty" ini:"http_addr,omitempty"`
	HttpPort   string `json:"http_port,omitempty" ini:"http_port,omitempty"`
	HttpPrefix string `json:"http_prefix,omitempty" ini:"http_prefix"`
}

type LokiService struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Type        v1.ServiceType    `json:"type,omitempty"`
	Ports       []v1.ServicePort  `json:"ports,omitempty"`
	ClusterIP   string            `json:"clusterIP,omitempty"`
}

type LokiIngress struct {
	Annotations   map[string]string      `json:"annotations,omitempty"`
	Hostname      string                 `json:"hostname,omitempty"`
	Labels        map[string]string      `json:"Labels,omitempty"`
	Path          string                 `json:"labels,omitempty"`
	Enabled       bool                   `json:"enabled,omitempty"`
	TLSEnabled    bool                   `json:"tlsEnabled,omitempty"`
	TLSSecretName string                 `json:"tlsSecretName,omitempty"`
	TargetPort    string                 `json:"targetPort,omitempty"`
	Termination   v12.TLSTerminationType `json:"termination,omitempty"`
}

// LokiDataStorage provides a means to configure the grafana data storage
type LokiDataStorage struct {
	Annotations map[string]string               `json:"annotations,omitempty"`
	Labels      map[string]string               `json:"labels,omitempty"`
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	Size        resource.Quantity               `json:"size"`
	Class       string                          `json:"class"`
}

type LokiRoute struct {
}

type LokiExternal struct {
	Url string `json:"url,omitempty"`
}

// LokiStatus defines the observed state of Loki
type LokiStatus struct {
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
	Url     string      `json:"url,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Loki is the Schema for the lokis API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=lokis,scope=Namespaced
type Loki struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LokiSpec   `json:"spec,omitempty"`
	Status LokiStatus `json:"status,omitempty"`
}

type LokiRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	UID       string `json:"uid"`
	LokiURL   string `json:"lokiurl"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// LokiList contains a list of Loki
type LokiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Loki `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Loki{}, &LokiList{})
}
