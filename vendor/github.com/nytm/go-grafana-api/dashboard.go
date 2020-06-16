package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
)

type DashboardMeta struct {
	IsStarred bool   `json:"isStarred"`
	Slug      string `json:"slug"`
	Folder    int64  `json:"folderId"`
}

type DashboardSaveResponse struct {
	Slug    string `json:"slug"`
	Id      int64  `json:"id"`
	Uid     string `json:"uid"`
	Status  string `json:"status"`
	Version int64  `json:"version"`
}

type DashboardSearchResponse struct {
	Id          uint     `json:"id"`
	Uid         string   `json:"uid"`
	Title       string   `json:"title"`
	Uri         string   `json:"uri"`
	Url         string   `json:"url"`
	Slug        string   `json:"slug"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	IsStarred   bool     `json:"isStarred"`
	FolderId    uint     `json:"folderId"`
	FolderUid   string   `json:"folderUid"`
	FolderTitle string   `json:"folderTitle"`
	FolderUrl   string   `json:"folderUrl"`
}

type Dashboard struct {
	Meta      DashboardMeta          `json:"meta"`
	Model     map[string]interface{} `json:"dashboard"`
	Folder    int64                  `json:"folderId"`
	Overwrite bool                   `json:"overwrite"`
}

// Deprecated: use NewDashboard instead
func (c *Client) SaveDashboard(model map[string]interface{}, overwrite bool) (*DashboardSaveResponse, error) {
	wrapper := map[string]interface{}{
		"dashboard": model,
		"overwrite": overwrite,
	}
	data, err := json.Marshal(wrapper)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/db", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		data, _ = ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("status: %d, body: %s", resp.StatusCode, data)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) NewDashboard(dashboard Dashboard) (*DashboardSaveResponse, error) {
	data, err := json.Marshal(dashboard)
	if err != nil {
		return nil, err
	}
	req, err := c.newRequest("POST", "/api/dashboards/db", nil, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &DashboardSaveResponse{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) Dashboards() ([]DashboardSearchResponse, error) {
	dashboards := make([]DashboardSearchResponse, 0)
	query := url.Values{}
	// search only dashboards
	query.Add("type", "dash-db")
	req, err := c.newRequest("GET", "/api/search", query, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return dashboards, err
	}
	if resp.StatusCode != 200 {
		return dashboards, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return dashboards, err
	}

	err = json.Unmarshal(data, &dashboards)
	return dashboards, err
}

// Deprecated: Starting from Grafana v5.0. Please update to use DashboardByUID instead.
func (c *Client) Dashboard(slug string) (*Dashboard, error) {
	return c.dashboard(fmt.Sprintf("/api/dashboards/db/%s", slug))
}

func (c *Client) DashboardByUID(uid string) (*Dashboard, error) {
	return c.dashboard(fmt.Sprintf("/api/dashboards/uid/%s", uid))
}

func (c *Client) dashboard(path string) (*Dashboard, error) {
	req, err := c.newRequest("GET", path, nil, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	result := &Dashboard{}
	err = json.Unmarshal(data, &result)
	result.Folder = result.Meta.Folder
	if os.Getenv("GF_LOG") != "" {
		log.Printf("got back dashboard response  %s", data)
	}
	return result, err
}

// Deprecated: Starting from Grafana v5.0. Please update to use DeleteDashboardByUID instead.
func (c *Client) DeleteDashboard(slug string) error {
	return c.deleteDashboard(fmt.Sprintf("/api/dashboards/db/%s", slug))
}

func (c *Client) DeleteDashboardByUID(uid string) error {
	return c.deleteDashboard(fmt.Sprintf("/api/dashboards/uid/%s", uid))
}

func (c *Client) deleteDashboard(path string) error {
	req, err := c.newRequest("DELETE", path, nil, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return nil
}
