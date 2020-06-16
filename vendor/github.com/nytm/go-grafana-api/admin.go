package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

func (c *Client) CreateUser(user User) (int64, error) {
	id := int64(0)
	data, err := json.Marshal(user)
	if err != nil {
		return id, err
	}

	req, err := c.newRequest("POST", "/api/admin/users", nil, bytes.NewBuffer(data))
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
	created := struct {
		Id int64 `json:"id"`
	}{}
	err = json.Unmarshal(data, &created)
	if err != nil {
		return id, err
	}
	return created.Id, err
}

func (c *Client) DeleteUser(id int64) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("/api/admin/users/%d", id), nil, nil)
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
