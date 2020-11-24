package common

import (
	"context"
	"fmt"
	grafanav1alpha1 "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/config"
	"github.com/integr8ly/grafana-operator/v3/pkg/controller/model"
	v12 "github.com/openshift/api/route/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type LokiState struct {
	LokiService    *v1.Service
	LokiDeployment *v13.Deployment
	LokiConfigMap  *v1.ConfigMap
	LokiRoute      *v12.Route
	LokiIngress    *v1beta1.Ingress
	LokiDataSource  *grafanav1alpha1.GrafanaDataSource
}

func NewLokiState() *LokiState{
	return &LokiState{}
}


func (i *LokiState) Read(ctx context.Context,cr *grafanav1alpha1.Loki,client client.Client){
	cfg := config.GetControllerConfig()
	//TODO remove me, I'm only here to prevent unused error
	fmt.Print(cfg)

	//isOpenshift := cfg.GetConfigBool(config.ConfigOpenshift,false)
	//TODO add remaining reads
}


func(i *LokiState) readLokiService(ctx context.Context, cr *grafanav1alpha1.Loki,client client.Client) error{
	currentState := &v1.Service{}
	selector := model.LokiServiceSelector(cr)
	if err := client.Get(ctx,selector,currentState); err != nil{
		if errors.IsNotFound(err){
			return err
		}
	}
	i.LokiService = currentState.DeepCopy()
	return nil
}