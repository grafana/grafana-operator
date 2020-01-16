package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
type KeystoneConnectorConfig struct {
	Cloud               string `json:"cloud"`
	Domain              string `json:"domain"`
	Host                string `json:"host"`
	AdminUsername       string `json:"adminUsername"`
	AdminPassword       string `json:"adminPassword"`
	AdminUserDomainName string `json:"adminUserDomain"`
	AdminProject        string `json:"adminProject"`
	AdminDomain         string `json:"adminDomain"`
	Prompt              string `json:"prompt"`
	//AuthScope            AuthScope `json:"authScope,omitempty"`
	IncludeRolesInGroups *bool  `json:"includeRolesInGroups,omitempty"`
	RoleNameFormat       string `json:"roleNameFormat,omitempty"`
	GroupNameFormat      string `json:"groupNameFormat,omitempty"`
}

type AuthScope struct {
	ProjectID   string `json:"projectID,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	DomainID    string `json:"domainID,omitempty"`
	DomainName  string `json:"domainName,omitempty"`
}
*/

type GrafanaProxyConnector struct {
	Type   string            `json:"type"`
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

// Web is the config format for the HTTP server.
type Web struct {
	HTTP           string   `json:"http"`
	HTTPS          string   `json:"https"`
	TLSCert        string   `json:"tlsCert"`
	TLSKey         string   `json:"tlsKey"`
	AllowedOrigins []string `json:"allowedOrigins"`
}

// OAuth2 describes enabled OAuth2 extensions.
type OAuth2 struct {
	ResponseTypes []string `json:"responseTypes"`
	// If specified, do not prompt the user to approve client authorization. The
	// act of logging in implies authorization.
	SkipApprovalScreen bool `json:"skipApprovalScreen"`
	// If specified, show the connector selection screen even if there's only one
	AlwaysShowLoginScreen bool `json:"alwaysShowLoginScreen"`
}

// Storage holds app's storage configuration.
type Storage struct {
	Type   string        `json:"type"`
	Config StorageConfig `json:"config"`
}

type StorageConfig struct {
	InCluster bool `json:"inCluster"`
}

type Logger struct {
	// Level sets logging level severity.
	Level string `json:"level"`
	// Format specifies the format to be used for logging.
	Format string `json:"format"`
}

// Expiry holds configuration for the validity period of components.
type Expiry struct {
	// SigningKeys defines the duration of time after which the SigningKeys will be rotated.
	SigningKeys string `json:"signingKeys"`
	// IdTokens defines the duration of time for which the IdTokens will be valid.
	IDTokens string `json:"idTokens"`
	// AuthRequests defines the duration of time for which the AuthRequests will be valid.
	AuthRequests string `json:"authRequests"`
}

type WebConfig struct {
	Dir     string `json:"dir"`
	LogoURL string `json:"logonUrl"`
	Issuer  string `json:"issuer"`
	Theme   string `json:"theme"`
}

// Client represents an OAuth2 client.
//
// For further reading see:
//   * Trusted peers: https://developers.google.com/identity/protocols/CrossClientAuth
//   * Public clients: https://developers.google.com/api-client-library/python/auth/installed-app
type Client struct {
	ID           string   `json:"id" yaml:"id"`
	Secret       string   `json:"secret" yaml:"secret"`
	RedirectURIs []string `json:"redirectURIs" yaml:"redirectURIs"`
	TrustedPeers []string `json:"trustedPeers" yaml:"trustedPeers"`
	Public       bool     `json:"public" yaml:"public"`
	Name         string   `json:"name" yaml:"name"`
	LogoURL      string   `json:"logoURL" yaml:"logoURL"`
}

// GrafanaProxyConfig provides a auth proxy
type GrafanaProxyConfig struct {
	HostName         string                  `json:"hostName"`
	Issuer           string                  `json:"issuer"`
	Storage          Storage                 `json:"storage"`
	Web              Web                     `json:"web"`
	OAuth2           OAuth2                  `json:"oauth2"`
	Expiry           Expiry                  `json:"expiry"`
	Logger           Logger                  `json:"logger"`
	Frontend         WebConfig               `json:"frontend"`
	Connectors       []GrafanaProxyConnector `json:"connectors"`
	StaticClients    []Client                `json:"staticClients"`
	EnablePasswordDB bool                    `json:"enablePasswordDB"`
	Enabled          bool                    `json:"enabled,omitempty"`
	ClientSecret     string                  `json:"client_secret"`
	ClientID         string                  `json:"client_id"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaProxySpec defines the desired state of GrafanaProxy
// +k8s:openapi-gen=true
type GrafanaProxySpec struct {
	Config     GrafanaProxyConfig       `json:"config"`
	Resources  *v1.ResourceRequirements `json:"resources,omitempty"`
	Deployment *GrafanaDeployment       `json:"deployment,omitempty"`
}

// GrafanaProxyStatus defines the observed state of GrafanaProxy
// +k8s:openapi-gen=true
type GrafanaProxyStatus struct {
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaProxy is the Schema for the grafanaproxies API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=grafanaproxies,scope=Namespaced
type GrafanaProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaProxySpec   `json:"spec,omitempty"`
	Status GrafanaProxyStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaProxyList contains a list of GrafanaProxy
type GrafanaProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GrafanaProxy{}, &GrafanaProxyList{})
}
