package grafanadashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-jsonnet"
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SourceType int

const (
	SourceTypeJson    SourceType = 1
	SourceTypeJsonnet SourceType = 2
	SourceTypeUnknown SourceType = 3
)

var (
	grafanaComDashboardApiUrlRoot string = "https://grafana.com/api/dashboards"
)

type DashboardPipeline interface {
	ProcessDashboard(knownHash string, folderId *int64, folderName string, forceRecreate bool) ([]byte, error)
	NewHash() string
}

type DashboardPipelineImpl struct {
	Client    client.Client
	Dashboard *v1alpha1.GrafanaDashboard
	JSON      string
	Board     map[string]interface{}
	Logger    logr.Logger
	Hash      string
	Context   context.Context
}

func NewDashboardPipeline(client client.Client, dashboard *v1alpha1.GrafanaDashboard, ctx context.Context) DashboardPipeline {
	return &DashboardPipelineImpl{
		Client:    client,
		Dashboard: dashboard,
		JSON:      "",
		Logger:    log.Log.WithName(fmt.Sprintf("dashboard-%v", dashboard.Name)),
		Context:   ctx,
	}
}

func (r *DashboardPipelineImpl) ProcessDashboard(knownHash string, folderId *int64, folderName string, forceRecreate bool) ([]byte, error) {
	err := r.obtainJson()
	if err != nil {
		return nil, err
	}

	// Dashboard unchanged?
	hash := r.Dashboard.Hash()
	if hash == knownHash && !forceRecreate {
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

	// Overwrite in case any user provided uid exists

	r.Board["uid"] = r.Dashboard.UID()
	r.Board["folderId"] = folderId
	r.Board["folderName"] = folderName

	raw, err := json.Marshal(r.Board)
	if err != nil {
		return nil, err
	}

	return bytes.TrimSpace(raw), nil
}

// Make sure the dashboard contains valid JSON
func (r *DashboardPipelineImpl) validateJson() error {
	contents, err := r.Dashboard.Parse(r.JSON)
	r.Board = contents
	return err
}

// Try to get the dashboard json definition either from a provided URL or from the
// raw json in the dashboard resource. The priority is as follows:
// 1) try to fetch from url or grafanaCom if provided
// 2) url or grafanaCom fails or not provided: try to fetch from configmap ref
// 3) no configmap specified: try to use embedded json
// 4) no json specified: try to use embedded jsonnet
func (r *DashboardPipelineImpl) obtainJson() error {
	// TODO(DeanBrunt): Add earlier validation for this
	if r.Dashboard.Spec.Url != "" && r.Dashboard.Spec.GrafanaCom != nil {
		return errors.New("both dashboard url and grafana.com source specified")
	}

	if r.Dashboard.Spec.GrafanaCom != nil {
		if err := r.loadDashboardFromGrafanaCom(); err != nil {
			r.Logger.Error(err, "failed to request dashboard from grafana.com, falling back to config map; if specified")
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.Url != "" {
		err := r.loadDashboardFromURL()
		if err != nil {
			r.Logger.Error(err, "failed to request dashboard url, falling back to config map; if specified")
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

	if r.Dashboard.Spec.Jsonnet != "" {
		json, err := r.loadJsonnet(r.Dashboard.Spec.Jsonnet)
		if err != nil {
			r.Logger.Error(err, "failed to parse jsonnet")
		} else {
			r.JSON = json
			return nil
		}
	}

	return errors.New("unable to obtain dashboard contents")
}

// Compiles jsonnet to json and makes the grafonnet library available to
// the template
func (r *DashboardPipelineImpl) loadJsonnet(source string) (string, error) {
	cfg := config.GetControllerConfig()
	jsonnetLocation := cfg.GetConfigString(config.ConfigJsonnetBasePath, config.JsonnetBasePath)

	vm := jsonnet.MakeVM()

	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{jsonnetLocation},
	})

	return vm.EvaluateSnippet(r.Dashboard.Name, source) // nolint
}

// Try to obtain the dashboard json from a provided url
func (r *DashboardPipelineImpl) loadDashboardFromURL() error {
	url, err := url.ParseRequestURI(r.Dashboard.Spec.Url)
	if err != nil {
		return fmt.Errorf("invalid url %v", r.Dashboard.Spec.Url)
	}

	resp, err := http.Get(r.Dashboard.Spec.Url)
	if err != nil {
		return fmt.Errorf("cannot request %v", r.Dashboard.Spec.Url)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("request failed with status %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	sourceType := r.getFileType(url.Path)

	switch sourceType {
	case SourceTypeJson, SourceTypeUnknown:
		// If unknown, assume json
		r.JSON = string(body)
	case SourceTypeJsonnet:
		json, err := r.loadJsonnet(string(body))
		if err != nil {
			return err
		}
		r.JSON = json
	}

	// Update dashboard spec so that URL would not be refetched
	if r.JSON != r.Dashboard.Spec.Json {
		r.Dashboard.Spec.Json = r.JSON
		err := r.Client.Update(r.Context, r.Dashboard)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DashboardPipelineImpl) loadDashboardFromGrafanaCom() error {
	url, err := r.getGrafanaComDashboardUrl()
	if err != nil {
		return fmt.Errorf("failed to get grafana.com dashboard url: %w", err)
	}

	resp, err := http.Get(url) // nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to request dashboard url '%s': %w", url, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	r.JSON = string(body)

	// Update JSON so dashboard is not refetched
	if r.JSON != r.Dashboard.Spec.Json {
		r.Dashboard.Spec.Json = r.JSON
		if err := r.Client.Update(r.Context, r.Dashboard); err != nil {
			return err
		}
	}

	return nil
}

func (r *DashboardPipelineImpl) getGrafanaComDashboardUrl() (string, error) {
	grafanaComSource := r.Dashboard.Spec.GrafanaCom
	var revision int
	if grafanaComSource.Revision == nil {
		var err error
		revision, err = r.getLatestRevisionForGrafanaComDashboard()
		if err != nil {
			return "", fmt.Errorf("failed to get latest revision for dashboard id %d: %w", r.Dashboard.Spec.GrafanaCom.Id, err)
		}
	} else {
		revision = *grafanaComSource.Revision
	}

	u, err := url.Parse(grafanaComDashboardApiUrlRoot)
	if err != nil {
		return "", err
	}

	u.Path = path.Join(u.Path, strconv.Itoa(grafanaComSource.Id), "revisions", strconv.Itoa(revision), "download")
	return u.String(), nil
}

func (r *DashboardPipelineImpl) getLatestRevisionForGrafanaComDashboard() (int, error) {
	u, err := url.Parse(grafanaComDashboardApiUrlRoot)
	if err != nil {
		return 0, err
	}

	u.Path = path.Join(u.Path, strconv.Itoa(r.Dashboard.Spec.GrafanaCom.Id), "revisions")
	resp, err := http.Get(u.String())
	if err != nil {
		return 0, fmt.Errorf("failed to make request to %s: %w", u.String(), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("request to list available grafana dashboard revisions failed with status code '%d'", resp.StatusCode)
	}

	listResponse, err := r.unmarshalListDashboardRevisionsResponseBody(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to unmarshal list dashboard revisions response: %w", err)
	}

	if listResponse == nil || len(listResponse.Items) == 0 {
		return 0, errors.New("list dashboard revisions request succeeded but no revisions returned")
	}

	return r.getMaximumRevisionFromListDashboardRevisionsResponse(listResponse), nil
}

// This will attempt to discover the latest revision, initially by using knowledge of the
// default sort method (ordering by revision), falling back on our own manual discovery of
// the maximum if the default changes away from revision ordering or using an unhandled
// direction.
func (r *DashboardPipelineImpl) getMaximumRevisionFromListDashboardRevisionsResponse(resp *listDashboardRevisionsResponse) int {
	if resp.OrderBy == "revision" {
		if resp.Direction == "asc" {
			return resp.Items[len(resp.Items)-1].Revision
		}

		if resp.Direction == "desc" {
			return resp.Items[0].Revision
		}
	}

	var maxRevision int
	for _, item := range resp.Items {
		if maxRevision < item.Revision {
			maxRevision = item.Revision
		}
	}

	return maxRevision
}

// This is an incomplete representation of the expected response,
// including only fields we care about.
type listDashboardRevisionsResponse struct {
	Items     []dashboardRevisionItem `json:"items"`
	OrderBy   string                  `json:"orderBy"`
	Direction string                  `json:"direction"`
}

type dashboardRevisionItem struct {
	Revision int `json:"revision"`
}

func (r *DashboardPipelineImpl) unmarshalListDashboardRevisionsResponseBody(
	body io.Reader,
) (*listDashboardRevisionsResponse, error) {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	resp := &listDashboardRevisionsResponse{}
	if err := json.Unmarshal(bodyBytes, resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal raw json to list dashboard revisions response: %w", err)
	}

	return resp, nil
}

// Try to determine the type (json or grafonnet) or a remote file by looking
// at the filename extension
func (r *DashboardPipelineImpl) getFileType(path string) SourceType {
	fragments := strings.Split(path, ".")
	if len(fragments) == 0 {
		return SourceTypeUnknown
	}

	extension := strings.TrimSpace(fragments[len(fragments)-1])
	switch strings.ToLower(extension) {
	case "json":
		return SourceTypeJson
	case "grafonnet":
		return SourceTypeJsonnet
	case "jsonnet":
		return SourceTypeJsonnet
	default:
		return SourceTypeUnknown
	}
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
			msg := "invalid datasource input rule, input or datasource empty"
			r.Logger.Info(msg)
			return errors.New(msg)
		}

		searchValue := fmt.Sprintf("${%s}", input.InputName)
		currentJson = strings.ReplaceAll(currentJson, searchValue, input.DatasourceName)
		r.Logger.Info("resolving input", "input.InputName", input.InputName, "input.DatasourceName", input.DatasourceName)
	}

	r.JSON = currentJson
	return nil
}
