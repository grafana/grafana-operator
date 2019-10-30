package v1alpha1

import (
	"github.com/ghodss/yaml"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	GrafanaDataSourceKind = "GrafanaDataSource"
	DatasourcesApiVersion = 1
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GrafanaDataSourceStatus defines the observed state of GrafanaDataSource
// +k8s:openapi-gen=true
type GrafanaDataSourceStatus struct {
	Phase      int    `json:"phase"`
	ID         uint   `json:"id"`
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

type GrafanaDataSourceSpec struct {
	GrafanaRef string                      `json:"grafanaRef"`
	DataSource GrafanaDataSourceSpecFields `json:"datasource"`
}

type GrafanaDataSourceSpecFields struct {
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
	TlsAuth                 bool   `json:"tlsAuth,omitempty"`
	TlsAuthWithCACert       bool   `json:"tlsAuthWithCACert,omitempty"`
	TlsSkipVerify           bool   `json:"tlsSkipVerify,omitempty"`
	GraphiteVersion         string `json:"graphiteVersion,omitempty"`
	TimeInterval            string `json:"timeInterval,omitempty"`
	EsVersion               int    `json:"esVersion,omitempty"`
	TimeField               string `json:"timeField,omitempty"`
	Interval                string `json:"interval,omitempty"`
	LogMessageField         string `json:"logMessageField,omitempty"`
	LogLevelField           string `json:"logLevelField,omitempty"`
	AuthType                string `json:"authType,omitempty"`
	AssumeRoleArn           string `json:"assumeRoleArn,omitempty"`
	DefaultRegion           string `json:"defaultRegion,omitempty"`
	CustomMetricsNamespaces string `json:"customMetricsNamespaces,omitempty"`
	TsdbVersion             string `json:"tsdbVersion,omitempty"`
	TsdbResolution          string `json:"tsdbResolution,omitempty"`
	Sslmode                 string `json:"sslmode,omitempty"`
	Encrypt                 string `json:"encrypt,omitempty"`
	PostgresVersion         int    `json:"postgresVersion,omitempty"`
	Timescaledb             bool   `json:"timescaledb,omitempty"`
	MaxOpenConns            int    `json:"maxOpenConns,omitempty"`
	MaxIdleConns            int    `json:"maxIdleConns,omitempty"`
	ConnMaxLifetime         int    `json:"connMaxLifetime,omitempty"`
}

// The most common secure json options
// See https://grafana.com/docs/administration/provisioning/#datasources
type GrafanaDataSourceSecureJsonData struct {
	TlsCaCert         string `json:"tlsCACert,omitempty"`
	TlsClientCert     string `json:"tlsClientCert,omitempty"`
	TlsClientKey      string `json:"tlsClientKey,omitempty"`
	Password          string `json:"password,omitempty"`
	BasicAuthPassword string `json:"basicAuthPassword,omitempty"`
	AccessKey         string `json:"accessKey,omitempty"`
	SecretKey         string `json:"secretKey,omitempty"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDataSource{}, &GrafanaDataSourceList{})
}

func ParseDataSource(cr *GrafanaDataSource) (string, error) {
	datasources := struct {
		ApiVersion int                           `json:"apiVersion"`
		Spec       []GrafanaDataSourceSpecFields `json:"datasources"`
	}{
		ApiVersion: DatasourcesApiVersion,
		Spec:       []GrafanaDataSourceSpecFields{cr.Spec.DataSource},
	}

	bytes, err := yaml.Marshal(datasources)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
