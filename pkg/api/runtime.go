package api

import (
	"context"

	"github.com/integr8ly/grafana-operator/pkg/api/config"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Runtime struct {
	Namespace types.NamespacedName
	Client    client.Client
	ctx       context.Context
	Config    config.Config
}
