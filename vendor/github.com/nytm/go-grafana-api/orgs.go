package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type Org struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *Client) Orgs() ([]Org, error) {
	orgs := make([]Org, 0)

	req, err := c.newRequest("GET", "/api/orgs/", nil, nil)
	if err != nil {
		return orgs, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return orgs, err
	}
	if resp.StatusCode != 200 {
		return orgs, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return orgs, err
	}
	err = json.Unmarshal(data, &orgs)
	return orgs, err
}

func (c *Client) OrgByName(name string) (Org, error) {
	org := Org{}
	req, err := c.newRequest("GET", fmt.Sprintf("/api/orgs/name/%s", name), nil, nil)
	if err != nil {
		return org, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return org, err
	}
	if resp.StatusCode != 200 {
		return org, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return org, err
	}
	err = json.Unmarshal(data, &org)
	return org, err
}

func (c *Client) Org(id int64) (Org, error) {
	org := Org{}
	req, err := c.newRequest("GET", fmt.Sprintf("/api/orgs/%d", id), nil, nil)
	if err != nil {
		return org, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return org, err
	}
	if resp.StatusCode != 200 {
		return org, errors.New(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return org, err
	}
	err = json.Unmarshal(data, &org)
	return org, err
}

func (c *Client) NewOrg(name string) (int64, error) {
	id := int64(0)

	dataMap := map[string]string{
		"name": name,
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		return id, err
	}
	req, err := c.newRequest("POST", "/api/orgs", nil, bytes.NewBuffer(data))
	if err != nil {
		return id, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return id, err
	}
	if resp.StatusCode != 200 {
		return id, errors.New(resp.Status)
	}
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return id, err
	}
	tmp := struct {
		Id int64 `json:"orgId"`
	}{}
	err = json.Unmarshal(data, &tmp)
	if err != nil {
		return id, err
	}
	id = tmp.Id
	return id, err
}

func (c *Client) UpdateOrg(id int64, name string) error {
	dataMap := map[string]string{
		"name": name,
	}
	data, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}
	req, err := c.newRequest("PUT", fmt.Sprintf("/api/orgs/%d", id), nil, bytes.NewBuffer(data))
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

func (c *Client) DeleteOrg(id int64) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("/api/orgs/%d", id), nil, nil)
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
