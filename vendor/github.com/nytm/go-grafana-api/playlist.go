package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type PlaylistItem struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Order int    `json:"order"`
	Title string `json:"title"`
}

type Playlist struct {
	Id       int            `json:"id"`
	Name     string         `json:"name"`
	Interval string         `json:"interval"`
	Items    []PlaylistItem `json:"items"`
}

func (c *Client) Playlist(id int) (*Playlist, error) {
	path := fmt.Sprintf("/api/playlists/%d", id)
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

	playlist := &Playlist{}

	err = json.Unmarshal(data, &playlist)
	if err != nil {
		return nil, err
	}

	return playlist, nil
}

func (c *Client) NewPlaylist(playlist Playlist) (int, error) {
	data, err := json.Marshal(playlist)
	if err != nil {
		return 0, err
	}

	req, err := c.newRequest("POST", "/api/playlists", nil, bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, errors.New(resp.Status)
	}

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	result := struct {
		Id int
	}{}

	err = json.Unmarshal(data, &result)
	if err != nil {
		return 0, err
	}

	return result.Id, nil
}

func (c *Client) UpdatePlaylist(playlist Playlist) error {
	path := fmt.Sprintf("/api/playlists/%d", playlist.Id)
	data, err := json.Marshal(playlist)
	if err != nil {
		return err
	}

	req, err := c.newRequest("PUT", path, nil, bytes.NewBuffer(data))
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

func (c *Client) DeletePlaylist(id int) error {
	path := fmt.Sprintf("/api/playlists/%d", id)
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
