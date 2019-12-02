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
	ProcessDashboard(knownHash string) (*sdk.Board, error)
	NewHash() string
}

type DashboardPipelineImpl struct {
	Dashboard      *v1alpha1.GrafanaDashboard
	JSON           string
	Board          sdk.Board
	Logger         logr.Logger
	Hash           string
	FixAnnotations bool
}

func NewDashboardPipeline(dashboard *v1alpha1.GrafanaDashboard, fixAnnotations bool) DashboardPipeline {
	return &DashboardPipelineImpl{
		Dashboard:      dashboard,
		JSON:           "",
		Logger:         logf.Log.WithName(fmt.Sprintf("dashboard-%v", dashboard.Name)),
		FixAnnotations: fixAnnotations,
	}
}

func (r *DashboardPipelineImpl) ProcessDashboard(knownHash string) (*sdk.Board, error) {
	err := r.obtainJson()
	if err != nil {
		return nil, err
	}

	// Dashboard unchanged?
	hash := r.generateHash()
	if hash == knownHash {
		r.Hash = knownHash
		return nil, nil
	}
	r.Hash = hash

	// Dashboard valid?
	err = r.validateJson()
	if err != nil {
		return nil, err
	}

	// Dashboards are never expected to come with an ID, it is
	// always assigned by Grafana. If there is one, we ignore it
	r.Board.ID = 0

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
	dashboardBytes := []byte(r.JSON)
	dashboardBytes, err := r.fixAnnotations(dashboardBytes)
	if err != nil {
		return err
	}

	var dashboard sdk.Board
	err = json.Unmarshal(dashboardBytes, &dashboard)
	if err != nil {
		return err
	}

	r.Board = dashboard
	return nil
}

// Try to get the dashboard json definition either from a provided URL or from the
// raw json in the dashboard resource
func (r *DashboardPipelineImpl) obtainJson() error {
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
func (r *DashboardPipelineImpl) generateHash() string {
	return fmt.Sprintf("%x", md5.Sum([]byte(r.Dashboard.Spec.Json+r.Dashboard.Spec.Url)))
}

// Try to obtain the dashboard json from a provided url
func (r *DashboardPipelineImpl) loadDashboardFromURL() error {
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

func (r *DashboardPipelineImpl) NewHash() string {
	return r.Hash
}

// Some older dashboards provide the tags list of an annotation as an array
// instead of a string
func (r *DashboardPipelineImpl) fixAnnotations(dashboardBytes []byte) ([]byte, error) {
	if !r.FixAnnotations {
		return dashboardBytes, nil
	}

	raw := map[string]interface{}{}
	err := json.Unmarshal(dashboardBytes, &raw)
	if err != nil {
		return nil, err
	}

	if raw != nil && raw["annotations"] != nil {
		annotations := raw["annotations"].(map[string]interface{})
		if annotations != nil && annotations["list"] != nil {
			annotationsList := annotations["list"].([]interface{})
			for _, annotation := range annotationsList {
				rawAnnotation := annotation.(map[string]interface{})
				if rawAnnotation["tags"] != nil {
					// Don't attempty to convert the tags, just replace them
					// with something that is compatible
					rawAnnotation["tags"] = ""
				}
			}
		}
	}

	dashboardBytes, err = json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	return dashboardBytes, nil
}
