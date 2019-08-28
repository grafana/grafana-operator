package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const GrafanaDataSourceKind = "GrafanaDataSource"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDataSourceSpec defines the desired state of GrafanaDataSource
// +k8s:openapi-gen=true
type GrafanaDataSourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Datasources []GrafanaDataSourceFields `json:"datasources"`
	Name        string                    `json:"name"`
}

// GrafanaDataSourceStatus defines the observed state of GrafanaDataSource
// +k8s:openapi-gen=true
type GrafanaDataSourceStatus struct {
	Phase      int    `json:"phase"`
	LastConfig string `json:"lastConfig"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDataSource is the Schema for the grafanadatasources API
// +k8s:openapi-gen=true
type GrafanaDataSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaDataSourceSpec   `json:"spec,omitempty"`
	Status GrafanaDataSourceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaDataSourceList contains a list of GrafanaDataSource
type GrafanaDataSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GrafanaDataSource `json:"items"`
}

type GrafanaDataSourceFields struct {
	Name              string                          `json:"name"`
	Type              string                          `json:"type"`
	Access            string                          `json:"access"`
	OrgId             int                             `json:"orgId,omitempty"`
	Url               string                          `json:"url"`
	Password          string                          `json:"password,omitempty"`
	User              string                          `json:"user,omitempty"`
	Database          string                          `json:"database,omitempty"`
	BasicAuth         bool                            `json:"basicAuth,omitempty"`
	BasicAuthUser     string                          `json:"basicAuthUser,omitempty"`
	BasicAuthPassword string                          `json:"basicAuthPassword,omitempty"`
	WithCredentials   bool                            `json:"withCredentials,omitempty"`
	IsDefault         bool                            `json:"isDefault,omitempty"`
	JsonData          GrafanaDataSourceJsonData       `json:"jsonData,omitempty"`
	SecureJsonData    GrafanaDataSourceSecureJsonData `json:"secureJsonData,omitempty"`
	Version           int                             `json:"version,omitempty"`
	Editable          bool                            `json:"editable,omitempty"`
}

// The most common json options
// See https://grafana.com/docs/administration/provisioning/#datasources
type GrafanaDataSourceJsonData struct {
	TlsAuth                 bool   `json:"tlsAuth"`
	TlsAuthWithCACert       bool   `json:"tlsAuthWithCACert"`
	TlsSkipVerify           bool   `json:"tlsSkipVerify"`
	GraphiteVersion         string `json:"graphiteVersion"`
	TimeInterval            string `json:"timeInterval"`
	EsVersion               int    `json:"esVersion"`
	TimeField               string `json:"timeField"`
	Interval                string `json:"interval"`
	LogMessageField         string `json:"logMessageField"`
	LogLevelField           string `json:"logLevelField"`
	AuthType                string `json:"authType"`
	AssumeRoleArn           string `json:"assumeRoleArn"`
	DefaultRegion           string `json:"defaultRegion"`
	CustomMetricsNamespaces string `json:"customMetricsNamespaces"`
	TsdbVersion             string `json:"tsdbVersion"`
	TsdbResolution          string `json:"tsdbResolution"`
	Sslmode                 string `json:"sslmode"`
	Encrypt                 string `json:"encrypt"`
	PostgresVersion         int    `json:"postgresVersion"`
	Timescaledb             bool   `json:"timescaledb"`
	MaxOpenConns            int    `json:"maxOpenConns"`
	MaxIdleConns            int    `json:"maxIdleConns"`
	ConnMaxLifetime         int    `json:"connMaxLifetime"`
}

// The most common secure json options
// See https://grafana.com/docs/administration/provisioning/#datasources
type GrafanaDataSourceSecureJsonData struct {
	TlsCaCert         string `json:"tlsCACert"`
	TlsClientCert     string `json:"tlsClientCert"`
	TlsClientKey      string `json:"tlsClientKey"`
	Password          string `json:"password"`
	BasicAuthPassword string `json:"basicAuthPassword"`
	AccessKey         string `json:"accessKey"`
	SecretKey         string `json:"secretKey"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDataSource{}, &GrafanaDataSourceList{})
}
