package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	client2 "github.com/grafana/grafana-operator/v5/controllers/client"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type CompleteReconciler struct {
	client client.Client
}

func NewCompleteReconciler(client client.Client) reconcilers.OperatorGrafanaReconciler {
	return &CompleteReconciler{
		client: client,
	}
}

func (r *CompleteReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, _ *v1beta1.OperatorReconcileVars, _ *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("CompleteReconciler")

	log.V(1).Info("fetching Grafana version from instance")
	version, err := r.getVersion(ctx, cr)
	if err != nil {
		cr.Status.Version = ""
		return v1beta1.OperatorStageResultFailed, fmt.Errorf("failed fetching version from instance: %w", err)
	}

	cr.Status.Version = version
	log.V(1).Info("reconciliation completed")

	return v1beta1.OperatorStageResultSuccess, nil
}

func (r *CompleteReconciler) getVersion(ctx context.Context, cr *v1beta1.Grafana) (string, error) {
	cl, err := client2.NewHTTPClient(ctx, r.client, cr)
	if err != nil {
		return "", fmt.Errorf("setup of the http client: %w", err)
	}

	gURL, err := client2.ParseAdminURL(cr.Status.AdminURL)
	if err != nil {
		return "", err
	}

	instanceURL := gURL.JoinPath("/frontend/settings").String()
	req, err := http.NewRequest(http.MethodGet, instanceURL, nil)
	if err != nil {
		return "", fmt.Errorf("building request to fetch version: %w", err)
	}

	err = client2.InjectAuthHeaders(context.Background(), r.client, cr, req)
	if err != nil {
		return "", fmt.Errorf("fetching credentials for version detection: %w", err)
	}

	resp, err := cl.Do(req)
	if err != nil {
		return "", err
	}

	data := struct {
		BuildInfo struct {
			Version string `json:"version"`
		} `json:"buildInfo"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", fmt.Errorf("parsing health endpoint data: %w", err)
	}
	if data.BuildInfo.Version == "" {
		return "", fmt.Errorf("empty version received from server")
	}

	return data.BuildInfo.Version, nil
}
