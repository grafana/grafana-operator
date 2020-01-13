package connectors

type KeystoneConnectorConfig struct {
	Cloud                string            `yaml:"cloud"`
	Domain               string            `yaml:"domain"`
	Host                 string            `yaml:"host"`
	AdminUsername        string            `yaml:"adminUsername"`
	AdminPassword        string            `yaml:"adminPassword"`
	AdminUserDomainName  string            `yaml:"adminUserDomain"`
	AdminProject         string            `yaml:"adminProject"`
	AdminDomain          string            `yaml:"adminDomain"`
	Prompt               string            `yaml:"prompt"`
	AuthScope            AuthScope         `yaml:"authScope,omitempty"`
	IncludeRolesInGroups *bool             `yaml:"includeRolesInGroups,omitempty"`
	RoleNameFormat       string            `yaml:"roleNameFormat,omitempty"`
	GroupNameFormat      string            `yaml:"groupNameFormat,omitempty"`
	RoleMap              map[string]string `yaml:"roleMap,omitempty"`
}

type AuthScope struct {
	ProjectID   string `yaml:"projectID,omitempty"`
	ProjectName string `yaml:"projectName,omitempty"`
	DomainID    string `yaml:"domainID,omitempty"`
	DomainName  string `yaml:"domainName,omitempty"`
}
