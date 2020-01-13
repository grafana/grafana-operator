package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"text/template"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	"github.com/integr8ly/grafana-operator/pkg/api"
	"github.com/integr8ly/grafana-operator/pkg/api/connectors"
	"github.com/integr8ly/grafana-operator/pkg/api/models"
	"github.com/integr8ly/grafana-operator/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	clientID               = "grafana"
	defaultProxyNameFormat = "%s-auth"
	grafanaNameFormat      = "%s-grafana"
)

var (
	logger      = logf.Log.WithName("api_create_grafana")
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

//NewCreateGrafanas creates grafana crd
func NewCreateGrafanas(rt *api.Runtime) operations.CreateGrafanaHandler {
	rand.Seed(time.Now().UnixNano())

	opts, _ := openstack.AuthOptionsFromEnv()
	opts.AllowReauth = true
	opts.Scope = &gophercloud.AuthScope{
		ProjectName: opts.TenantName,
		DomainName:  os.Getenv("OS_PROJECT_DOMAIN_NAME"),
	}
	serviceType := "keystone"
	eo := gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic}
	eo.ApplyDefaults(serviceType)

	pc, _ := openstack.AuthenticatedClient(opts)
	url, _ := pc.EndpointLocator(eo)
	client := gophercloud.ServiceClient{
		ProviderClient: pc,
		Endpoint:       url,
	}
	return &createGrafana{rt, &client}
}

type createGrafana struct {
	*api.Runtime
	client *gophercloud.ServiceClient
}

func (d *createGrafana) Handle(params operations.CreateGrafanaParams, principal *models.Principal) middleware.Responder {
	gr := params.Body

	err := d.mergeRequestGrafanaWith(gr)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}

	g, err := d.paramsToGrafana(principal)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}
	r := &models.CreateGrafanaCreatedBody{
		Hostname: g.Spec.Ingress.Hostname,
		Name:     g.ObjectMeta.Name,
	}
	if err = d.Client.Create(params.HTTPRequest.Context(), &g); err != nil {
		if errors.IsAlreadyExists(err) {
			return operations.NewCreateGrafanaCreated().WithPayload(r)
		}
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}

	return operations.NewCreateGrafanaCreated().WithPayload(r)
}

func (d *createGrafana) paramsToGrafana(p *models.Principal) (g v1alpha1.Grafana, err error) {
	gc := d.Runtime.Config.Grafana
	if gc.Name == nil {
		gc.Name = &p.Account
	}
	if gc.Namespace == nil {
		gc.Namespace = &p.Account
	}
	m := v1.ObjectMeta{
		Name:      fmt.Sprintf(grafanaNameFormat, *gc.Name),
		Namespace: *gc.Namespace,
	}
	if gc.AuthProxy.ClientSecret == "" {
		gc.AuthProxy.ClientSecret = createSecretString(15)
	}
	if gc.Config.Hostname == "" {
		return g, fmt.Errorf("No Grafana hostname provided")
	}

	proxyName := fmt.Sprintf(defaultProxyNameFormat, *gc.Name)
	proxyHost := fmt.Sprintf(gc.Config.Hostname, proxyName)
	gc.Config.Hostname = fmt.Sprintf(gc.Config.Hostname, *gc.Name)

	connectors, err := d.createProxyConnector(proxyHost, p)
	if err != nil {
		return
	}
	cYaml, err := yaml.Marshal(connectors)
	if err != nil {
		return
	}

	g = v1alpha1.Grafana{
		ObjectMeta: m,
		Status:     v1alpha1.GrafanaStatus{},
		Spec: v1alpha1.GrafanaSpec{
			Ingress: v1alpha1.GrafanaIngress{
				Hostname:   gc.Config.Hostname,
				TLSEnabled: false,
				Enabled:    true,
			},
			DashboardLabelSelector: []*v1.LabelSelector{{
				MatchExpressions: []v1.LabelSelectorRequirement{{
					Key:      "app",
					Values:   []string{"grafana"},
					Operator: v1.LabelSelectorOpIn,
				}},
			}},
			Config: v1alpha1.GrafanaConfig{
				Security: v1alpha1.GrafanaConfigSecurity{
					AdminUser:     gc.Config.AdminUser,
					AdminPassword: gc.Config.AdminPassword,
				},
				Users: v1alpha1.GrafanaConfigUsers{
					AutoAssignOrg:     gc.Config.AutoAssignOrg,
					AutoAssignOrgRole: gc.Config.AutoAssignOrgRole,
				},

				AuthGenericOauth: v1alpha1.GrafanaConfigAuthGenericOauth{
					Enabled:      gc.AuthProxy.Enabled,
					ClientId:     clientID,
					Scopes:       "groups openid email",
					ClientSecret: gc.AuthProxy.ClientSecret,
					AuthUrl:      fmt.Sprintf("http://%s/auth", proxyHost),
					TokenUrl:     fmt.Sprintf("http://%s/token", proxyHost),
					GroupRoleMap: gc.Config.GrafanaGroupRoleMap,
				},
				Server: v1alpha1.GrafanaConfigServer{
					RootUrl: fmt.Sprintf("http://%s/", gc.Config.Hostname),
				},
				Auth: v1alpha1.GrafanaConfigAuth{
					DisableLoginForm:   true,
					DisableSignoutMenu: false,
					OauthAutoLogin:     true,
				},

				AuthAnonymous: v1alpha1.GrafanaConfigAuthAnonymous{
					Enabled: false,
				},
			},
			AuthProxy: v1alpha1.GrafanaAuthProxy{
				Enabled:      gc.AuthProxy.Enabled,
				Host:         proxyHost,
				ClientSecret: gc.AuthProxy.ClientSecret,
				ClientID:     clientID,
				Connectors:   string(cYaml),
			},
		},
	}
	return
}

func (d *createGrafana) createProxyConnector(proxyHost string, p *models.Principal) (c []connectors.GrafanaAuthProxyConnector, err error) {
	for _, cn := range d.Runtime.Config.Grafana.AuthProxy.Connectors {
		switch cn {
		case "keystone":
			k := connectors.KeystoneConnectorConfig{}
			if err = yamlUnmarshalHandler(d.Config.AuthProxy.Connectors[cn], &k); err != nil {
				return
			}
			k.Domain = p.Domain
			k.AuthScope.ProjectID = p.Account
			k.RoleNameFormat = "%s"
			gc := connectors.GrafanaAuthProxyConnector{
				Type:   cn,
				ID:     cn,
				Name:   "Converged Cloud",
				Config: k,
			}
			c = append(c, gc)
		}
	}
	return
}

func (d *createGrafana) createApplicationCredentials(p *models.Principal) (ac *applicationcredentials.ApplicationCredential, err error) {
	createOpts := applicationcredentials.CreateOpts{
		Name:   p.Account,
		Secret: createSecretString(15),
		Roles: []applicationcredentials.Role{
			applicationcredentials.Role{Name: "monitoring_viewer"},
		},
		AccessRules: []applicationcredentials.AccessRule{
			{
				Method:  "GET",
				Service: "maia",
			},
		},
		Unrestricted: false,
	}

	ac, err = applicationcredentials.Create(d.client, p.ID, createOpts).Extract()
	fmt.Println(ac, err)
	return
}

func (d *createGrafana) mergeRequestGrafanaWith(with *models.Grafana) (err error) {
	withByte, err := json.Marshal(with)
	if err != nil {
		return
	}
	err = json.Unmarshal(withByte, &d.Runtime.Config.Grafana)
	if err != nil {
		return
	}
	return
}

func createSecretString(n int) string {
	r := make([]rune, n)
	for i := range r {
		r[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(r)
}

func yamlUnmarshalHandler(in interface{}, out interface{}) error {
	var tpl bytes.Buffer
	h, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	t, err := template.New("config").Parse(string(h))
	err = t.Execute(&tpl, nil)

	return yaml.Unmarshal(h, out)
}
