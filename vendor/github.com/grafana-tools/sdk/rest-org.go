// +build draft

package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2019 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// CreateOrg creates a new organization.
// It reflects POST /api/orgs API call.
func (r *Client) CreateOrg(org Org) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post("api/orgs", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetAllOrgs returns all organizations.
// It reflects GET /api/orgs API call.
func (r *Client) GetAllOrgs() ([]Org, error) {
	var (
		raw  []byte
		orgs []Org
		code int
		err  error
	)
	if raw, code, err = r.get("api/orgs", nil); err != nil {
		return orgs, err
	}

	if code != http.StatusOK {
		return orgs, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&orgs); err != nil {
		return orgs, fmt.Errorf("unmarshal orgs: %s\n%s", err, raw)
	}
	return orgs, err
}

// GetActualOrg gets current organization.
// It reflects GET /api/org API call.
func (r *Client) GetActualOrg() (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get("api/org", nil); err != nil {
		return org, err
	}
	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// GetOrgById gets organization by organization Id.
// It reflects GET /api/orgs/:orgId API call.
func (r *Client) GetOrgById(oid uint) (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get(fmt.Sprintf("api/orgs/%d", oid), nil); err != nil {
		return org, err
	}

	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// GetOrgByOrgName gets organization by organization name.
// It reflects GET /api/orgs/name/:orgName API call.
func (r *Client) GetOrgByOrgName(name string) (Org, error) {
	var (
		raw  []byte
		org  Org
		code int
		err  error
	)
	if raw, code, err = r.get(fmt.Sprintf("api/orgs/name/%s", name), nil); err != nil {
		return org, err
	}

	if code != http.StatusOK {
		return org, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&org); err != nil {
		return org, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return org, err
}

// UpdateActualOrg updates current organization.
// It reflects PUT /api/org API call.
func (r *Client) UpdateActualOrg(org Org) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put("api/org", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateOrg updates the organization identified by oid.
// It reflects PUT /api/orgs/:orgId API call.
func (r *Client) UpdateOrg(org Org, oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(org); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(fmt.Sprintf("api/orgs/%d", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteOrg deletes the organization identified by the oid.
// Reflects DELETE /api/orgs/:orgId API call.
func (r *Client) DeleteOrg(oid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, _, err = r.delete(fmt.Sprintf("api/orgs/%d", oid)); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetActualOrgUsers get all users within the actual organisation.
// Reflects GET /api/org/users API call.
func (r *Client) GetActualOrgUsers() ([]OrgUser, error) {
	var (
		raw   []byte
		users []OrgUser
		code  int
		err   error
	)
	if raw, code, err = r.get("api/org/users", nil); err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&users); err != nil {
		return nil, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return users, err
}

// GetOrgUsers gets the users for the organization specified by oid.
// Reflects GET /api/orgs/:orgId/users API call.
func (r *Client) GetOrgUsers(oid uint) ([]OrgUser, error) {
	var (
		raw   []byte
		users []OrgUser
		code  int
		err   error
	)
	if raw, code, err = r.get(fmt.Sprintf("api/orgs/%d/users", oid), nil); err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&users); err != nil {
		return nil, fmt.Errorf("unmarshal org: %s\n%s", err, raw)
	}
	return users, err
}

// AddActualOrgUser adds a global user to the current organization.
// Reflects POST /api/org/users API call.
func (r *Client) AddActualOrgUser(userRole UserRole) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(userRole); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post("api/org/users", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateUser updates the existing user.
// Reflects POST /api/org/users/:userId API call.
func (r *Client) UpdateActualOrgUser(user UserRole, uid uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(fmt.Sprintf("api/org/users/%s", uid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteActualOrgUser delete user in actual organization.
// Reflects DELETE /api/org/users/:userId API call.
func (r *Client) DeleteActualOrgUser(uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(fmt.Sprintf("api/org/users/%d", uid)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// AddUserToOrg add user to organization with oid.
// Reflects POST /api/orgs/:orgId/users API call.
func (r *Client) AddOrgUser(user UserRole, oid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(fmt.Sprintf("api/orgs/%d/users", oid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// UpdateOrgUser updates the user specified by uid within the organization specified by oid.
// Reflects PATCH /api/orgs/:orgId/users/:userId API call.
func (r *Client) UpdateOrgUser(user UserRole, oid, uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, err = json.Marshal(user); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.patch(fmt.Sprintf("api/orgs/%d/users/%d", oid, uid), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// DeleteOrgUser deletes the user specified by uid within the organization specified by oid.
// Reflects DELETE /api/orgs/:orgId/users/:userId API call.
func (r *Client) DeleteOrgUser(oid, uid uint) (StatusMessage, error) {
	var (
		raw   []byte
		reply StatusMessage
		err   error
	)
	if raw, _, err = r.delete(fmt.Sprintf("api/orgs/%d/users/%d", oid, uid)); err != nil {
		return StatusMessage{}, err
	}
	err = json.Unmarshal(raw, &reply)
	return reply, err
}

// UpdateActualOrgPreferences updates preferences of the actual organization.
// Reflects PUT /api/org/preferences API call.
func (r *Client) UpdateActualOrgPreferences(prefs Preferences) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(prefs); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put("api/org/preferences/", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetActualOrgPreferences gets preferences of the actual organization.
// It reflects GET /api/org/preferences API call.
func (r *Client) GetActualOrgPreferences() (Preferences, error) {
	var (
		raw  []byte
		pref Preferences
		code int
		err  error
	)
	if raw, code, err = r.get("/api/org/preferences", nil); err != nil {
		return pref, err
	}

	if code != http.StatusOK {
		return pref, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&pref); err != nil {
		return pref, fmt.Errorf("unmarshal prefs: %s\n%s", err, raw)
	}
	return pref, err
}
