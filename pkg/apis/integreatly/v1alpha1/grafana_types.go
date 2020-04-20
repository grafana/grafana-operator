package v1alpha1

import (
	v12 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusPhase string

var (
	NoPhase          StatusPhase
	PhaseReconciling StatusPhase = "reconciling"
	PhaseFailing     StatusPhase = "failing"
)

// GrafanaSpec defines the desired state of Grafana
// +k8s:openapi-gen=true
type GrafanaSpec struct {
	Config                 GrafanaConfig            `json:"config"`
	Containers             []v1.Container           `json:"containers,omitempty"`
	DashboardLabelSelector []*metav1.LabelSelector  `json:"dashboardLabelSelector,omitempty"`
	Ingress                *GrafanaIngress          `json:"ingress,omitempty"`
	Secrets                []string                 `json:"secrets,omitempty"`
	ConfigMaps             []string                 `json:"configMaps,omitempty"`
	Service                *GrafanaService          `json:"service,omitempty"`
	Deployment             *GrafanaDeployment       `json:"deployment,omitempty"`
	Resources              *v1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccount         *GrafanaServiceAccount   `json:"serviceAccount,omitempty"`
	Client                 *GrafanaClient           `json:"client,omitempty"`
	Compat                 *GrafanaCompat           `json:"compat"`
}

// Backwards compatibility switches
type GrafanaCompat struct {
	FixAnnotations bool `json:"fixAnnotations"`
	FixHeights     bool `json:"fixHeights"`
}

// Grafana API client settings
type GrafanaClient struct {
	TimeoutSeconds *int `json:"timeout"`
	PreferService  bool `json:"preferService"`
}

// GrafanaService provides a means to configure the service
type GrafanaService struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Type        v1.ServiceType    `json:"type,omitempty"`
	Ports       []v1.ServicePort  `json:"ports,omitempty"`
}

type GrafanaServiceAccount struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// GrafanaDeployment provides a means to configure the deployment
type GrafanaDeployment struct {
	Annotations     map[string]string      `json:"annotations,omitempty"`
	Labels          map[string]string      `json:"labels,omitempty"`
	Replicas        int32                  `json:"replicas"`
	NodeSelector    map[string]string      `json:"nodeSelector,omitempty"`
	Tolerations     []v1.Toleration        `json:"tolerations,omitempty"`
	Affinity        *v1.Affinity           `json:"affinity,omitempty"`
	SecurityContext *v1.PodSecurityContext `json:"securityContext,omitempty"`
	EnvFrom         []v1.EnvFromSource     `json:"envFrom,omitempty"`
}

// GrafanaIngress provides a means to configure the ingress created
type GrafanaIngress struct {
	Annotations   map[string]string      `json:"annotations,omitempty"`
	Hostname      string                 `json:"hostname,omitempty"`
	Labels        map[string]string      `json:"labels,omitempty"`
	Path          string                 `json:"path,omitempty"`
	Enabled       bool                   `json:"enabled,omitempty"`
	TLSEnabled    bool                   `json:"tlsEnabled,omitempty"`
	TLSSecretName string                 `json:"tlsSecretName,omitempty"`
	TargetPort    string                 `json:"targetPort,omitempty"`
	Termination   v12.TLSTerminationType `json:"termination,omitempty"`
}

// GrafanaConfig is the configuration for grafana
type GrafanaConfig struct {
	Paths                         *GrafanaConfigPaths                         `json:"paths,omitempty" ini:"paths,omitempty"`
	Server                        *GrafanaConfigServer                        `json:"server,omitempty" ini:"server,omitempty"`
	Database                      *GrafanaConfigDatabase                      `json:"database,omitempty" ini:"database,omitempty"`
	RemoteCache                   *GrafanaConfigRemoteCache                   `json:"remote_cache,omitempty" ini:"remote_cache,omitempty"`
	Security                      *GrafanaConfigSecurity                      `json:"security,omitempty" ini:"security,omitempty"`
	Users                         *GrafanaConfigUsers                         `json:"users,omitempty" ini:"users,omitempty"`
	Auth                          *GrafanaConfigAuth                          `json:"auth,omitempty" ini:"auth,omitempty"`
	AuthBasic                     *GrafanaConfigAuthBasic                     `json:"auth.basic,omitempty" ini:"auth.basic,omitempty"`
	AuthAnonymous                 *GrafanaConfigAuthAnonymous                 `json:"auth.anonymous,omitempty" ini:"auth.anonymous,omitempty"`
	AuthGoogle                    *GrafanaConfigAuthGoogle                    `json:"auth.google,omitempty" ini:"auth.google,omitempty"`
	AuthGithub                    *GrafanaConfigAuthGithub                    `json:"auth.github,omitempty" ini:"auth.github,omitempty"`
	AuthGitlab                    *GrafanaConfigAuthGitlab                    `json:"auth.gitlab,omitempty" ini:"auth.gitlab,omitempty"`
	AuthGenericOauth              *GrafanaConfigAuthGenericOauth              `json:"auth.generic_oauth,omitempty" ini:"auth.generic_oauth,omitempty"`
	AuthLdap                      *GrafanaConfigAuthLdap                      `json:"auth.ldap,omitempty" ini:"auth.ldap,omitempty"`
	AuthProxy                     *GrafanaConfigAuthProxy                     `json:"auth.proxy,omitempty" ini:"auth.proxy,omitempty"`
	DataProxy                     *GrafanaConfigDataProxy                     `json:"dataproxy,omitempty" ini:"dataproxy,omitempty"`
	Analytics                     *GrafanaConfigAnalytics                     `json:"analytics,omitempty" ini:"analytics,omitempty"`
	Dashboards                    *GrafanaConfigDashboards                    `json:"dashboards,omitempty" ini:"dashboards,omitempty"`
	Smtp                          *GrafanaConfigSmtp                          `json:"smtp,omitempty" ini:"smtp,omitempty"`
	Log                           *GrafanaConfigLog                           `json:"log,omitempty" ini:"log,omitempty"`
	Metrics                       *GrafanaConfigMetrics                       `json:"metrics,omitempty" ini:"metrics,omitempty"`
	MetricsGraphite               *GrafanaConfigMetricsGraphite               `json:"metrics.graphite,omitempty" ini:"metrics.graphite,omitempty"`
	Snapshots                     *GrafanaConfigSnapshots                     `json:"snapshots,omitempty" ini:"snapshots,omitempty"`
	ExternalImageStorage          *GrafanaConfigExternalImageStorage          `json:"external_image_storage,omitempty" ini:"external_image_storage,omitempty"`
	ExternalImageStorageS3        *GrafanaConfigExternalImageStorageS3        `json:"external_image_storage.s3,omitempty" ini:"external_image_storage.s3,omitempty"`
	ExternalImageStorageWebdav    *GrafanaConfigExternalImageStorageWebdav    `json:"external_image_storage.webdav,omitempty" ini:"external_image_storage.webdav,omitempty"`
	ExternalImageStorageGcs       *GrafanaConfigExternalImageStorageGcs       `json:"external_image_storage.gcs,omitempty" ini:"external_image_storage.gcs,omitempty"`
	ExternalImageStorageAzureBlob *GrafanaConfigExternalImageStorageAzureBlob `json:"external_image_storage.azure_blob,omitempty" ini:"external_image_storage.azure_blob,omitempty"`
	Alerting                      *GrafanaConfigAlerting                      `json:"alerting,omitempty" ini:"alerting,omitempty"`
	Panels                        *GrafanaConfigPanels                        `json:"panels,omitempty" ini:"panels,omitempty"`
	Plugins                       *GrafanaConfigPlugins                       `json:"plugins,omitempty" ini:"plugins,omitempty"`
}

type GrafanaConfigPaths struct {
	TempDataLifetime string `json:"temp_data_lifetime,omitempty" ini:"temp_data_lifetime,omitempty"`
}

type GrafanaConfigServer struct {
	HttpAddr         string `json:"http_addr,omitempty" ini:"http_addr,omitempty"`
	HttpPort         string `json:"http_port,omitempty" ini:"http_port,omitempty"`
	Protocol         string `json:"protocol,omitempty" ini:"protocol,omitempty"`
	Socket           string `json:"socket,omitempty" ini:"socket,omitempty"`
	Domain           string `json:"domain,omitempty" ini:"domain,omitempty"`
	EnforceDomain    *bool  `json:"enforce_domain,omitempty" ini:"enforce_domain"`
	RootUrl          string `json:"root_url,omitempty" ini:"root_url,omitempty"`
	ServeFromSubPath *bool  `json:"serve_from_sub_path,omitempty" ini:"serve_from_sub_path"`
	StaticRootPath   string `json:"static_root_path,omitempty" ini:"static_root_path,omitempty"`
	EnableGzip       *bool  `json:"enable_gzip,omitempty" ini:"enable_gzip"`
	CertFile         string `json:"cert_file,omitempty" ini:"cert_file,omitempty"`
	CertKey          string `json:"cert_key,omitempty" ini:"cert_key,omitempty"`
	RouterLogging    *bool  `json:"router_logging,omitempty" ini:"router_logging"`
}

type GrafanaConfigDatabase struct {
	Url             string `json:"url,omitempty" ini:"url,omitempty"`
	Type            string `json:"type,omitempty" ini:"type,omitempty"`
	Path            string `json:"path,omitempty" ini:"path,omitempty"`
	Host            string `json:"host,omitempty" ini:"host,omitempty"`
	Name            string `json:"name,omitempty" ini:"name,omitempty"`
	User            string `json:"user,omitempty" ini:"user,omitempty"`
	Password        string `json:"password,omitempty" ini:"password,omitempty"`
	SslMode         string `json:"ssl_mode,omitempty" ini:"ssl_mode,omitempty"`
	CaCertPath      string `json:"ca_cert_path,omitempty" ini:"ca_cert_path,omitempty"`
	ClientKeyPath   string `json:"client_key_path,omitempty" ini:"client_key_path,omitempty"`
	ClientCertPath  string `json:"client_cert_path,omitempty" ini:"client_cert_path,omitempty"`
	ServerCertName  string `json:"server_cert_name,omitempty" ini:"server_cert_name,omitempty"`
	MaxIdleConn     *int   `json:"max_idle_conn,omitempty" ini:"max_idle_conn,omitempty"`
	MaxOpenConn     *int   `json:"max_open_conn,omitempty" ini:"max_open_conn,omitempty"`
	ConnMaxLifetime *int   `json:"conn_max_lifetime,omitempty" ini:"conn_max_lifetime,omitempty"`
	LogQueries      *bool  `json:"log_queries,omitempty" ini:"log_queries"`
	CacheMode       string `json:"cache_mode,omitempty" ini:"cache_mode,omitempty"`
}

type GrafanaConfigRemoteCache struct {
	Type    string `json:"type,omitempty" ini:"type,omitempty"`
	ConnStr string `json:"connstr,omitempty" ini:"connstr,omitempty"`
}

type GrafanaConfigSecurity struct {
	AdminUser                            string `json:"admin_user,omitempty" ini:"admin_user,omitempty"`
	AdminPassword                        string `json:"admin_password,omitempty" ini:"admin_password,omitempty"`
	LoginRememberDays                    *int   `json:"login_remember_days,omitempty" ini:"login_remember_days,omitempty"`
	SecretKey                            string `json:"secret_key,omitempty" ini:"secret_key,omitempty"`
	DisableGravatar                      *bool  `json:"disable_gravatar,omitempty" ini:"disable_gravatar"`
	DataSourceProxyWhitelist             string `json:"data_source_proxy_whitelist,omitempty" ini:"data_source_proxy_whitelist,omitempty"`
	CookieSecure                         *bool  `json:"cookie_secure,omitempty" ini:"cookie_secure"`
	CookieSamesite                       string `json:"cookie_samesite,omitempty" ini:"cookie_samesite,omitempty"`
	AllowEmbedding                       *bool  `json:"allow_embedding,omitempty" ini:"allow_embedding"`
	StrictTransportSecurity              *bool  `json:"strict_transport_security,omitempty" ini:"strict_transport_security"`
	StrictTransportSecurityMaxAgeSeconds *int   `json:"strict_transport_security_max_age_seconds,omitempty" ini:"strict_transport_security_max_age_seconds,omitempty"`
	StrictTransportSecurityPreload       *bool  `json:"strict_transport_security_preload,omitempty" ini:"strict_transport_security_preload"`
	StrictTransportSecuritySubdomains    *bool  `json:"strict_transport_security_subdomains,omitempty" ini:"strict_transport_security_subdomains"`
	XContentTypeOptions                  *bool  `json:"x_content_type_options,omitempty" ini:"x_content_type_options"`
	XXssProtection                       *bool  `json:"x_xss_protection,omitempty" ini:"x_xss_protection"`
}

type GrafanaConfigUsers struct {
	AllowSignUp       *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	AllowOrgCreate    *bool  `json:"allow_org_create,omitempty" ini:"allow_org_create"`
	AutoAssignOrg     *bool  `json:"auto_assign_org,omitempty" ini:"auto_assign_org"`
	AutoAssignOrgId   string `json:"auto_assign_org_id,omitempty" ini:"auto_assign_org_id,omitempty"`
	AutoAssignOrgRole string `json:"auto_assign_org_role,omitempty" ini:"auto_assign_org_role,omitempty"`
	ViewersCanEdit    *bool  `json:"viewers_can_edit,omitempty" ini:"viewers_can_edit"`
	EditorsCanAdmin   *bool  `json:"editors_can_admin,omitempty" ini:"editors_can_admin"`
	LoginHint         string `json:"login_hint,omitempty" ini:"login_hint,omitempty"`
	PasswordHint      string `json:"password_hint,omitempty" ini:"password_hint,omitempty"`
}

type GrafanaConfigAuth struct {
	LoginCookieName                  string `json:"login_cookie_name,omitempty" ini:"login_cookie_name,omitempty"`
	LoginMaximumInactiveLifetimeDays *int   `json:"login_maximum_inactive_lifetime_days,omitempty" ini:"login_maximum_inactive_lifetime_days,omitempty"`
	LoginMaximumLifetimeDays         *int   `json:"login_maximum_lifetime_days,omitempty" ini:"login_maximum_lifetime_days,omitempty"`
	TokenRotationIntervalMinutes     *int   `json:"token_rotation_interval_minutes,omitempty" ini:"token_rotation_interval_minutes,omitempty"`
	DisableLoginForm                 *bool  `json:"disable_login_form,omitempty" ini:"disable_login_form"`
	DisableSignoutMenu               *bool  `json:"disable_signout_menu,omitempty" ini:"disable_signout_menu"`
	SignoutRedirectUrl               string `json:"signout_redirect_url,omitempty" ini:"signout_redirect_url,omitempty"`
	OauthAutoLogin                   *bool  `json:"oauth_auto_login,omitempty" ini:"oauth_auto_login"`
}

type GrafanaConfigAuthBasic struct {
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
}

type GrafanaConfigAuthAnonymous struct {
	Enabled *bool  `json:"enabled,omitempty" ini:"enabled"`
	OrgName string `json:"org_name,omitempty" ini:"org_name,omitempty"`
	OrgRole string `json:"org_role,omitempty" ini:"org_role,omitempty"`
}

type GrafanaConfigAuthGoogle struct {
	Enabled        *bool  `json:"enabled,omitempty" ini:"enabled"`
	ClientId       string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes         string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl        string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl       string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	AllowedDomains string `json:"allowed_domains,omitempty" ini:"allowed_domains,omitempty"`
	AllowSignUp    *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
}

type GrafanaConfigAuthGithub struct {
	Enabled              *bool  `json:"enabled,omitempty" ini:"enabled"`
	AllowSignUp          *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId             string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret         string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes               string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl              string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl             string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	ApiUrl               string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	TeamIds              string `json:"team_ids,omitempty" ini:"team_ids,omitempty"`
	AllowedOrganizations string `json:"allowed_organizations,omitempty" ini:"allowed_organizations,omitempty"`
}

type GrafanaConfigAuthGitlab struct {
	Enabled       *bool  `json:"enabled,omitempty" ini:"enabled"`
	AllowSignUp   *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId      string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret  string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes        string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl       string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl      string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	ApiUrl        string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	AllowedGroups string `json:"allowed_groups,omitempty" ini:"allowed_groups,omitempty"`
}

type GrafanaConfigAuthGenericOauth struct {
	Enabled           *bool  `json:"enabled,omitempty" ini:"enabled"`
	AllowSignUp       *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId          string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret      string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes            string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl           string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl          string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	ApiUrl            string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	AllowedDomains    string `json:"allowed_domains,omitempty" ini:"allowed_domains,omitempty"`
	RoleAttributePath string `json:"role_attribute_path,omitempty" ini:"role_attribute_path,omitempty"`
}

type GrafanaConfigAuthLdap struct {
	Enabled     *bool  `json:"enabled,omitempty" ini:"enabled"`
	AllowSignUp *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ConfigFile  string `json:"config_file,omitempty" ini:"config_file,omitempty"`
}

type GrafanaConfigAuthProxy struct {
	Enabled          *bool  `json:"enabled,omitempty" ini:"enabled"`
	HeaderName       string `json:"header_name,omitempty" ini:"header_name,omitempty"`
	HeaderProperty   string `json:"header_property,omitempty" ini:"header_property,omitempty"`
	AutoSignUp       *bool  `json:"auto_sign_up,omitempty" ini:"auto_sign_up"`
	LdapSyncTtl      string `json:"ldap_sync_ttl,omitempty" ini:"ldap_sync_ttl,omitempty"`
	Whitelist        string `json:"whitelist,omitempty" ini:"whitelist,omitempty"`
	Headers          string `json:"headers,omitempty" ini:"headers,omitempty"`
	EnableLoginToken *bool  `json:"enable_login_token,omitempty" ini:"enable_login_token"`
}

type GrafanaConfigDataProxy struct {
	Logging        *bool `json:"logging,omitempty" ini:"logging"`
	Timeout        *int  `json:"timeout,omitempty" ini:"timeout,omitempty"`
	SendUserHeader *bool `json:"send_user_header,omitempty" ini:"send_user_header,omitempty"`
}

type GrafanaConfigAnalytics struct {
	ReportingEnabled    *bool  `json:"reporting_enabled,omitempty" ini:"reporting_enabled"`
	GoogleAnalyticsUaId string `json:"google_analytics_ua_id,omitempty" ini:"google_analytics_ua_id,omitempty"`
	CheckForUpdates     *bool  `json:"check_for_updates,omitempty" ini:"check_for_updates"`
}

type GrafanaConfigDashboards struct {
	VersionsToKeep *int `json:"versions_to_keep,omitempty" ini:"versions_to_keep,omitempty"`
}

type GrafanaConfigSmtp struct {
	Enabled      *bool  `json:"enabled,omitempty" ini:"enabled"`
	Host         string `json:"host,omitempty" ini:"host,omitempty"`
	User         string `json:"user,omitempty" ini:"user,omitempty"`
	Password     string `json:"password,omitempty" ini:"password,omitempty"`
	CertFile     string `json:"cert_file,omitempty" ini:"cert_file,omitempty"`
	KeyFile      string `json:"key_file,omitempty" ini:"key_file,omitempty"`
	SkipVerify   *bool  `json:"skip_verify,omitempty" ini:"skip_verify"`
	FromAddress  string `json:"from_address,omitempty" ini:"from_address,omitempty"`
	FromName     string `json:"from_name,omitempty" ini:"from_name,omitempty"`
	EhloIdentity string `json:"ehlo_identity,omitempty" ini:"ehlo_identity,omitempty"`
}

type GrafanaConfigLog struct {
	Mode    string `json:"mode,omitempty" ini:"mode,omitempty"`
	Level   string `json:"level,omitempty" ini:"level,omitempty"`
	Filters string `json:"filters,omitempty" ini:"filters,omitempty"`
}

type GrafanaConfigMetrics struct {
	Enabled           *bool  `json:"enabled,omitempty" ini:"enabled"`
	BasicAuthUsername string `json:"basic_auth_username,omitempty" ini:"basic_auth_username,omitempty"`
	BasicAuthPassword string `json:"basic_auth_password,omitempty" ini:"basic_auth_password,omitempty"`
	IntervalSeconds   *int   `json:"interval_seconds,omitempty" ini:"interval_seconds,omitempty"`
}

type GrafanaConfigMetricsGraphite struct {
	Address string `json:"address,omitempty" ini:"address,omitempty"`
	Prefix  string `json:"prefix,omitempty" ini:"prefix,omitempty"`
}

type GrafanaConfigSnapshots struct {
	ExternalEnabled       *bool  `json:"external_enabled,omitempty" ini:"external_enabled"`
	ExternalSnapshotUrl   string `json:"external_snapshot_url,omitempty" ini:"external_snapshot_url,omitempty"`
	ExternalSnapshotName  string `json:"external_snapshot_name,omitempty" ini:"external_snapshot_name,omitempty"`
	SnapshotRemoveExpired *bool  `json:"snapshot_remove_expired,omitempty" ini:"snapshot_remove_expired"`
}

type GrafanaConfigExternalImageStorage struct {
	Provider string `json:"provider,omitempty" ini:"provider,omitempty"`
}

type GrafanaConfigExternalImageStorageS3 struct {
	Bucket    string `json:"bucket,omitempty" ini:"bucket,omitempty"`
	Region    string `json:"region,omitempty" ini:"region,omitempty"`
	Path      string `json:"path,omitempty" ini:"path,omitempty"`
	BucketUrl string `json:"bucket_url,omitempty" ini:"bucket_url,omitempty"`
	AccessKey string `json:"access_key,omitempty" ini:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty" ini:"secret_key,omitempty"`
}

type GrafanaConfigExternalImageStorageWebdav struct {
	Url       string `json:"url,omitempty" ini:"url,omitempty"`
	PublicUrl string `json:"public_url,omitempty" ini:"public_url,omitempty"`
	Username  string `json:"username,omitempty" ini:"username,omitempty"`
	Password  string `json:"password,omitempty" ini:"password,omitempty"`
}

type GrafanaConfigExternalImageStorageGcs struct {
	KeyFile string `json:"key_file,omitempty" ini:"key_file,omitempty"`
	Bucket  string `json:"bucket,omitempty" ini:"bucket,omitempty"`
	Path    string `json:"path,omitempty" ini:"path,omitempty"`
}

type GrafanaConfigExternalImageStorageAzureBlob struct {
	AccountName   string `json:"account_name,omitempty" ini:"account_name,omitempty"`
	AccountKey    string `json:"account_key,omitempty" ini:"account_key,omitempty"`
	ContainerName string `json:"container_name,omitempty" ini:"container_name,omitempty"`
}

type GrafanaConfigAlerting struct {
	Enabled                    *bool  `json:"enabled,omitempty" ini:"enabled"`
	ExecuteAlerts              *bool  `json:"execute_alerts,omitempty" ini:"execute_alerts"`
	ErrorOrTimeout             string `json:"error_or_timeout,omitempty" ini:"error_or_timeout,omitempty"`
	NodataOrNullvalues         string `json:"nodata_or_nullvalues,omitempty" ini:"nodata_or_nullvalues,omitempty"`
	ConcurrentRenderLimit      *int   `json:"concurrent_render_limit,omitempty" ini:"concurrent_render_limit,omitempty"`
	EvaluationTimeoutSeconds   *int   `json:"evaluation_timeout_seconds,omitempty" ini:"evaluation_timeout_seconds,omitempty"`
	NotificationTimeoutSeconds *int   `json:"notification_timeout_seconds,omitempty" ini:"notification_timeout_seconds,omitempty"`
	MaxAttempts                *int   `json:"max_attempts,omitempty" ini:"max_attempts,omitempty"`
}

type GrafanaConfigPanels struct {
	DisableSanitizeHtml *bool `json:"disable_sanitize_html,omitempty" ini:"disable_sanitize_html"`
}

type GrafanaConfigPlugins struct {
	EnableAlpha *bool `json:"enable_alpha,omitempty" ini:"enable_alpha"`
}

// GrafanaStatus defines the observed state of Grafana
// +k8s:openapi-gen=true
type GrafanaStatus struct {
	Phase               StatusPhase                       `json:"phase"`
	Message             string                            `json:"message"`
	InstalledDashboards map[string][]*GrafanaDashboardRef `json:"dashboards"`
	InstalledPlugins    PluginList                        `json:"installedPlugins"`
	FailedPlugins       PluginList                        `json:"failedPlugins"`
}

// GrafanaPlugin contains information about a single plugin
type GrafanaPlugin struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Grafana is the Schema for the grafanas API
// +k8s:openapi-gen=true
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaSpec   `json:"spec,omitempty"`
	Status GrafanaStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GrafanaList contains a list of Grafana
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Grafana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Grafana{}, &GrafanaList{})
}
