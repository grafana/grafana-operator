package grafanadashboard

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/grafana-tools/sdk"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"io/ioutil"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type DashboardPipeline interface {
	ProcessDashboard() (*sdk.Board, error)
}

type DashboardPipelineImpl struct {
	Dashboard *v1alpha1.GrafanaDashboard
	JSON      string
	Board     sdk.Board
	Logger    logr.Logger
	Hash      string
}

func NewDashboardPipeline(dashboard *v1alpha1.GrafanaDashboard) DashboardPipeline {
	return &DashboardPipelineImpl{
		Dashboard: dashboard,
		JSON:      "",
		Logger:    logf.Log.WithName(fmt.Sprintf("dashboard-%v", dashboard.Name)),
	}
}

func (r *DashboardPipelineImpl) ProcessDashboard() (*sdk.Board, error) {
	err := r.obtainJson()
	if err != nil {
		return nil, err
	}

	// Dashboard unchanged?
	r.generateHash()
	if r.Dashboard.Status.Hash == r.Hash {
		return nil, nil
	}

	// Update hash if it actually changed
	r.Dashboard.Status.Hash = r.Hash

	// Dashboard valid?
	err = r.validateJson()
	if err != nil {
		return nil, err
	}

	// This dashboard has previously been imported
	// To make sure its updated we have to set the metadata
	if r.Dashboard.Status.Phase == v1alpha1.PhaseReconciling {
		r.Board.Slug = r.Dashboard.Status.Slug
		r.Board.UID = r.Dashboard.Status.UID
		r.Board.ID = r.Dashboard.Status.ID
	}

	return &r.Board, nil
}

// Make sure the dashboard contains valid JSON
func (r *DashboardPipelineImpl) validateJson() error {
	log.Info("validating dashboard contents")
	var dashboard sdk.Board
	err := json.Unmarshal([]byte(r.JSON), &dashboard)
	if err != nil {
		return err
	}

	r.Board = dashboard
	return nil
}

// Try to get the dashboard json definition either from a provided URL or from the
// raw json in the dashboard resource
func (r *DashboardPipelineImpl) obtainJson() error {
	r.Logger.Info("obtaining dashboard json")

	if r.Dashboard.Spec.Url != "" {
		err := r.loadDashboardFromURL()
		if err != nil {
			r.Logger.Error(err, "failed to request dashboard url, falling back to raw json")
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.Json != "" {
		r.JSON = r.Dashboard.Spec.Json
		return nil
	}

	return errors.New("dashboard does not contain json")
}

// Create a hash of the dashboard to detect if there are actually changes to the json
// If there are no changes we should avoid sending update requests as this will create
// a new dashboard version in Grafana
func (r *DashboardPipelineImpl) generateHash() {
	r.Logger.Info("generating dashboard hash")
	r.Hash = fmt.Sprintf("%x", md5.Sum([]byte(r.Dashboard.Spec.Json+r.Dashboard.Spec.Url)))
}

// Try to obtain the dashboard json from a provided url
func (r *DashboardPipelineImpl) loadDashboardFromURL() error {
	r.Logger.Info("loading dashboard from url")
	_, err := url.ParseRequestURI(r.Dashboard.Spec.Url)
	if err != nil {
		return errors.New(fmt.Sprintf("invalid url %v", r.Dashboard.Spec.Url))
	}

	resp, err := http.Get(r.Dashboard.Spec.Url)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot request %v", r.Dashboard.Spec.Url))
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	r.JSON = string(body)

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("request failed with status %v", resp.StatusCode))
	}

	return nil
}
