package v1alpha1

import (
	"fmt"
	"strings"

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
	Phase   StatusPhase `json:"phase"`
	Message string      `json:"message"`
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
	Uid               string                          `json:"uid,omitempty"`
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
	OauthPassThru           bool   `json:"oauthPassThru,omitempty"`
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
	//  Useful fields for clickhouse datasource
	//  See https://github.com/Vertamedia/clickhouse-grafana/tree/master/dist/README.md#configure-the-datasource-with-provisioning
	//  See https://github.com/Vertamedia/clickhouse-grafana/tree/master/src/datasource.ts#L44
	AddCorsHeader               bool   `json:"addCorsHeader,omitempty"`
	DefaultDatabase             string `json:"defaultDatabase,omitempty"`
	UsePOST                     bool   `json:"usePOST,omitempty"`
	UseYandexCloudAuthorization bool   `json:"useYandexCloudAuthorization,omitempty"`
	XHeaderUser                 string `json:"xHeaderUser,omitempty"`
	XHeaderKey                  string `json:"xHeaderKey,omitempty"`
	// Custom HTTP headers for datasources
	// See https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources
	HTTPHeaderName1 string `json:"httpHeaderName1,omitempty"`
	HTTPHeaderName2 string `json:"httpHeaderName2,omitempty"`
	HTTPHeaderName3 string `json:"httpHeaderName3,omitempty"`
	HTTPHeaderName4 string `json:"httpHeaderName4,omitempty"`
	HTTPHeaderName5 string `json:"httpHeaderName5,omitempty"`
	HTTPHeaderName6 string `json:"httpHeaderName6,omitempty"`
	HTTPHeaderName7 string `json:"httpHeaderName7,omitempty"`
	HTTPHeaderName8 string `json:"httpHeaderName8,omitempty"`
	HTTPHeaderName9 string `json:"httpHeaderName9,omitempty"`
	// Fields for Stackdriver data sources
	TokenUri           string `json:"tokenUri,omitempty"`
	ClientEmail        string `json:"clientEmail,omitempty"`
	AuthenticationType string `json:"authenticationType,omitempty"`
	DefaultProject     string `json:"defaultProject,omitempty"`
	// Fields for Azure data sources
	AppInsightsAppId             string `json:"appInsightsAppId,omitempty"`
	AzureLogAnalyticsSameAs      string `json:"azureLogAnalyticsSameAs,omitempty"`
	ClientId                     string `json:"clientId,omitempty"`
	CloudName                    string `json:"cloudName,omitempty"`
	LogAnalyticsDefaultWorkspace string `json:"logAnalyticsDefaultWorkspace,omitempty"`
	LogAnalyticsClientId         string `json:"logAnalyticsClientId,omitempty"`
	LogAnalyticsSubscriptionId   string `json:"logAnalyticsSubscriptionId,omitempty"`
	LogAnalyticsTenantId         string `json:"logAnalyticsTenantId,omitempty"`
	SubscriptionId               string `json:"subscriptionId,omitempty"`
	TenantId                     string `json:"tenantId,omitempty"`
	// Fields for InfluxDB data sources
	HTTPMode      string `json:"httpMode,omitempty"`
	Version       string `json:"version,omitempty"`
	Organization  string `json:"organization,omitempty"`
	DefaultBucket string `json:"defaultBucket,omitempty"`
	// Fields for Loki data sources
	MaxLines      int                                  `json:"maxLines,omitempty"`
	DerivedFields []GrafanaDataSourceJsonDerivedFields `json:"derivedFields,omitempty"`
	// Fields for Prometheus data sources
	CustomQueryParameters string `json:"customQueryParameters,omitempty"`
	HTTPMethod            string `json:"httpMethod,omitempty"`
}

type GrafanaDataSourceJsonDerivedFields struct {
	DatasourceUid string `json:"datasourceUid,omitempty"`
	MatcherRegex  string `json:"matcherRegex,omitempty"`
	Name          string `json:"name,omitempty"`
	Url           string `json:"url,omitempty"`
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
	// Custom HTTP headers for datasources
	// See https://grafana.com/docs/grafana/latest/administration/provisioning/#datasources
	HTTPHeaderValue1 string `json:"httpHeaderValue1,omitempty"`
	HTTPHeaderValue2 string `json:"httpHeaderValue2,omitempty"`
	HTTPHeaderValue3 string `json:"httpHeaderValue3,omitempty"`
	HTTPHeaderValue4 string `json:"httpHeaderValue4,omitempty"`
	HTTPHeaderValue5 string `json:"httpHeaderValue5,omitempty"`
	HTTPHeaderValue6 string `json:"httpHeaderValue6,omitempty"`
	HTTPHeaderValue7 string `json:"httpHeaderValue7,omitempty"`
	HTTPHeaderValue8 string `json:"httpHeaderValue8,omitempty"`
	HTTPHeaderValue9 string `json:"httpHeaderValue9,omitempty"`
	// Fields for Stackdriver data sources
	PrivateKey string `json:"privateKey,omitempty"`
	// Fields for Azure data sources
	ClientSecret             string `json:"clientSecret,omitempty"`
	AppInsightsApiKey        string `json:"appInsightsApiKey,omitempty"`
	LogAnalyticsClientSecret string `json:"logAnalyticsClientSecret,omitempty"`
	// Fields for InfluxDB data sources
	Token string `json:"token,omitempty"`
}

func init() {
	SchemeBuilder.Register(&GrafanaDataSource{}, &GrafanaDataSourceList{})
}

// return a unique per namespace key of the datasource
func (ds *GrafanaDataSource) Filename() string {
	return fmt.Sprintf("%v_%v.yaml", ds.Namespace, strings.ToLower(ds.Name))
}
