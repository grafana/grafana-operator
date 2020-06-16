package gapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
)

type SearchTeam struct {
	TotalCount int64   `json:"totalCount,omitempty"`
	Teams      []*Team `json:"teams,omitempty"`
	Page       int64   `json:"page,omitempty"`
	PerPage    int64   `json:"perPage,omitempty"`
}

// Team consists of a get response
// It's used in  Add and Update API
type Team struct {
	Id          int64  `json:"id,omitempty"`
	OrgId       int64  `json:"orgId,omitempty"`
	Name        string `json:"name"`
	Email       string `json:"email,omitempty"`
	AvatarUrl   string `json:"avatarUrl,omitempty"`
	MemberCount int64  `json:"memberCount,omitempty"`
	Permission  int64  `json:"permission,omitempty"`
}

// TeamMember
type TeamMember struct {
	OrgId      int64  `json:"orgId,omitempty"`
	TeamId     int64  `json:"teamId,omitempty"`
	UserId     int64  `json:"userId,omitempty"`
	Email      string `json:"email,omitempty"`
	Login      string `json:"login,omitempty"`
	AvatarUrl  string `json:"avatarUrl,omitempty"`
	Permission int64  `json:"permission,omitempty"`
}

type Preferences struct {
	Theme           string `json:"theme"`
	HomeDashboardId int64  `json:"homeDashboardId"`
	Timezone        string `json:"timezone"`
}

func (c *Client) SearchTeam(query string) (*SearchTeam, error) {
	var result SearchTeam

	page := "1"
	perPage := "1000"
	path := "/api/teams/search"
	queryValues := url.Values{}
	queryValues.Set("page", page)
	queryValues.Set("perPage", perPage)
	queryValues.Set("query", query)

	req, err := c.newRequest("GET", path, queryValues, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Team(id int64) (*Team, error) {
	var team Team

	req, err := c.newRequest("GET", fmt.Sprintf("/api/teams/%d", id), nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &team); err != nil {
		return nil, err
	}
	return &team, nil
}

// AddTeam makes a new team
// email arg is an optional value.
// If you don't want to set email, please set "" (empty string).
func (c *Client) AddTeam(name string, email string) error {
	path := fmt.Sprintf("/api/teams")
	team := Team{
		Name:  name,
		Email: email,
	}
	data, err := json.Marshal(team)
	if err != nil {
		return err
	}
	req, err := c.newRequest("POST", path, nil, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *Client) UpdateTeam(id int64, name string, email string) error {
	path := fmt.Sprintf("/api/teams/%d", id)
	team := Team{
		Name: name,
	}
	// add param if email exists
	if email != "" {
		team.Email = email
	}
	data, err := json.Marshal(team)
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
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *Client) DeleteTeam(id int64) error {
	req, err := c.newRequest("DELETE", fmt.Sprintf("/api/teams/%d", id), nil, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *Client) TeamMembers(id int64) ([]*TeamMember, error) {
	members := make([]*TeamMember, 0)

	req, err := c.newRequest("GET", fmt.Sprintf("/api/teams/%d/members", id), nil, nil)
	if err != nil {
		return members, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return members, err
	}
	if resp.StatusCode != 200 {
		return members, fmt.Errorf(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return members, err
	}
	if err := json.Unmarshal(data, &members); err != nil {
		return members, err
	}
	return members, nil
}

func (c *Client) AddTeamMember(id int64, userId int64) error {
	path := fmt.Sprintf("/api/teams/%d/members", id)
	member := TeamMember{UserId: userId}
	data, err := json.Marshal(member)
	if err != nil {
		return err
	}
	req, err := c.newRequest("POST", path, nil, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *Client) RemoveMemberFromTeam(id int64, userId int64) error {
	path := fmt.Sprintf("/api/teams/%d/members/%d", id, userId)

	req, err := c.newRequest("DELETE", path, nil, nil)
	if err != nil {
		return err
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf(resp.Status)
	}
	return nil
}

func (c *Client) TeamPreferences(id int64) (*Preferences, error) {
	var preferences Preferences

	req, err := c.newRequest("GET", fmt.Sprintf("/api/teams/%d/preferences", id), nil, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &preferences); err != nil {
		return nil, err
	}
	return &preferences, nil
}

func (c *Client) UpdateTeamPreferences(id int64, theme string, homeDashboardId int64, timezone string) error {
	path := fmt.Sprintf("/api/teams/%d", id)
	preferences := Preferences{
		Theme:           theme,
		HomeDashboardId: homeDashboardId,
		Timezone:        timezone,
	}
	data, err := json.Marshal(preferences)
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
		return fmt.Errorf(resp.Status)
	}
	return nil
}
