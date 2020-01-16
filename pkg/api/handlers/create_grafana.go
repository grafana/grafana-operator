package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/applicationcredentials"
	"github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
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

type createGrafana struct {
	*api.Runtime
	oc *gophercloud.ServiceClient
}

//NewCreateGrafanas creates grafana crd
func NewCreateGrafanas(rt *api.Runtime) operations.CreateGrafanaHandler {
	rand.Seed(time.Now().UnixNano())

	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		logger.Error(err, "Openstack auth failed")
		os.Exit(1)
	}
	opts.AllowReauth = true
	opts.Scope = &gophercloud.AuthScope{
		ProjectName: opts.TenantName,
		DomainName:  os.Getenv("OS_PROJECT_DOMAIN_NAME"),
	}
	serviceType := "keystone"
	eo := gophercloud.EndpointOpts{Availability: gophercloud.AvailabilityPublic}
	eo.ApplyDefaults(serviceType)

	_, err = openstack.AuthenticatedClient(opts)
	if err != nil {
		logger.Error(err, "Openstack auth failed")
		//os.Exit(1)
	}
	//url, _ := pc.EndpointLocator(eo)
	client := gophercloud.ServiceClient{
		ProviderClient: nil,
		Endpoint:       "url",
		Type:           serviceType,
	}
	return &createGrafana{rt, &client}
}

func (d *createGrafana) Handle(params operations.CreateGrafanaParams, principal *models.Principal) middleware.Responder {
	gr := params.Body

	grafana, err := d.mergeRequestGrafanaWith(gr)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}
	d.validateGrafanaData(&grafana, principal)

	g, err := d.paramsToGrafana(&grafana, principal)
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
	if err = d.handleProxy(&grafana, params, principal); err != nil {
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}

	return operations.NewCreateGrafanaCreated().WithPayload(r)
}

func (d *createGrafana) validateGrafanaData(gc *models.Grafana, p *models.Principal) (*models.Grafana, error) {
	if gc.Name == nil {
		gc.Name = &p.Account
	}
	if gc.Namespace == nil {
		gc.Namespace = &p.Account
	}

	if gc.AuthProxy.ClientSecret == "" {
		gc.AuthProxy.ClientSecret = createSecretString(15)
	}
	if gc.Config.Hostname == "" {
		return gc, fmt.Errorf("No Grafana hostname provided")
	}
	return gc, nil
}

func (d *createGrafana) paramsToGrafana(gc *models.Grafana, p *models.Principal) (g v1alpha1.Grafana, err error) {
	proxyHost, err := getProxyHost(gc)
	if err != nil {
		return
	}
	hostname := fmt.Sprintf(gc.Config.Hostname, *gc.Name)

	m := v1.ObjectMeta{
		Name:      fmt.Sprintf(grafanaNameFormat, *gc.Name),
		Namespace: *gc.Namespace,
	}

	g = v1alpha1.Grafana{
		ObjectMeta: m,
		Status:     v1alpha1.GrafanaStatus{},
		Spec: v1alpha1.GrafanaSpec{
			Ingress: &v1alpha1.GrafanaIngress{
				Hostname:   hostname,
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
				Security: &v1alpha1.GrafanaConfigSecurity{
					AdminUser:     gc.Config.AdminUser,
					AdminPassword: gc.Config.AdminPassword,
				},
				Users: &v1alpha1.GrafanaConfigUsers{
					AutoAssignOrg:     &gc.Config.AutoAssignOrg,
					AutoAssignOrgRole: gc.Config.AutoAssignOrgRole,
				},

				AuthGenericOauth: &v1alpha1.GrafanaConfigAuthGenericOauth{
					Enabled:      &gc.AuthProxy.Enabled,
					ClientId:     clientID,
					Scopes:       "groups openid email",
					ClientSecret: gc.AuthProxy.ClientSecret,
					AuthUrl:      fmt.Sprintf("http://%s/auth", proxyHost),
					TokenUrl:     fmt.Sprintf("http://%s/token", proxyHost),
					GroupRoleMap: gc.Config.GrafanaGroupRoleMap,
				},
				Server: &v1alpha1.GrafanaConfigServer{
					RootUrl: fmt.Sprintf("http://%s/", hostname),
				},
				Auth: &v1alpha1.GrafanaConfigAuth{
					DisableLoginForm:   newTrue(),
					DisableSignoutMenu: newFalse(),
					OauthAutoLogin:     newTrue(),
				},

				AuthAnonymous: &v1alpha1.GrafanaConfigAuthAnonymous{
					Enabled: newFalse(),
				},
			},
			Compat: &v1alpha1.GrafanaCompat{},
		},
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

	ac, err = applicationcredentials.Create(d.oc, p.ID, createOpts).Extract()
	fmt.Println("-----------------_", ac, err)
	return
}

func (d *createGrafana) mergeRequestGrafanaWith(with *models.Grafana) (g models.Grafana, err error) {
	g = models.Grafana{}
	withByte, err := json.Marshal(with)
	if err != nil {
		return
	}
	err = json.Unmarshal(withByte, &g)
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
