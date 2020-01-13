package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/integr8ly/grafana-operator/pkg/api"
	"github.com/integr8ly/grafana-operator/pkg/api/models"
	"github.com/integr8ly/grafana-operator/pkg/api/rest/operations"
	"github.com/sapcc/kubernikus/pkg/version"
)

func NewInfo(rt *api.Runtime) operations.InfoHandler {
	return &info{rt}
}

type info struct {
	*api.Runtime
}

func (d *info) Handle(params operations.InfoParams) middleware.Responder {
	info := &models.Info{
		Version: version.GitCommit,
	}
	return operations.NewInfoOK().WithPayload(info)
}
