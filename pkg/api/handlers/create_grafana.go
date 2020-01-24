package handlers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/templates"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	clientID          = "grafana"
	grafanaNameFormat = "%s-grafana"
)

var (
	logger = logf.Log.WithName("api_create_grafana")
)

type createGrafana struct {
	*api.Runtime
}

//NewCreateGrafanas creates grafana crd
func NewCreateGrafanas(rt *api.Runtime) operations.CreateGrafanaHandler {
	rand.Seed(time.Now().UnixNano())
	return &createGrafana{rt}
}

func (d *createGrafana) Handle(params operations.CreateGrafanaParams, principal *models.Principal) middleware.Responder {
	gr := params.Body

	err := d.mergeRequestGrafanaWithConfig(gr)
	fmt.Println(gr)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}
	if err = d.validateGrafanaData(gr, principal); err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())

	}

	g, err := d.paramsToGrafana(gr, principal)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}

	if err = createNamespaceIfNotExists(d.Client, g.Namespace); err != nil {
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

	tpl := templates.NewHandler(d.Runtime)
	if err = tpl.CopyDashboards(g.Namespace); err != nil {
		log.Error(err, err.Error())
	}
	if err = tpl.CopyDatasources(g.Namespace, principal); err != nil {
		log.Error(err, err.Error())
	}

	return operations.NewCreateGrafanaCreated().WithPayload(r)
}

func (d *createGrafana) validateGrafanaData(gc *models.Grafana, p *models.Principal) (err error) {
	if gc.Name == nil {
		gc.Name = &p.Account
	}
	if gc.Namespace == nil {
		gc.Namespace = &p.Account
	}

	if gc.Config.AuthProxy.ClientSecret == nil {
		s, err := createSecret([]byte(d.Config.SecretKey), []byte(p.Account))
		if err != nil {
			return err
		}
		gc.Config.AuthProxy.ClientSecret = &s
	}
	return
}

func (d *createGrafana) paramsToGrafana(gc *models.Grafana, p *models.Principal) (g v1alpha1.Grafana, err error) {
	proxyHost, err := getProxyHost(gc.Config.IngressHost, *gc.Name)
	if err != nil {
		return
	}
	hostname := fmt.Sprintf(gc.Config.IngressHost, *gc.Name)

	m := metav1.ObjectMeta{
		Name:      fmt.Sprintf(grafanaNameFormat, *gc.Name),
		Namespace: *gc.Namespace,
	}
	g = v1alpha1.Grafana{
		ObjectMeta: m,
		Status:     v1alpha1.GrafanaStatus{},
		Spec: v1alpha1.GrafanaSpec{
			Ingress: &v1alpha1.GrafanaIngress{
				Hostname:      hostname,
				TLSEnabled:    true,
				Enabled:       true,
				Annotations:   map[string]string{"vice-president": "true"},
				TLSSecretName: "tls-" + strings.ReplaceAll(hostname, ".", "-"),
			},
			DashboardLabelSelector: []*metav1.LabelSelector{{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "app",
					Values:   []string{"grafana"},
					Operator: metav1.LabelSelectorOpIn,
				}},
			}},
			Config: v1alpha1.GrafanaConfig{
				Security: &v1alpha1.GrafanaConfigSecurity{
					AdminUser:     gc.Config.AdminUser,
					AdminPassword: gc.Config.AdminPassword,
				},
				Users: &v1alpha1.GrafanaConfigUsers{},
				AuthGenericOauth: &v1alpha1.GrafanaConfigAuthGenericOauth{
					Enabled:      &gc.Config.AuthProxy.Enabled,
					AllowSignUp:  newTrue(),
					ClientId:     clientID,
					Scopes:       "groups openid email",
					ClientSecret: *gc.Config.AuthProxy.ClientSecret,
					AuthUrl:      fmt.Sprintf("https://%s/auth", proxyHost),
					TokenUrl:     fmt.Sprintf("https://%s/token", proxyHost),
					GroupRoleMap: gc.Config.GrafanaGroupRoleMap,
					OrgName:      gc.Config.OrgName,
				},
				Server: &v1alpha1.GrafanaConfigServer{
					RootUrl: fmt.Sprintf("https://%s/", hostname),
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

func (d *createGrafana) mergeRequestGrafanaWithConfig(request *models.Grafana) (err error) {
	cfgByte, err := json.Marshal(d.Config.Grafana)
	if err != nil {
		return
	}
	err = json.Unmarshal(cfgByte, request)
	if err != nil {
		return
	}
	return
}
