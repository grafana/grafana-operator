package handlers

import (
	"fmt"

	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *createGrafana) handleProxy(grafana *models.Grafana, params operations.CreateGrafanaParams, principal *models.Principal) (err error) {
	if !grafana.AuthProxy.Enabled {
		return
	}
	gp, err := d.paramsToGrafanaProxy(grafana, principal)
	if err != nil {
		log.Error(err, err.Error())
		return err
	}
	if err = d.Client.Create(params.HTTPRequest.Context(), &gp); err != nil {
		return
	}
	return
}

func (d *createGrafana) paramsToGrafanaProxy(gc *models.Grafana, p *models.Principal) (g v1alpha1.GrafanaProxy, err error) {
	proxyHost, err := getProxyHost(gc)
	hostname := fmt.Sprintf(gc.Config.Hostname, *gc.Name)
	con, err := d.createProxyConnector(gc, p)
	if err != nil {
		return
	}
	m := v1.ObjectMeta{
		Name:      fmt.Sprintf(grafanaNameFormat, *gc.Name),
		Namespace: *gc.Namespace,
	}
	g = v1alpha1.GrafanaProxy{
		ObjectMeta: m,
		Status:     v1alpha1.GrafanaProxyStatus{},
		Spec: v1alpha1.GrafanaProxySpec{
			Config: v1alpha1.GrafanaProxyConfig{
				HostName: hostname,
				Issuer:   "http://" + proxyHost,
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
							"http://" + hostname + "/login/generic_oauth",
						},
						Secret:       gc.AuthProxy.ClientSecret,
						TrustedPeers: []string{},
					},
				},
				Connectors: con,
			},
		},
	}
	return
}

func (d *createGrafana) createProxyConnector(gc *models.Grafana, p *models.Principal) (c []v1alpha1.GrafanaProxyConnector, err error) {
	for _, cn := range gc.AuthProxy.Connectors {
		switch cn {
		case "keystone":
			fmt.Println(d.Config.AuthProxy.Connectors[cn])
			/*
				k := connectors.KeystoneConnectorConfig{}
				if err = yamlUnmarshalHandler(d.Config.AuthProxy.Connectors[cn], &k); err != nil {
					return
				}
			*/
			cfg := d.Config.AuthProxy.Connectors[cn]
			cfg["Domain"] = p.Domain
			cfg["AuthScope:ProjectID"] = p.Account
			cfg["RoleNameFormat"] = "%s"

			cn := v1alpha1.GrafanaProxyConnector{
				Type:   cn,
				ID:     cn,
				Name:   "Converged Cloud",
				Config: d.Config.AuthProxy.Connectors[cn],
			}
			c = append(c, cn)
		}
	}
	fmt.Println(c)
	return
}
