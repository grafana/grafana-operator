package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	grafanaProxyNameFormat = "%s-grafana-proxy"
)

type createGrafanaProxy struct {
	*api.Runtime
}

//NewCreateGrafanaProxies creates grafana crd
func NewCreateGrafanaProxies(rt *api.Runtime) operations.CreateGrafanaProxyHandler {
	return &createGrafanaProxy{rt}
}

func (d *createGrafanaProxy) Handle(params operations.CreateGrafanaProxyParams, principal *models.Principal) middleware.Responder {
	g := params.Body
	err := d.mergeRequestGrafanaProxyWithConfig(g)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaProxyDefault{}, 500, err.Error())
	}
	if g.Name == nil {
		g.Name = &principal.Account
	}
	if g.Namespace == nil {
		g.Namespace = &principal.Account
	}
	if err = createNamespaceIfNotExists(d.Client, *g.Namespace); err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaDefault{}, 500, err.Error())
	}

	gp, err := d.paramsToGrafanaProxy(g, principal)
	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.CreateGrafanaProxyDefault{}, 500, err.Error())
	}
	if err = d.Client.Create(params.HTTPRequest.Context(), &gp); err != nil {
		return NewErrorResponse(&operations.CreateGrafanaProxyDefault{}, 500, err.Error())
	}
	r := &models.CreateGrafanaProxyCreatedBody{
		Hostname: gp.Spec.Config.Hostname,
		Name:     gp.ObjectMeta.Name,
	}
	return operations.NewCreateGrafanaProxyCreated().WithPayload(r)
}

func (d *createGrafanaProxy) paramsToGrafanaProxy(g *models.GrafanaProxy, p *models.Principal) (gp v1alpha1.GrafanaProxy, err error) {
	proxyHost, err := getProxyHost(g.Config.IngressHost, *g.Name)
	con, err := d.createProxyConnector(g, p)
	if err != nil {
		return
	}
	if g.Config.ClientSecret == nil {
		s, err := createSecret([]byte(d.Config.SecretKey), []byte(p.Account))
		if err != nil {
			return gp, err
		}
		g.Config.ClientSecret = &s
	}
	grafanaHostname := fmt.Sprintf(g.Config.IngressHost, *g.Name)

	m := metav1.ObjectMeta{
		Name:      fmt.Sprintf(grafanaProxyNameFormat, *g.Name),
		Namespace: *g.Namespace,
	}
	gp = v1alpha1.GrafanaProxy{
		ObjectMeta: m,
		Status:     v1alpha1.GrafanaProxyStatus{},
		Spec: v1alpha1.GrafanaProxySpec{
			Config: v1alpha1.GrafanaProxyConfig{
				Hostname: proxyHost,
				Issuer:   "https://" + proxyHost,
				Storage: v1alpha1.Storage{
					Type: "kubernetes",
					Config: v1alpha1.StorageConfig{
						InCluster: true,
					},
				},
				Web: v1alpha1.Web{
					HTTP:           "0.0.0.0:80",
					AllowedOrigins: []string{},
				},
				Frontend: v1alpha1.WebConfig{
					Theme:  "ccloud",
					Issuer: "Converged Cloud Grafana",
				},
				Expiry: v1alpha1.Expiry{
					SigningKeys: "5m",
					IDTokens:    "10m",
				},
				Logger: v1alpha1.Logger{
					Level: "debug",
				},
				OAuth2: v1alpha1.OAuth2{
					SkipApprovalScreen: false,
					ResponseTypes:      []string{"code", "token", "id_token"},
				},
				StaticClients: []v1alpha1.Client{
					{
						ID:   "grafana",
						Name: "Grafana UI",
						RedirectURIs: []string{
							"https://" + grafanaHostname + "/login/generic_oauth",
						},
						Secret:       *g.Config.ClientSecret,
						TrustedPeers: []string{},
					},
				},
				Connectors: con,
			},
		},
	}
	return
}

func (d *createGrafanaProxy) mergeRequestGrafanaProxyWithConfig(request *models.GrafanaProxy) (err error) {
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

func (d *createGrafanaProxy) createProxyConnector(g *models.GrafanaProxy, p *models.Principal) (kc []v1alpha1.KeystoneConnector, err error) {
	for _, cn := range g.Config.Connectors {
		switch cn {
		case "keystone":

			k := v1alpha1.KeystoneConnector{}
			if err = yamlUnmarshalHandler(d.Config.AuthProxy.Connectors[cn], &k); err != nil {
				return
			}
			s := v1alpha1.AuthScope{
				ProjectID: p.Account,
			}
			k.Config.Domain = p.Domain
			k.Config.AuthScope = s
			k.Config.RoleNameFormat = "%s"
			kc = append(kc, k)
		}
	}
	return
}
