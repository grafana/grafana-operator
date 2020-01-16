package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/integr8ly/grafana-operator/v3/pkg/api"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/api/rest/operations"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("api_list_grafana")

//NewListGrafanas creates
func NewListGrafanas(rt *api.Runtime) operations.ListGrafanasHandler {
	return &listGrafanas{rt}
}

type listGrafanas struct {
	*api.Runtime
}

func (d *listGrafanas) Handle(params operations.ListGrafanasParams, principal *models.Principal) middleware.Responder {
	gl := &v1alpha1.GrafanaList{}

	err := d.Client.List(params.HTTPRequest.Context(), gl)

	if err != nil {
		log.Error(err, err.Error())
		return NewErrorResponse(&operations.ListGrafanasDefault{}, 500, err.Error())
	}

	grafanas := make([]*models.Grafana, 0)
	for _, g := range gl.Items {
		grafanas = append(grafanas, grafanaFromCRD(&g))
	}

	return operations.NewListGrafanasOK().WithPayload(grafanas)

}

func accountSelector(principal *models.Principal) labels.Selector {
	return labels.SelectorFromSet(map[string]string{"account": principal.Account})
}
