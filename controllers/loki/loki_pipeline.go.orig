package loki

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type LokiPipeline interface {
	ProcessLoki(loki v1alpha1.Loki) ([]byte, error)
}

func NewLokiPipeline(loki *v1alpha1.Loki, ctx context.Context) LokiPipeline {
	return &LokiPipelineImpl{
		Loki:    loki,
		Logger:  logf.Log.WithName(fmt.Sprintf("loki-%v", loki.Name)),
		Context: ctx,
	}
}

type LokiPipelineImpl struct {
	Loki    *v1alpha1.Loki
	Board   map[string]interface{}
	Logger  logr.Logger
	Context context.Context
}

func (r *LokiPipelineImpl) ProcessLoki(loki v1alpha1.Loki) ([]byte, error) {

	//TODO add remaining fields
	r.Board["ExternalURL"] = loki.Spec.External

	raw, err := json.Marshal(r.Loki)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(raw), nil
}
