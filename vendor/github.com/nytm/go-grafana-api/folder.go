package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Folder struct {
	Id    int64  `json:"id"`
	Uid   string `json:"uid"`
	Title string `json:"title"`
}

func (c *Client) Folders() ([]Folder, error) {
	folders := make([]Folder, 0)

	req, err := c.newRequest("GET", "/api/folders/", nil, nil)
	if err != nil {
		return folders, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return folders, err
	}
	if resp.StatusCode != 200 {
		return folders, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return folders, err
	}
	err = json.Unmarshal(data, &folders)
	return folders, err
}

func (c *Client) Folder(id int64) (*Folder, error) {
	folder := &Folder{}
	req, err := c.newRequest("GET", fmt.Sprintf("/api/folders/id/%d", id), nil, nil)
	if err != nil {
		return folder, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return folder, err
	}
	if resp.StatusCode != 200 {
		return folder, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return folder, err
	}
	err = json.Unmarshal(data, &folder)
	return folder, err
}

func (c *Client) NewFolder(title string) (Folder, error) {
	folder := Folder{}
	dataMap := map[string]string{
		"title": title,
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		return folder, err
	}
	req, err := c.newRequest("POST", "/api/folders", nil, bytes.NewBuffer(data))
	if err != nil {
		return folder, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return folder, err
	}
	if resp.StatusCode != 200 {
		data, _ = ioutil.ReadAll(resp.Body)
		return folder, fmt.Errorf("status: %s body: %s", resp.Status, data)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return folder, err
	}
	err = json.Unmarshal(data, &folder)
	if err != nil {
		return folder, err
	}
	return folder, err
}

func (c *Client) UpdateFolder(id string, name string) error {
	dataMap := map[string]string{
		"name": name,
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}
	req, err := c.newRequest("PUT", fmt.Sprintf("/api/folders/%s", id), nil, bytes.NewBuffer(data))
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
	return err
}

func (c *Client) DeleteFolder(id string) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("/api/folders/%s", id), nil, nil)
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
	return err
}
