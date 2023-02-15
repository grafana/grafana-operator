package v1alpha1

import (
	v12 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusPhase string

var (
	NoPhase          StatusPhase
	PhaseReconciling StatusPhase = "reconciling"
	PhaseFailing     StatusPhase = "failing"
)

// GrafanaSpec defines the desired state of Grafana

type GrafanaSpec struct {
	Config                     GrafanaConfig            `json:"config"`
	Containers                 []v1.Container           `json:"containers,omitempty"`
	DashboardLabelSelector     []*metav1.LabelSelector  `json:"dashboardLabelSelector,omitempty"`
	Ingress                    *GrafanaIngress          `json:"ingress,omitempty"`
	InitResources              *v1.ResourceRequirements `json:"initResources,omitempty"`
	Secrets                    []string                 `json:"secrets,omitempty"`
	ConfigMaps                 []string                 `json:"configMaps,omitempty"`
	Service                    *GrafanaService          `json:"service,omitempty"`
	Deployment                 *GrafanaDeployment       `json:"deployment,omitempty"`
	Resources                  *v1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccount             *GrafanaServiceAccount   `json:"serviceAccount,omitempty"`
	Client                     *GrafanaClient           `json:"client,omitempty"`
	DashboardNamespaceSelector *metav1.LabelSelector    `json:"dashboardNamespaceSelector,omitempty"`
	DataStorage                *GrafanaDataStorage      `json:"dataStorage,omitempty"`
	Jsonnet                    *JsonnetConfig           `json:"jsonnet,omitempty"`
	BaseImage                  string                   `json:"baseImage,omitempty"`
	InitImage                  string                   `json:"initImage,omitempty"`
	LivenessProbeSpec          *LivenessProbeSpec       `json:"livenessProbeSpec,omitempty"`
	ReadinessProbeSpec         *ReadinessProbeSpec      `json:"readinessProbeSpec,omitempty"`

	// DashboardContentCacheDuration sets a default for when a `GrafanaDashboard` resource doesn't specify a `contentCacheDuration`.
	// If left unset or 0 the default behavior is to cache indefinitely.
	DashboardContentCacheDuration metav1.Duration `json:"dashboardContentCacheDuration,omitempty"`
}

type ReadinessProbeSpec struct {
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	TimeOutSeconds      *int32 `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       *int32 `json:"periodSeconds,omitempty"`
	SuccessThreshold    *int32 `json:"successThreshold,omitempty"`
	FailureThreshold    *int32 `json:"failureThreshold,omitempty"`
	// URIScheme identifies the scheme used for connection to a host for Get actions. Deprecated in favor of config.server.protocol.
	Scheme v1.URIScheme `json:"scheme,omitempty"`
}
type LivenessProbeSpec struct {
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,omitempty"`
	TimeOutSeconds      *int32 `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       *int32 `json:"periodSeconds,omitempty"`
	SuccessThreshold    *int32 `json:"successThreshold,omitempty"`
	FailureThreshold    *int32 `json:"failureThreshold,omitempty"`
	// URIScheme identifies the scheme used for connection to a host for Get actions. Deprecated in favor of config.server.protocol.
	Scheme v1.URIScheme `json:"scheme,omitempty"`
}

type JsonnetConfig struct {
	LibraryLabelSelector *metav1.LabelSelector `json:"libraryLabelSelector,omitempty"`
}

// GrafanaClient contains the Grafana API client settings
type GrafanaClient struct {
	// +nullable
	TimeoutSeconds *int `json:"timeout,omitempty"`
	// +nullable
	PreferService *bool `json:"preferService,omitempty"`
}

// GrafanaService provides a means to configure the service
type GrafanaService struct {
	Name        string            `json:"name,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Type        v1.ServiceType    `json:"type,omitempty"`
	AppProtocol string            `json:"appProtocol,omitempty"`
	Ports       []v1.ServicePort  `json:"ports,omitempty"`
	ClusterIP   string            `json:"clusterIP,omitempty"`
}

// GrafanaDataStorage provides a means to configure the grafana data storage
type GrafanaDataStorage struct {
	Annotations map[string]string               `json:"annotations,omitempty"`
	Labels      map[string]string               `json:"labels,omitempty"`
	AccessModes []v1.PersistentVolumeAccessMode `json:"accessModes,omitempty"`
	Size        resource.Quantity               `json:"size,omitempty"`
	Class       string                          `json:"class,omitempty"`
	VolumeName  string                          `json:"volumeName,omitempty"`
}

type GrafanaServiceAccount struct {
	Skip             *bool                     `json:"skip,omitempty"`
	Annotations      map[string]string         `json:"annotations,omitempty"`
	Labels           map[string]string         `json:"labels,omitempty"`
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty"`
}

// GrafanaDeployment provides a means to configure the deployment
type GrafanaDeployment struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	// +nullable
	Replicas                      *int32                        `json:"replicas,omitempty"`
	NodeSelector                  map[string]string             `json:"nodeSelector,omitempty"`
	Tolerations                   []v1.Toleration               `json:"tolerations,omitempty"`
	Affinity                      *v1.Affinity                  `json:"affinity,omitempty"`
	TopologySpreadConstraints     []v1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`
	SecurityContext               *v1.PodSecurityContext        `json:"securityContext,omitempty"`
	ContainerSecurityContext      *v1.SecurityContext           `json:"containerSecurityContext,omitempty"`
	TerminationGracePeriodSeconds *int64                        `json:"terminationGracePeriodSeconds,omitempty"`
	EnvFrom                       []v1.EnvFromSource            `json:"envFrom,omitempty"`
	Env                           []v1.EnvVar                   `json:"env,omitempty"`
	// +nullable
	SkipCreateAdminAccount *bool  `json:"skipCreateAdminAccount,omitempty"`
	PriorityClassName      string `json:"priorityClassName,omitempty"`
	// +nullable
	HostNetwork       *bool                      `json:"hostNetwork,omitempty"`
	ExtraVolumes      []v1.Volume                `json:"extraVolumes,omitempty"`
	ExtraVolumeMounts []v1.VolumeMount           `json:"extraVolumeMounts,omitempty"`
	Strategy          *appsv1.DeploymentStrategy `json:"strategy,omitempty"`
	HttpProxy         *GrafanaHttpProxy          `json:"httpProxy,omitempty"`
}

// GrafanaHttpProxy provides a means to configure the Grafana deployment
// to use an HTTP(S) proxy when making requests and resolving plugins.
type GrafanaHttpProxy struct {
	Enabled   bool   `json:"enabled"`
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureUrl,omitempty"`
	NoProxy   string `json:"noProxy,omitempty"`
}

// GrafanaIngress provides a means to configure the ingress created
type GrafanaIngress struct {
	Annotations      map[string]string      `json:"annotations,omitempty"`
	Hostname         string                 `json:"hostname,omitempty"`
	Labels           map[string]string      `json:"labels,omitempty"`
	Path             string                 `json:"path,omitempty"`
	Enabled          bool                   `json:"enabled,omitempty"`
	TLSEnabled       bool                   `json:"tlsEnabled,omitempty"`
	TLSSecretName    string                 `json:"tlsSecretName,omitempty"`
	TargetPort       string                 `json:"targetPort,omitempty"`
	Termination      v12.TLSTerminationType `json:"termination,omitempty"`
	IngressClassName string                 `json:"ingressClassName,omitempty"`
	PathType         string                 `json:"pathType,omitempty"`
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
	AuthAzureAD                   *GrafanaConfigAuthAzureAD                   `json:"auth.azuread,omitempty" ini:"auth.azuread,omitempty"`
	AuthGoogle                    *GrafanaConfigAuthGoogle                    `json:"auth.google,omitempty" ini:"auth.google,omitempty"`
	AuthGithub                    *GrafanaConfigAuthGithub                    `json:"auth.github,omitempty" ini:"auth.github,omitempty"`
	AuthGitlab                    *GrafanaConfigAuthGitlab                    `json:"auth.gitlab,omitempty" ini:"auth.gitlab,omitempty"`
	AuthGenericOauth              *GrafanaConfigAuthGenericOauth              `json:"auth.generic_oauth,omitempty" ini:"auth.generic_oauth,omitempty"`
	AuthJwt                       *GrafanaConfigAuthJwt                       `json:"auth.jwt,omitempty" ini:"auth.jwt,omitempty"`
	AuthOkta                      *GrafanaConfigAuthOkta                      `json:"auth.okta,omitempty" ini:"auth.okta,omitempty"`
	AuthLdap                      *GrafanaConfigAuthLdap                      `json:"auth.ldap,omitempty" ini:"auth.ldap,omitempty"`
	AuthProxy                     *GrafanaConfigAuthProxy                     `json:"auth.proxy,omitempty" ini:"auth.proxy,omitempty"`
	AuthSaml                      *GrafanaConfigAuthSaml                      `json:"auth.saml,omitempty" ini:"auth.saml,omitempty"`
	DataProxy                     *GrafanaConfigDataProxy                     `json:"dataproxy,omitempty" ini:"dataproxy,omitempty"`
	Analytics                     *GrafanaConfigAnalytics                     `json:"analytics,omitempty" ini:"analytics,omitempty"`
	Dashboards                    *GrafanaConfigDashboards                    `json:"dashboards,omitempty" ini:"dashboards,omitempty"`
	Smtp                          *GrafanaConfigSmtp                          `json:"smtp,omitempty" ini:"smtp,omitempty"`
	Live                          *GrafanaConfigLive                          `json:"live,omitempty" ini:"live,omitempty"`
	Log                           *GrafanaConfigLog                           `json:"log,omitempty" ini:"log,omitempty"`
	LogConsole                    *GrafanaConfigLogConsole                    `json:"log.console,omitempty" ini:"log.console,omitempty"`
	LogFrontend                   *GrafanaConfigLogFrontend                   `json:"log.frontend,omitempty" ini:"log.frontend,omitempty"`
	Metrics                       *GrafanaConfigMetrics                       `json:"metrics,omitempty" ini:"metrics,omitempty"`
	MetricsGraphite               *GrafanaConfigMetricsGraphite               `json:"metrics.graphite,omitempty" ini:"metrics.graphite,omitempty"`
	Snapshots                     *GrafanaConfigSnapshots                     `json:"snapshots,omitempty" ini:"snapshots,omitempty"`
	ExternalImageStorage          *GrafanaConfigExternalImageStorage          `json:"external_image_storage,omitempty" ini:"external_image_storage,omitempty"`
	ExternalImageStorageS3        *GrafanaConfigExternalImageStorageS3        `json:"external_image_storage.s3,omitempty" ini:"external_image_storage.s3,omitempty"`
	ExternalImageStorageWebdav    *GrafanaConfigExternalImageStorageWebdav    `json:"external_image_storage.webdav,omitempty" ini:"external_image_storage.webdav,omitempty"`
	ExternalImageStorageGcs       *GrafanaConfigExternalImageStorageGcs       `json:"external_image_storage.gcs,omitempty" ini:"external_image_storage.gcs,omitempty"`
	ExternalImageStorageAzureBlob *GrafanaConfigExternalImageStorageAzureBlob `json:"external_image_storage.azure_blob,omitempty" ini:"external_image_storage.azure_blob,omitempty"`
	Alerting                      *GrafanaConfigAlerting                      `json:"alerting,omitempty" ini:"alerting,omitempty"`
	UnifiedAlerting               *GrafanaConfigUnifiedAlerting               `json:"unified_alerting,omitempty" ini:"unified_alerting,omitempty"`
	Panels                        *GrafanaConfigPanels                        `json:"panels,omitempty" ini:"panels,omitempty"`
	Plugins                       *GrafanaConfigPlugins                       `json:"plugins,omitempty" ini:"plugins,omitempty"`
	Rendering                     *GrafanaConfigRendering                     `json:"rendering,omitempty" ini:"rendering,omitempty"`
	FeatureToggles                *GrafanaConfigFeatureToggles                `json:"feature_toggles,omitempty" ini:"feature_toggles,omitempty"`
}

type GrafanaConfigPaths struct {
	TempDataLifetime string `json:"temp_data_lifetime,omitempty" ini:"temp_data_lifetime,omitempty"`
}

type GrafanaConfigServer struct {
	HttpAddr string `json:"http_addr,omitempty" ini:"http_addr,omitempty"`
	HttpPort string `json:"http_port,omitempty" ini:"http_port,omitempty"`
	// +kubebuilder:validation:Enum=http;https
	Protocol string `json:"protocol,omitempty" ini:"protocol,omitempty"`
	Socket   string `json:"socket,omitempty" ini:"socket,omitempty"`
	Domain   string `json:"domain,omitempty" ini:"domain,omitempty"`
	// +nullable
	EnforceDomain *bool  `json:"enforce_domain,omitempty" ini:"enforce_domain"`
	RootUrl       string `json:"root_url,omitempty" ini:"root_url,omitempty"`
	// +nullable
	ServeFromSubPath *bool  `json:"serve_from_sub_path,omitempty" ini:"serve_from_sub_path"`
	StaticRootPath   string `json:"static_root_path,omitempty" ini:"static_root_path,omitempty"`
	// +nullable
	EnableGzip *bool  `json:"enable_gzip,omitempty" ini:"enable_gzip"`
	CertFile   string `json:"cert_file,omitempty" ini:"cert_file,omitempty"`
	CertKey    string `json:"cert_key,omitempty" ini:"cert_key,omitempty"`
	// +nullable
	RouterLogging *bool `json:"router_logging,omitempty" ini:"router_logging"`
}

type GrafanaConfigDatabase struct {
	Url            string `json:"url,omitempty" ini:"url,omitempty"`
	Type           string `json:"type,omitempty" ini:"type,omitempty"`
	Path           string `json:"path,omitempty" ini:"path,omitempty"`
	Host           string `json:"host,omitempty" ini:"host,omitempty"`
	Name           string `json:"name,omitempty" ini:"name,omitempty"`
	User           string `json:"user,omitempty" ini:"user,omitempty"`
	Password       string `json:"password,omitempty" ini:"password,omitempty"`
	SslMode        string `json:"ssl_mode,omitempty" ini:"ssl_mode,omitempty"`
	CaCertPath     string `json:"ca_cert_path,omitempty" ini:"ca_cert_path,omitempty"`
	ClientKeyPath  string `json:"client_key_path,omitempty" ini:"client_key_path,omitempty"`
	ClientCertPath string `json:"client_cert_path,omitempty" ini:"client_cert_path,omitempty"`
	ServerCertName string `json:"server_cert_name,omitempty" ini:"server_cert_name,omitempty"`
	// +nullable
	MaxIdleConn *int `json:"max_idle_conn,omitempty" ini:"max_idle_conn,omitempty"`
	// +nullable
	MaxOpenConn *int `json:"max_open_conn,omitempty" ini:"max_open_conn,omitempty"`
	// +nullable
	ConnMaxLifetime *int `json:"conn_max_lifetime,omitempty" ini:"conn_max_lifetime,omitempty"`
	// +nullable
	LogQueries *bool  `json:"log_queries,omitempty" ini:"log_queries"`
	CacheMode  string `json:"cache_mode,omitempty" ini:"cache_mode,omitempty"`
}

type GrafanaConfigRemoteCache struct {
	Type    string `json:"type,omitempty" ini:"type,omitempty"`
	ConnStr string `json:"connstr,omitempty" ini:"connstr,omitempty"`
}

type GrafanaConfigSecurity struct {
	AdminUser     string `json:"admin_user,omitempty" ini:"admin_user,omitempty"`
	AdminPassword string `json:"admin_password,omitempty" ini:"admin_password,omitempty"`
	// +nullable
	LoginRememberDays *int   `json:"login_remember_days,omitempty" ini:"login_remember_days,omitempty"`
	SecretKey         string `json:"secret_key,omitempty" ini:"secret_key,omitempty"`
	// +nullable
	DisableGravatar          *bool  `json:"disable_gravatar,omitempty" ini:"disable_gravatar"`
	DataSourceProxyWhitelist string `json:"data_source_proxy_whitelist,omitempty" ini:"data_source_proxy_whitelist,omitempty"`
	// +nullable
	CookieSecure   *bool  `json:"cookie_secure,omitempty" ini:"cookie_secure"`
	CookieSamesite string `json:"cookie_samesite,omitempty" ini:"cookie_samesite,omitempty"`
	// +nullable
	AllowEmbedding *bool `json:"allow_embedding,omitempty" ini:"allow_embedding"`
	// +nullable
	StrictTransportSecurity *bool `json:"strict_transport_security,omitempty" ini:"strict_transport_security"`
	// +nullable
	StrictTransportSecurityMaxAgeSeconds *int `json:"strict_transport_security_max_age_seconds,omitempty" ini:"strict_transport_security_max_age_seconds,omitempty"`
	// +nullable
	StrictTransportSecurityPreload *bool `json:"strict_transport_security_preload,omitempty" ini:"strict_transport_security_preload"`
	// +nullable
	StrictTransportSecuritySubdomains *bool `json:"strict_transport_security_subdomains,omitempty" ini:"strict_transport_security_subdomains"`
	// +nullable
	XContentTypeOptions *bool `json:"x_content_type_options,omitempty" ini:"x_content_type_options"`
	// +nullable
	XXssProtection *bool `json:"x_xss_protection,omitempty" ini:"x_xss_protection"`
}

type GrafanaConfigUsers struct {
	// +nullable
	AllowSignUp *bool `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	// +nullable
	AllowOrgCreate *bool `json:"allow_org_create,omitempty" ini:"allow_org_create"`
	// +nullable
	AutoAssignOrg     *bool  `json:"auto_assign_org,omitempty" ini:"auto_assign_org"`
	AutoAssignOrgId   string `json:"auto_assign_org_id,omitempty" ini:"auto_assign_org_id,omitempty"`
	AutoAssignOrgRole string `json:"auto_assign_org_role,omitempty" ini:"auto_assign_org_role,omitempty"`
	// +nullable
	ViewersCanEdit *bool `json:"viewers_can_edit,omitempty" ini:"viewers_can_edit"`
	// +nullable
	EditorsCanAdmin *bool  `json:"editors_can_admin,omitempty" ini:"editors_can_admin"`
	LoginHint       string `json:"login_hint,omitempty" ini:"login_hint,omitempty"`
	PasswordHint    string `json:"password_hint,omitempty" ini:"password_hint,omitempty"`
	DefaultTheme    string `json:"default_theme,omitempty" ini:"default_theme,omitempty"`
}

type GrafanaConfigAuth struct {
	LoginCookieName string `json:"login_cookie_name,omitempty" ini:"login_cookie_name,omitempty"`
	// +nullable
	LoginMaximumInactiveLifetimeDays     *int   `json:"login_maximum_inactive_lifetime_days,omitempty" ini:"login_maximum_inactive_lifetime_days,omitempty"`
	LoginMaximumInactiveLifetimeDuration string `json:"login_maximum_inactive_lifetime_duration,omitempty" ini:"login_maximum_inactive_lifetime_duration,omitempty"`
	// +nullable
	LoginMaximumLifetimeDays     *int   `json:"login_maximum_lifetime_days,omitempty" ini:"login_maximum_lifetime_days,omitempty"`
	LoginMaximumLifetimeDuration string `json:"login_maximum_lifetime_duration,omitempty" ini:"login_maximum_lifetime_duration,omitempty"`
	// +nullable
	TokenRotationIntervalMinutes *int `json:"token_rotation_interval_minutes,omitempty" ini:"token_rotation_interval_minutes,omitempty"`
	// +nullable
	DisableLoginForm *bool `json:"disable_login_form,omitempty" ini:"disable_login_form"`
	// +nullable
	DisableSignoutMenu *bool `json:"disable_signout_menu,omitempty" ini:"disable_signout_menu"`
	// +nullable
	SigV4AuthEnabled   *bool  `json:"sigv4_auth_enabled,omitempty" ini:"sigv4_auth_enabled"`
	SignoutRedirectUrl string `json:"signout_redirect_url,omitempty" ini:"signout_redirect_url,omitempty"`
	// +nullable
	OauthAutoLogin *bool `json:"oauth_auto_login,omitempty" ini:"oauth_auto_login"`
}

type GrafanaConfigAuthBasic struct {
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
}

type GrafanaConfigAuthAnonymous struct {
	// +nullable
	Enabled *bool  `json:"enabled,omitempty" ini:"enabled"`
	OrgName string `json:"org_name,omitempty" ini:"org_name,omitempty"`
	OrgRole string `json:"org_role,omitempty" ini:"org_role,omitempty"`
}

type GrafanaConfigAuthSaml struct {
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	SingleLogout *bool `json:"single_logout,omitempty" ini:"single_logout,omitempty"`
	// +nullable
	AllowIdpInitiated        *bool  `json:"allow_idp_initiated,omitempty" ini:"allow_idp_initiated,omitempty"`
	CertificatePath          string `json:"certificate_path,omitempty" ini:"certificate_path"`
	KeyPath                  string `json:"private_key_path,omitempty" ini:"private_key_path"`
	SignatureAlgorithm       string `json:"signature_algorithm,omitempty" ini:"signature_algorithm,omitempty"`
	IdpUrl                   string `json:"idp_metadata_url,omitempty" ini:"idp_metadata_url"`
	MaxIssueDelay            string `json:"max_issue_delay,omitempty" ini:"max_issue_delay,omitempty"`
	MetadataValidDuration    string `json:"metadata_valid_duration,omitempty" ini:"metadata_valid_duration,omitempty"`
	RelayState               string `json:"relay_state,omitempty" ini:"relay_state,omitempty"`
	AssertionAttributeName   string `json:"assertion_attribute_name,omitempty" ini:"assertion_attribute_name,omitempty"`
	AssertionAttributeLogin  string `json:"assertion_attribute_login,omitempty" ini:"assertion_attribute_login,omitempty"`
	AssertionAttributeEmail  string `json:"assertion_attribute_email,omitempty" ini:"assertion_attribute_email,omitempty"`
	AssertionAttributeGroups string `json:"assertion_attribute_groups,omitempty" ini:"assertion_attribute_groups,omitempty"`
	AssertionAttributeRole   string `json:"assertion_attribute_role,omitempty" ini:"assertion_attribute_role,omitempty"`
	AssertionAttributeOrg    string `json:"assertion_attribute_org,omitempty" ini:"assertion_attribute_org,omitempty"`
	AllowedOrganizations     string `json:"allowed_organizations,omitempty" ini:"allowed_organizations,omitempty"`
	OrgMapping               string `json:"org_mapping,omitempty" ini:"org_mapping,omitempty"`
	RoleValuesEditor         string `json:"role_values_editor,omitempty" ini:"role_values_editor,omitempty"`
	RoleValuesAdmin          string `json:"role_values_admin,omitempty" ini:"role_values_admin,omitempty"`
	RoleValuesGrafanaAdmin   string `json:"role_values_grafana_admin,omitempty" ini:"role_values_grafana_admin,omitempty"`
}

type GrafanaConfigAuthAzureAD struct {
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	AllowSignUp    *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId       string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret   string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes         string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl        string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl       string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	AllowedDomains string `json:"allowed_domains,omitempty" ini:"allowed_domains,omitempty"`
	AllowedGroups  string `json:"allowed_groups,omitempty" ini:"allowed_groups,omitempty"`
}

type GrafanaConfigAuthGoogle struct {
	// +nullable
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
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
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
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	AllowSignUp       *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId          string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret      string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes            string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl           string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl          string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	ApiUrl            string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	AllowedGroups     string `json:"allowed_groups,omitempty" ini:"allowed_groups,omitempty"`
	RoleAttributePath string `json:"role_attribute_path,omitempty" ini:"role_attribute_path,omitempty"`
	// +nullable
	RoleAttributeStrict *bool `json:"role_attribute_strict,omitempty" ini:"role_attribute_strict,omitempty"`
	// +nullable
	AllowAssignGrafanaAdmin *bool `json:"allow_assign_grafana_admin,omitempty" ini:"allow_assign_grafana_admin,omitempty"`
}

type GrafanaConfigAuthGenericOauth struct {
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	Name                 string `json:"name,omitempty" ini:"name,omitempty"`
	AllowSignUp          *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId             string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret         string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes               string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl              string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl             string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	UsePkce              *bool  `json:"use_pkce,omitempty" ini:"use_pkce,omitempty"`
	ApiUrl               string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	TeamsURL             string `json:"teams_url,omitempty" ini:"teams_url,omitempty"`
	TeamIds              string `json:"team_ids,omitempty" ini:"team_ids,omitempty"`
	TeamIdsAttributePath string `json:"team_ids_attribute_path,omitempty" ini:"team_ids_attribute_path,omitempty"`
	AllowedDomains       string `json:"allowed_domains,omitempty" ini:"allowed_domains,omitempty"`
	RoleAttributePath    string `json:"role_attribute_path,omitempty" ini:"role_attribute_path,omitempty"`
	// +nullable
	RoleAttributeStrict *bool  `json:"role_attribute_strict,omitempty" ini:"role_attribute_strict,omitempty"`
	EmailAttributePath  string `json:"email_attribute_path,omitempty" ini:"email_attribute_path,omitempty"`
	// +nullable
	TLSSkipVerifyInsecure *bool  `json:"tls_skip_verify_insecure,omitempty" ini:"tls_skip_verify_insecure,omitempty"`
	TLSClientCert         string `json:"tls_client_cert,omitempty" ini:"tls_client_cert,omitempty"`
	TLSClientKey          string `json:"tls_client_key,omitempty" ini:"tls_client_key,omitempty"`
	TLSClientCa           string `json:"tls_client_ca,omitempty" ini:"tls_auth_ca,omitempty"`
}

type GrafanaConfigAuthJwt struct {
	Enabled                 *bool  `json:"enabled,omitempty" ini:"enabled"`
	EnableLoginToken        *bool  `json:"enable_login_token,omitempty" ini:"enable_login_token"`
	HeaderName              string `json:"header_name,omitempty" ini:"header_name,omitempty"`
	EmailClaim              string `json:"email_claim,omitempty" ini:"email_claim,omitempty"`
	ExpectClaims            string `json:"expect_claims,omitempty" ini:"expect_claims,omitempty"`
	UsernameClaim           string `json:"username_claim,omitempty" ini:"username_claim,omitempty"`
	JwkSetUrl               string `json:"jwk_set_url,omitempty" ini:"jwk_set_url,omitempty"`
	JwkSetFile              string `json:"jwk_set_file,omitempty" ini:"jwk_set_file,omitempty"`
	KeyFile                 string `json:"key_file,omitempty" ini:"key_file,omitempty"`
	RoleAttributePath       string `json:"role_attribute_path,omitempty" ini:"role_attribute_path,omitempty"`
	RoleAttributeStrict     *bool  `json:"role_attribute_strict,omitempty" ini:"role_attribute_strict,omitempty"`
	AutoSignUp              *bool  `json:"auto_sign_up,omitempty" ini:"auto_sign_up,omitempty"`
	CacheTtl                string `json:"cache_ttl,omitempty" ini:"cache_ttl,omitempty"`
	UrlLogin                *bool  `json:"url_login,omitempty" ini:"url_login,omitempty"`
	AllowAssignGrafanaAdmin *bool  `json:"allow_assign_grafana_admin,omitempty" ini:"allow_assign_grafana_admin,omitempty"`
	SkipOrgRoleSync         *bool  `json:"skip_org_role_sync,omitempty" ini:"skip_org_role_sync,omitempty"`
}

type GrafanaConfigAuthOkta struct {
	// +nullable
	Enabled *bool  `json:"enabled,omitempty" ini:"enabled"`
	Name    string `json:"name,omitempty" ini:"name,omitempty"`
	// +nullable
	AllowSignUp       *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ClientId          string `json:"client_id,omitempty" ini:"client_id,omitempty"`
	ClientSecret      string `json:"client_secret,omitempty" ini:"client_secret,omitempty"`
	Scopes            string `json:"scopes,omitempty" ini:"scopes,omitempty"`
	AuthUrl           string `json:"auth_url,omitempty" ini:"auth_url,omitempty"`
	TokenUrl          string `json:"token_url,omitempty" ini:"token_url,omitempty"`
	ApiUrl            string `json:"api_url,omitempty" ini:"api_url,omitempty"`
	AllowedDomains    string `json:"allowed_domains,omitempty" ini:"allowed_domains,omitempty"`
	AllowedGroups     string `json:"allowed_groups,omitempty" ini:"allowed_groups,omitempty"`
	RoleAttributePath string `json:"role_attribute_path,omitempty" ini:"role_attribute_path,omitempty"`
	// +nullable
	RoleAttributeStrict *bool `json:"role_attribute_strict,omitempty" ini:"role_attribute_strict,omitempty"`
}

type GrafanaConfigAuthLdap struct {
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	AllowSignUp *bool  `json:"allow_sign_up,omitempty" ini:"allow_sign_up"`
	ConfigFile  string `json:"config_file,omitempty" ini:"config_file,omitempty"`
}

type GrafanaConfigAuthProxy struct {
	// +nullable
	Enabled        *bool  `json:"enabled,omitempty" ini:"enabled"`
	HeaderName     string `json:"header_name,omitempty" ini:"header_name,omitempty"`
	HeaderProperty string `json:"header_property,omitempty" ini:"header_property,omitempty"`
	// +nullable
	AutoSignUp  *bool  `json:"auto_sign_up,omitempty" ini:"auto_sign_up"`
	LdapSyncTtl string `json:"ldap_sync_ttl,omitempty" ini:"ldap_sync_ttl,omitempty"`
	Whitelist   string `json:"whitelist,omitempty" ini:"whitelist,omitempty"`
	Headers     string `json:"headers,omitempty" ini:"headers,omitempty"`
	// +nullable
	EnableLoginToken *bool `json:"enable_login_token,omitempty" ini:"enable_login_token"`
}

type GrafanaConfigDataProxy struct {
	DialTimeout                  *int `json:"dialTimeout,omitempty" ini:"dialTimeout,omitempty"`
	ExpectContinueTimeoutSeconds *int `json:"expect_continue_timeout_seconds,omitempty" ini:"expect_continue_timeout_seconds,omitempty"`
	IdleConnTimeoutSeconds       *int `json:"idle_conn_timeout_seconds,omitempty" ini:"idle_conn_timeout_seconds,omitempty"`
	KeepAliveSeconds             *int `json:"keep_alive_seconds,omitempty" ini:"keep_alive_seconds,omitempty"`
	// +nullable
	Logging            *bool `json:"logging,omitempty" ini:"logging"`
	MaxIdleConnections *int  `json:"max_idle_connections,omitempty" ini:"max_idle_connections,omitempty"`
	MaxConnsPerHost    *int  `json:"max_conns_per_host,omitempty" ini:"max_conns_per_host,omitempty"`
	ResponseLimit      *int  `json:"response_limit,omitempty" ini:"response_limit,omitempty"`
	RowLimit           *int  `json:"row_limit,omitempty" ini:"row_limit,omitempty"`
	// +nullable
	SendUserHeader *bool `json:"send_user_header,omitempty" ini:"send_user_header,omitempty"`
	// +nullable
	Timeout                    *int `json:"timeout,omitempty" ini:"timeout,omitempty"`
	TlsHandshakeTimeoutSeconds *int `json:"tls_handshake_timeout_seconds,omitempty" ini:"tls_handshake_timeout_seconds,omitempty"`
}

type GrafanaConfigAnalytics struct {
	// +nullable
	ReportingEnabled    *bool  `json:"reporting_enabled,omitempty" ini:"reporting_enabled"`
	GoogleAnalyticsUaId string `json:"google_analytics_ua_id,omitempty" ini:"google_analytics_ua_id,omitempty"`
	// +nullable
	CheckForUpdates *bool `json:"check_for_updates,omitempty" ini:"check_for_updates"`
}

type GrafanaConfigDashboards struct {
	// +nullable
	VersionsToKeep           *int   `json:"versions_to_keep,omitempty" ini:"versions_to_keep,omitempty"`
	DefaultHomeDashboardPath string `json:"default_home_dashboard_path,omitempty" ini:"default_home_dashboard_path,omitempty"`
	// Prevents users from setting the dashboard refresh interval to a lower value than a given interval value
	MinRefreshInterval string `json:"min_refresh_interval,omitempty" ini:"min_refresh_interval,omitempty"`
}

type GrafanaConfigSmtp struct {
	// +nullable
	Enabled  *bool  `json:"enabled,omitempty" ini:"enabled"`
	Host     string `json:"host,omitempty" ini:"host,omitempty"`
	User     string `json:"user,omitempty" ini:"user,omitempty"`
	Password string `json:"password,omitempty" ini:"password,omitempty"`
	CertFile string `json:"cert_file,omitempty" ini:"cert_file,omitempty"`
	KeyFile  string `json:"key_file,omitempty" ini:"key_file,omitempty"`
	// +nullable
	SkipVerify   *bool  `json:"skip_verify,omitempty" ini:"skip_verify"`
	FromAddress  string `json:"from_address,omitempty" ini:"from_address,omitempty"`
	FromName     string `json:"from_name,omitempty" ini:"from_name,omitempty"`
	EhloIdentity string `json:"ehlo_identity,omitempty" ini:"ehlo_identity,omitempty"`
}

type GrafanaConfigLive struct {
	// +nullable
	MaxConnections *int   `json:"max_connections,omitempty" ini:"max_connections,omitempty"`
	AllowedOrigins string `json:"allowed_origins,omitempty" ini:"allowed_origins,omitempty"`
}

type GrafanaConfigLog struct {
	Mode    string `json:"mode,omitempty" ini:"mode,omitempty"`
	Level   string `json:"level,omitempty" ini:"level,omitempty"`
	Filters string `json:"filters,omitempty" ini:"filters,omitempty"`
}

type GrafanaConfigLogFrontend struct {
	// +nullable
	Enabled        *bool  `json:"enabled,omitempty" ini:"enabled,omitempty"`
	SentryDsn      string `json:"sentry_dsn,omitempty" ini:"sentry_dsn,omitempty"`
	CustomEndpoint string `json:"custom_endpoint,omitempty" ini:"custom_endpoint,omitempty"`
	SampleRate     string `json:"sample_rate,omitempty" ini:"sample_rate,omitempty"`
	// +nullable
	LogEndpointRequestsPerSecondLimit *int `json:"log_endpoint_requests_per_second_limit,omitempty" ini:"log_endpoint_requests_per_second_limit,omitempty"`
	// +nullable
	LogEndpointBurstLimit *int `json:"log_endpoint_burst_limit,omitempty" ini:"log_endpoint_burst_limit,omitempty"`
}

type GrafanaConfigLogConsole struct {
	Level  string `json:"level,omitempty" ini:"level,omitempty"`
	Format string `json:"format,omitempty" ini:"format,omitempty"`
}

type GrafanaConfigMetrics struct {
	// +nullable
	Enabled           *bool  `json:"enabled,omitempty" ini:"enabled"`
	BasicAuthUsername string `json:"basic_auth_username,omitempty" ini:"basic_auth_username,omitempty"`
	BasicAuthPassword string `json:"basic_auth_password,omitempty" ini:"basic_auth_password,omitempty"`
	// +nullable
	IntervalSeconds *int `json:"interval_seconds,omitempty" ini:"interval_seconds,omitempty"`
}

type GrafanaConfigMetricsGraphite struct {
	Address string `json:"address,omitempty" ini:"address,omitempty"`
	Prefix  string `json:"prefix,omitempty" ini:"prefix,omitempty"`
}

type GrafanaConfigSnapshots struct {
	// +nullable
	ExternalEnabled      *bool  `json:"external_enabled,omitempty" ini:"external_enabled"`
	ExternalSnapshotUrl  string `json:"external_snapshot_url,omitempty" ini:"external_snapshot_url,omitempty"`
	ExternalSnapshotName string `json:"external_snapshot_name,omitempty" ini:"external_snapshot_name,omitempty"`
	// +nullable
	SnapshotRemoveExpired *bool `json:"snapshot_remove_expired,omitempty" ini:"snapshot_remove_expired"`
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
	// +nullable
	Enabled *bool `json:"enabled,omitempty" ini:"enabled"`
	// +nullable
	ExecuteAlerts      *bool  `json:"execute_alerts,omitempty" ini:"execute_alerts"`
	ErrorOrTimeout     string `json:"error_or_timeout,omitempty" ini:"error_or_timeout,omitempty"`
	NodataOrNullvalues string `json:"nodata_or_nullvalues,omitempty" ini:"nodata_or_nullvalues,omitempty"`
	// +nullable
	ConcurrentRenderLimit *int `json:"concurrent_render_limit,omitempty" ini:"concurrent_render_limit,omitempty"`
	// +nullable
	EvaluationTimeoutSeconds *int `json:"evaluation_timeout_seconds,omitempty" ini:"evaluation_timeout_seconds,omitempty"`
	// +nullable
	NotificationTimeoutSeconds *int `json:"notification_timeout_seconds,omitempty" ini:"notification_timeout_seconds,omitempty"`
	// +nullable
	MaxAttempts *int `json:"max_attempts,omitempty" ini:"max_attempts,omitempty"`
}

type GrafanaConfigUnifiedAlerting struct {
	// +nullable
	Enabled           *bool  `json:"enabled,omitempty" ini:"enabled"`
	ExecuteAlerts     *bool  `json:"execute_alerts,omitempty" ini:"execute_alerts"`
	EvaluationTimeout string `json:"evaluation_timeout,omitempty" ini:"evaluation_timeout,omitempty"`
	MaxAttempts       *int   `json:"max_attempts,omitempty" ini:"max_attempts,omitempty"`
	MinInterval       string `json:"min_interval,omitempty" ini:"min_interval,omitempty"`
}

type GrafanaConfigPanels struct {
	// +nullable
	DisableSanitizeHtml *bool `json:"disable_sanitize_html,omitempty" ini:"disable_sanitize_html"`
}

type GrafanaConfigPlugins struct {
	// +nullable
	// Set to true if you want to test alpha plugins that are not yet ready for general usage. Default is false.
	EnableAlpha *bool `json:"enable_alpha,omitempty" ini:"enable_alpha"`
	// Enter a comma-separated list of plugin identifiers to identify plugins to load even if they are unsigned. Plugins with modified signatures are never loaded.
	// We do not recommend using this option. For more information, refer to https://grafana.com/docs/grafana/next/administration/plugin-management/#plugin-signatures
	AllowLoadingUnsignedPlugins string `json:"allow_loading_unsigned_plugins,omitempty" ini:"allow_loading_unsigned_plugins"`
	// +nullable
	// Available to Grafana administrators only, enables installing / uninstalling / updating plugins directly from the Grafana UI. Set to true by default. Setting it to false will hide the install / uninstall / update controls.
	// For more information, refer to https://grafana.com/docs/grafana/next/administration/plugin-management/#plugin-catalog
	PluginAdminEnabled *bool `json:"plugin_admin_enabled" ini:"plugin_admin_enabled"`
	// Custom install/learn more URL for enterprise plugins. Defaults to https://grafana.com/grafana/plugins/.
	PluginCatalogURL string `json:"plugin_catalog_url,omitempty" ini:"plugin_catalog_url"`
	// Enter a comma-separated list of plugin identifiers to hide in the plugin catalog.
	PluginCatalogHiddenPlugins string `json:"plugin_catalog_hidden_plugins,omitempty" ini:"plugin_catalog_hidden_plugins"`
}

type GrafanaConfigRendering struct {
	ServerURL   string `json:"server_url,omitempty" ini:"server_url"`
	CallbackURL string `json:"callback_url,omitempty" ini:"callback_url"`
	// +nullable
	ConcurrentRenderRequestLimit *int `json:"concurrent_render_request_limit,omitempty" ini:"concurrent_render_request_limit,omitempty"`
}

type GrafanaConfigFeatureToggles struct {
	Enable string `json:"enable,omitempty" ini:"enable,omitempty"`
}

// GrafanaStatus defines the observed state of Grafana
type GrafanaStatus struct {
	Phase               StatusPhase `json:"phase,omitempty"`
	PreviousServiceName string      `json:"previousServiceName,omitempty"`
	Message             string      `json:"message,omitempty"`
	// +nullable
	InstalledDashboards []*GrafanaDashboardRef `json:"dashboards,omitempty"`
	// +nullable
	InstalledPlugins PluginList `json:"installedPlugins,omitempty"`
	// +nullable
	FailedPlugins PluginList `json:"failedPlugins,omitempty"`
}

// GrafanaPlugin contains information about a single plugin
// +k8s:openapi-gen=true
type GrafanaPlugin struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Version string `json:"version"`
}

// Grafana is the Schema for the grafanas API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Grafana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GrafanaSpec   `json:"spec,omitempty"`
	Status GrafanaStatus `json:"status,omitempty"`
}

// GrafanaList contains a list of Grafana
// +kubebuilder:object:root=true
type GrafanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Grafana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Grafana{}, &GrafanaList{})
}

func (cr *Grafana) GetPreferServiceValue() bool {
	if cr.Spec.Client != nil && cr.Spec.Client.PreferService != nil {
		return *cr.Spec.Client.PreferService
	}
	return false
}

func (cr *Grafana) GetScheme() v1.URIScheme {
	if cr.Spec.Config.Server != nil && cr.Spec.Config.Server.Protocol == "https" {
		return v1.URISchemeHTTPS
	}
	return v1.URISchemeHTTP
}
