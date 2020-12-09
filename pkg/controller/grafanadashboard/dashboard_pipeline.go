package grafanadashboard

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type DashboardPipeline interface {
	ProcessDashboard(knownHash string) ([]byte, error)
	NewHash() string
}

type DashboardPipelineImpl struct {
	Client         client.Client
	Dashboard      *v1alpha1.GrafanaDashboard
	JSON           string
	Board          map[string]interface{}
	Logger         logr.Logger
	Hash           string
	FixAnnotations bool
	FixHeights     bool
}

func NewDashboardPipeline(client client.Client, dashboard *v1alpha1.GrafanaDashboard, fixAnnotations bool, fixHeights bool) DashboardPipeline {
	return &DashboardPipelineImpl{
		Client:         client,
		Dashboard:      dashboard,
		JSON:           "",
		Logger:         logf.Log.WithName(fmt.Sprintf("dashboard-%v", dashboard.Name)),
		FixAnnotations: fixAnnotations,
		FixHeights:     fixHeights,
	}
}

func (r *DashboardPipelineImpl) ProcessDashboard(knownHash string) ([]byte, error) {
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

	// Datasource inputs to resolve?
	err = r.resolveDatasources()
	if err != nil {
		return nil, err
	}

	// Dashboard valid?
	err = r.validateJson()
	if err != nil {
		return nil, err
	}

	// Dashboards are never expected to come with an ID, it is
	// always assigned by Grafana. If there is one, we ignore it
	r.Board["id"] = nil

	// This dashboard has previously been imported
	// To make sure its updated we have to set the metadata
	if r.Dashboard.Status.Phase == v1alpha1.PhaseReconciling {
		r.Board["slug"] = r.Dashboard.Status.Slug
		r.Board["uid"] = r.Dashboard.Status.UID
	}

	raw, err := json.Marshal(r.Board)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(raw), nil
}

// Make sure the dashboard contains valid JSON
func (r *DashboardPipelineImpl) validateJson() error {
	dashboardBytes := []byte(r.JSON)
	dashboardBytes, err := r.fixAnnotations(dashboardBytes)
	if err != nil {
		return err
	}

	dashboardBytes, err = r.addEnvToAlert(dashboardBytes)
	if err != nil {
		return err
	}

	dashboardBytes, err = r.addEnvToDedupkey(dashboardBytes)
	if err != nil {
		return err
	}

	dashboardBytes, err = r.fixHeights(dashboardBytes)
	if err != nil {
		return err
	}

	r.Board = make(map[string]interface{})
	return json.Unmarshal(dashboardBytes, &r.Board)
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

	if r.Dashboard.Spec.ConfigMapRef != nil {
		err := r.loadDashboardFromConfigMap()
		if err != nil {
			r.Logger.Error(err, "failed to get config map, falling back to raw json")
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
	var datasources strings.Builder
	for _, input := range r.Dashboard.Spec.Datasources {
		datasources.WriteString(input.DatasourceName)
		datasources.WriteString(input.InputName)
	}

	return fmt.Sprintf("%x", md5.Sum([]byte(
		r.Dashboard.Spec.Json+r.Dashboard.Spec.Url+datasources.String()+r.Dashboard.Spec.Environment)))
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

// Try to obtain the dashboard json from a config map
func (r *DashboardPipelineImpl) loadDashboardFromConfigMap() error {
	ctx := context.Background()
	objectKey := client.ObjectKey{Name: r.Dashboard.Spec.ConfigMapRef.Name, Namespace: r.Dashboard.Namespace}

	var cm corev1.ConfigMap
	err := r.Client.Get(ctx, objectKey, &cm)
	if err != nil {
		return err
	}

	r.JSON = cm.Data[r.Dashboard.Spec.ConfigMapRef.Key]

	return nil
}

func (r *DashboardPipelineImpl) NewHash() string {
	return r.Hash
}

func (r *DashboardPipelineImpl) resolveDatasources() error {
	if len(r.Dashboard.Spec.Datasources) == 0 {
		return nil
	}

	currentJson := r.JSON
	for _, input := range r.Dashboard.Spec.Datasources {
		if input.DatasourceName == "" || input.InputName == "" {
			msg := fmt.Sprintf("invalid datasource input rule, input or datasource empty")
			r.Logger.Info(msg)
			return errors.New(msg)
		}

		searchValue := fmt.Sprintf("${%s}", input.InputName)
		currentJson = strings.ReplaceAll(currentJson, searchValue, input.DatasourceName)
		r.Logger.Info(fmt.Sprintf("resolving input %s to %s", input.InputName, input.DatasourceName))
	}

	r.JSON = currentJson
	return nil
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

// Some dashboards have a height property encoded as a number where the SDK expects a string
func (r *DashboardPipelineImpl) fixHeights(dashboardBytes []byte) ([]byte, error) {
	if !r.FixHeights {
		return dashboardBytes, nil
	}

	raw := map[string]interface{}{}
	err := json.Unmarshal(dashboardBytes, &raw)
	if err != nil {
		return nil, err
	}

	if raw != nil && raw["panels"] != nil {
		panels := raw["panels"].([]interface{})
		for _, panel := range panels {
			rawPanel := panel.(map[string]interface{})
			if rawPanel["height"] != nil {
				rawPanel["height"] = fmt.Sprintf("%v", rawPanel["height"])
			}
		}
	}

	dashboardBytes, err = json.Marshal(raw)
	if err != nil {
		return nil, err
	}

	return dashboardBytes, nil
}

// add cluster env to alert message
func (r *DashboardPipelineImpl) addEnvToAlert(dashboardBytes []byte) ([]byte, error) {
	if r.Dashboard.Spec.Environment == "" {
		return dashboardBytes, nil
	}

	raw := map[string]interface{}{}
	err := json.Unmarshal(dashboardBytes, &raw)
	if err != nil {
		return nil, err
	}

	if raw != nil && raw["panels"] != nil {
		panelList, ok := raw["panels"].([]interface{})
		if panelList != nil && ok {
			for _, p := range panelList {
				panel, ok := p.(map[string]interface{})
				if panel != nil && panel["alert"] != nil && ok {
					alert, ok := panel["alert"].(map[string]interface{})
					if alert != nil && ok {
						if alert["message"] != nil {
							message := alert["message"].(string)
							alert["message"] = r.Dashboard.Spec.Environment + " " + message
							r.Logger.Info(fmt.Sprintf("updating alert message %v", alert["message"]))
						}
						if alert["name"] != nil {
							name := alert["name"].(string)
							alert["name"] = r.Dashboard.Spec.Environment + " " + name
							r.Logger.Info(fmt.Sprintf("updating alert name %v", alert["name"]))
						}
					}
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

// add cluster env to PD dedupkey
func (r *DashboardPipelineImpl) addEnvToDedupkey(dashboardBytes []byte) ([]byte, error) {
	if r.Dashboard.Spec.Environment == "" {
		return dashboardBytes, nil
	}

	raw := map[string]interface{}{}
	err := json.Unmarshal(dashboardBytes, &raw)
	if err != nil {
		return nil, err
	}

	if raw != nil && raw["panels"] != nil {
		panelList, ok := raw["panels"].([]interface{})
		if panelList != nil && ok {
			for _, p := range panelList {
				panel, ok := p.(map[string]interface{})
				if panel != nil && panel["alert"] != nil && ok {
					alert, ok := panel["alert"].(map[string]interface{})
					if alert != nil && alert["alertRuleTags"] != nil && ok {
						tags, ok := alert["alertRuleTags"].(map[string]interface{})
						if tags != nil && tags["dedup_key"] != nil && ok {
							dedupKey := tags["dedup_key"].(string)
							tags["dedup_key"] = r.Dashboard.Spec.Environment + "-" + dedupKey
							r.Logger.Info(fmt.Sprintf("updating dedup key %v", tags["dedup_key"]))
						}
					}
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
