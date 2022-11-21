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
	"time"

	"github.com/go-logr/logr"
	"github.com/google/go-jsonnet"
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/grafana-operator/grafana-operator/v4/controllers/config"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type SourceType int

const (
	SourceTypeJson    SourceType = 1
	SourceTypeJsonnet SourceType = 2
	SourceTypeUnknown SourceType = 3
)

var grafanaComDashboardApiUrlRoot string = "https://grafana.com/api/dashboards"

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
//  0. try to use previously fetched content from url or grafanaCom if it is valid
//  1. try to fetch from url or grafanaCom if provided
//     1.1) if downloaded content is identical to spec.json, clear spec.json to clean up from fetch behavior pre 4.5.0
//  2. url or grafanaCom fails or not provided: try to fetch from configmap ref
//  3. no configmap specified: try to use embedded json
//  4. no json specified: try to use embedded jsonnet
func (r *DashboardPipelineImpl) obtainJson() error {
	var returnErr error

	if r.Dashboard.Spec.GrafanaCom != nil {
		url, err := r.getGrafanaComDashboardUrl()
		if err != nil {
			return fmt.Errorf("failed to get grafana.com dashboard url: %w", err)
		}
		if err := r.loadDashboardFromURL(url); err != nil {
			returnErr = fmt.Errorf("failed to request dashboard from grafana.com, falling back to url; if specified: %w", err)
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.Url != "" {
		err := r.loadDashboardFromURL(r.Dashboard.Spec.Url)
		if err != nil {
			returnErr = fmt.Errorf("failed to request dashboard url, falling back to config map; if specified: %w", err)
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.ConfigMapRef != nil {
		err := r.loadDashboardFromConfigMap(r.Dashboard.Spec.ConfigMapRef, false)
		if err != nil {
			returnErr = fmt.Errorf("failed to get config map, falling back to raw json: %w", err)
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.GzipConfigMapRef != nil {
		err := r.loadDashboardFromConfigMap(r.Dashboard.Spec.GzipConfigMapRef, true)
		if err != nil {
			returnErr = fmt.Errorf("failed to get config map, falling back to raw json: %w", err)
		} else {
			return nil
		}
	}

	if r.Dashboard.Spec.GzipJson != nil {
		jsonBytes, err := v1alpha1.Gunzip(r.Dashboard.Spec.GzipJson)
		if err != nil {
			returnErr = fmt.Errorf("failed to decode/decompress gzipped json: %w", err)
		} else {
			r.JSON = string(jsonBytes)
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
			returnErr = fmt.Errorf("failed to parse jsonnet: %w", err)
		} else {
			r.JSON = json
			return nil
		}
	}

	return returnErr
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
func (r *DashboardPipelineImpl) loadDashboardFromURL(source string) error {
	r.JSON = r.Dashboard.GetContentCache(source)
	if r.JSON != "" {
		return nil
	}

	url, err := url.ParseRequestURI(source)
	if err != nil {
		return fmt.Errorf("invalid url %v", source)
	}

	resp, err := http.Get(url.String())
	if err != nil {
		return fmt.Errorf("cannot request %v", source)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		retries := 0
		r.refreshDashboard()
		if r.Dashboard.Status.Error != nil {
			retries = r.Dashboard.Status.Error.Retries
		}
		r.Dashboard.Status = v1alpha1.GrafanaDashboardStatus{
			Error: &v1alpha1.GrafanaDashboardError{
				Message: string(body),
				Code:    resp.StatusCode,
				Retries: retries + 1,
			},
			ContentTimestamp: metav1.Time{Time: time.Now()},
		}

		if err := r.Client.Status().Update(r.Context, r.Dashboard); err != nil {
			return fmt.Errorf("failed to request dashboard: %s\nfailed to update status : %w", string(body), err)
		}

		return fmt.Errorf("request failed with status %v", resp.StatusCode)
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

	if r.Dashboard.Spec.Json == r.JSON {
		// Content downloaded to `json` field pre 4.5.0 can be removed since it is identical to the downloaded content.
		r.refreshDashboard()
		r.Dashboard.Spec.Json = ""
		err = r.Client.Update(r.Context, r.Dashboard)
		if err != nil {
			return err
		}
	}

	content, err := v1alpha1.Gzip(r.JSON)
	if err != nil {
		return err
	}

	r.refreshDashboard()
	r.Dashboard.Status = v1alpha1.GrafanaDashboardStatus{
		ContentCache:     content,
		ContentTimestamp: metav1.Time{Time: time.Now()},
		ContentUrl:       source,
	}

	if err := r.Client.Status().Update(r.Context, r.Dashboard); err != nil {
		if !k8serrors.IsConflict(err) {
			return fmt.Errorf("failed to update status with content for dashboard %s/%s: %w", r.Dashboard.Namespace, r.Dashboard.Name, err)
		}
	}

	return nil
}

func (r *DashboardPipelineImpl) refreshDashboard() {
	err := r.Client.Get(r.Context, types.NamespacedName{Name: r.Dashboard.Name, Namespace: r.Dashboard.Namespace}, r.Dashboard)
	if err != nil {
		r.Logger.V(1).Error(err, "refreshing dashboard generation failed")
	}
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
		// resp.Direction seems to be inverted in the response (as of 2022-05-09), so let's ignore it and grab the bigger value
		first := resp.Items[0].Revision
		last := resp.Items[len(resp.Items)-1].Revision
		if first > last {
			return first
		} else {
			return last
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
func (r *DashboardPipelineImpl) loadDashboardFromConfigMap(ref *corev1.ConfigMapKeySelector, binaryCompressed bool) error {
	ctx := context.Background()
	objectKey := client.ObjectKey{Name: ref.Name, Namespace: r.Dashboard.Namespace}

	var cm corev1.ConfigMap
	err := r.Client.Get(ctx, objectKey, &cm)
	if err != nil {
		return err
	}

	if binaryCompressed {
		jsonBytes, err := v1alpha1.Gunzip(cm.BinaryData[ref.Key])
		if err != nil {
			return err
		}
		r.JSON = string(jsonBytes)
	} else {
		r.JSON = cm.Data[ref.Key]
	}

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
