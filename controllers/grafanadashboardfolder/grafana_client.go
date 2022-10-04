package grafanadashboardfolder

import (
	"bytes"
	"crypto/md5" //nolint
	"encoding/json"
	"fmt"
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"io"
	"net/http"
	"net/url"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"
)

const (
	CreateOrUpdateFolderUrl = "%v/api/folders"
	FolderPermissionsUrl    = "%v/api/folders/%v/permissions"
)

type GrafanaFolderPermissionsResponse struct {
	ID      *int64 `json:"id"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type GrafanaFolderResponse struct {
	ID    *int64 `json:"id"`
	Title string `json:"title"`
	UID   string `json:"uid"`
}

type GrafanaFolderRequest struct {
	Title string `json:"title"`
	UID   string `json:"uid"`
}

type GrafanaClient interface {
	FindOrCreateFolder(folderName string) (GrafanaFolderResponse, error)
	ApplyFolderPermissions(folderName string, folderPermissions []*v1alpha1.GrafanaPermissionItem) (GrafanaFolderPermissionsResponse, error)
}

type GrafanaClientImpl struct {
	url      string
	user     string
	password string
	client   *http.Client
}

func setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "grafana-operator")
}

func NewGrafanaClient(url, user, password string, transport *http.Transport, timeoutSeconds time.Duration) GrafanaClient {
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * timeoutSeconds,
	}

	return &GrafanaClientImpl{
		url:      url,
		user:     user,
		password: password,
		client:   client,
	}
}

var logger = logf.Log.WithName("folder-grafana-client")

func (r *GrafanaClientImpl) getAllFolders() ([]GrafanaFolderResponse, error) {
	rawURL := fmt.Sprintf(CreateOrUpdateFolderUrl, r.url)
	parsed, err := url.Parse(rawURL)

	if err != nil {
		return nil, err
	}

	parsed.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("GET", parsed.String(), nil)

	if err != nil {
		return nil, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		// Grafana might be unavailable, no reason to panic, other checks are in place
		if resp.StatusCode == 503 {
			return nil, nil
		} else {
			return nil, fmt.Errorf(
				"error getting folders, expected status 200 but got %v",
				resp.StatusCode)
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var folders []GrafanaFolderResponse
	err = json.Unmarshal(data, &folders)

	return folders, err
}

func newFolderResponse() GrafanaFolderResponse {
	var id int64 = 0

	return GrafanaFolderResponse{
		ID: &id,
	}
}

func (r *GrafanaClientImpl) FindOrCreateFolder(folderName string) (GrafanaFolderResponse, error) {
	response := newFolderResponse()

	existingFolders, err := r.getAllFolders()
	if err != nil {
		return response, err
	}

	for _, folder := range existingFolders {
		if strings.EqualFold(folder.Title, folderName) {
			return folder, nil
		}
	}

	rawURL := fmt.Sprintf(CreateOrUpdateFolderUrl, r.url)
	apiUrl, err := url.Parse(rawURL)
	if err != nil {
		return response, err
	}

	raw, err := json.Marshal(GrafanaFolderRequest{
		Title: folderName,
		UID:   buildFolderUidFromName(folderName),
	})
	if err != nil {
		return response, err
	}

	apiUrl.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("POST", apiUrl.String(), bytes.NewBuffer(raw))
	if err != nil {
		return response, err
	}

	setHeaders(req)

	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 503 {
			return GrafanaFolderResponse{}, nil
		} else {
			return response, fmt.Errorf(
				"error creating folder, expected status 200 but got %v",
				resp.StatusCode)
		}
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}

	err = json.Unmarshal(data, &response)

	return response, err
}

func buildFolderUidFromName(folderName string) string {
	// uid must not exceed 40 chars
	return fmt.Sprintf("%x", md5.Sum([]byte(folderName))) //nolint
}

func (r *GrafanaClientImpl) ApplyFolderPermissions(folderName string, folderPermissions []*v1alpha1.GrafanaPermissionItem) (GrafanaFolderPermissionsResponse, error) {
	response := GrafanaFolderPermissionsResponse{}

	// ensure folder exists - permissions can only be applied to existing folders, so we need its UID
	existingFolder, err := r.FindOrCreateFolder(folderName)
	if err != nil {
		return response, err
	}

	rawURL := fmt.Sprintf(FolderPermissionsUrl, r.url, existingFolder.UID)
	apiUrl, err := url.Parse(rawURL)
	if err != nil {
		return response, err
	}

	requestBody := buildFolderPermissionRequestBody(folderPermissions)
	apiUrl.User = url.UserPassword(r.user, r.password)
	req, err := http.NewRequest("POST", apiUrl.String(), bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return response, err
	}

	setHeaders(req)
	resp, err := r.client.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 503 {
		return response, nil
	} else if resp.StatusCode != 200 {
		logger.V(1).Info(fmt.Sprintf("used request-data: url '%s' with body '%s'", rawURL, requestBody)) // use rawURL instead of apiUrl to not expose credentials in log
		return response, fmt.Errorf("error setting folder-permissions, expected status 200 but got %v", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = json.Unmarshal(responseBody, &response)

	return response, err
}

// buildFolderPermissionRequestBody creates a JSON-String according to the Spec of the HTTP-API:
// https://grafana.com/docs/grafana/latest/developers/http_api/folder_permissions/#update-permissions-for-a-folder
// Due to the rather dynamic key-names, that must be done programmatically rather than via marshaling an object
func buildFolderPermissionRequestBody(folderPermissions []*v1alpha1.GrafanaPermissionItem) string {
	var b strings.Builder
	b.WriteString("{ \"items\": [ ")

	// TODO: support other targetTypes (teamId and userId) - value-type must be integer then
	for i, item := range folderPermissions {
		fmt.Fprintf(&b, "{%q: %q, \"permission\": %d}", item.PermissionTargetType, item.PermissionTarget, item.PermissionLevel)
		if i+1 < len(folderPermissions) {
			b.WriteString(",")
		}
	}

	b.WriteString(" ]}")
	return b.String()
}
