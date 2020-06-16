package gapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

type DataSource struct {
	Id     int64  `json:"id,omitempty"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	URL    string `json:"url"`
	Access string `json:"access"`

	Database string `json:"database,omitempty"`
	User     string `json:"user,omitempty"`
	// Deprecated in favor of secureJsonData.password
	Password string `json:"password,omitempty"`

	OrgId     int64 `json:"orgId,omitempty"`
	IsDefault bool  `json:"isDefault"`

	BasicAuth     bool   `json:"basicAuth"`
	BasicAuthUser string `json:"basicAuthUser,omitempty"`
	// Deprecated in favor of secureJsonData.basicAuthPassword
	BasicAuthPassword string `json:"basicAuthPassword,omitempty"`

	JSONData       JSONData       `json:"jsonData,omitempty"`
	SecureJSONData SecureJSONData `json:"secureJsonData,omitempty"`
}

// JSONData is a representation of the datasource `jsonData` property
type JSONData struct {
	// Used by all datasources
	TlsAuth           bool `json:"tlsAuth,omitempty"`
	TlsAuthWithCACert bool `json:"tlsAuthWithCACert,omitempty"`
	TlsSkipVerify     bool `json:"tlsSkipVerify,omitempty"`

	// Used by Graphite
	GraphiteVersion string `json:"graphiteVersion,omitempty"`

	// Used by Prometheus, Elasticsearch, InfluxDB, MySQL, PostgreSQL and MSSQL
	TimeInterval string `json:"timeInterval,omitempty"`

	// Used by Elasticsearch
	EsVersion       int64  `json:"esVersion,omitempty"`
	TimeField       string `json:"timeField,omitempty"`
	Interval        string `json:"inteval,omitempty"`
	LogMessageField string `json:"logMessageField,omitempty"`
	LogLevelField   string `json:"logLevelField,omitempty"`

	// Used by Cloudwatch
	AuthType                string `json:"authType,omitempty"`
	AssumeRoleArn           string `json:"assumeRoleArn,omitempty"`
	DefaultRegion           string `json:"defaultRegion,omitempty"`
	CustomMetricsNamespaces string `json:"customMetricsNamespaces,omitempty"`

	// Used by OpenTSDB
	TsdbVersion    string `json:"tsdbVersion,omitempty"`
	TsdbResolution string `json:"tsdbResolution,omitempty"`

	// Used by MSSQL
	Encrypt string `json:"encrypt,omitempty"`

	// Used by PostgreSQL
	Sslmode         string `json:"sslmode,omitempty"`
	PostgresVersion int64  `json:"postgresVersion,omitempty"`
	Timescaledb     bool   `json:"timescaledb,omitempty"`

	// Used by MySQL, PostgreSQL and MSSQL
	MaxOpenConns    int64 `json:"maxOpenConns,omitempty"`
	MaxIdleConns    int64 `json:"maxIdleConns,omitempty"`
	ConnMaxLifetime int64 `json:"connMaxLifetime,omitempty"`

	// Used by Prometheus
	HttpMethod   string `json:"httpMethod,omitempty"`
	QueryTimeout string `json:"queryTimeout,omitempty"`

	// Used by Stackdriver
	AuthenticationType string `json:"authenticationType,omitempty"`
	ClientEmail        string `json:"clientEmail,omitempty"`
	DefaultProject     string `json:"defaultProject,omitempty"`
	TokenUri           string `json:"tokenUri,omitempty"`
}

// SecureJSONData is a representation of the datasource `secureJsonData` property
type SecureJSONData struct {
	// Used by all datasources
	TlsCACert         string `json:"tlsCACert,omitempty"`
	TlsClientCert     string `json:"tlsClientCert,omitempty"`
	TlsClientKey      string `json:"tlsClientKey,omitempty"`
	Password          string `json:"password,omitempty"`
	BasicAuthPassword string `json:"basicAuthPassword,omitempty"`

	// Used by Cloudwatch
	AccessKey string `json:"accessKey,omitempty"`
	SecretKey string `json:"secretKey,omitempty"`

	// Used by Stackdriver
	PrivateKey string `json:"privateKey,omitempty"`
}

func (c *Client) NewDataSource(s *DataSource) (int64, error) {
	data, err := json.Marshal(s)
	if err != nil {
		return 0, err
	}
	req, err := c.newRequest("POST", "/api/datasources", nil, bytes.NewBuffer(data))
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
		Id int64 `json:"id"`
	}{}
	err = json.Unmarshal(data, &result)
	return result.Id, err
}

func (c *Client) UpdateDataSource(s *DataSource) error {
	path := fmt.Sprintf("/api/datasources/%d", s.Id)
	data, err := json.Marshal(s)
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

func (c *Client) DataSource(id int64) (*DataSource, error) {
	path := fmt.Sprintf("/api/datasources/%d", id)
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

	result := &DataSource{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func (c *Client) DeleteDataSource(id int64) error {
	path := fmt.Sprintf("/api/datasources/%d", id)
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
