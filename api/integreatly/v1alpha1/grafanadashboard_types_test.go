package v1alpha1

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Encoded via cat | gzip | base64
const encodedCompressedDashboard = `
H4sIAAAAAAAAA3WQMU/DQAyF9/6KU2aQYAAkVliZqFgQqpzGSaxczief2wqq/nd8l5DAwOb3+dnP
8nnjXEVN9ejCwfurrJTUo4Hqlcbo0T1D6msGaVwrPLonDi117gViNdmhS+Z+/ygq6ec03IAMs4FG
/OJQaC18SihTAxtSqItd5YCF9dSgJaiwz1tb8GlqdAKx3zJ7pWiN2wIjBPS/0nOUqbPVpvK5OTTw
6fq+L5nZwzOrTF+WsUj7wQ5bhjPbcVTisAYYF2wFU7+joChHmNPXVWg/A6XQras8Jf3rghBY4Wf3
v7Y5K997l6afpX2PI7yhJBvOf3go+LiAm6I9hWG+7LL5BgwYIaHkAQAA
`

const decodedDashboard = `
{
  "id": null,
  "title": "Simple Dashboard from Config Map",
  "tags": [],
  "style": "dark",
  "timezone": "browser",
  "editable": true,
  "hideControls": false,
  "graphTooltip": 1,
  "panels": [],
  "time": {
    "from": "now-6h",
    "to": "now"
  },
  "timepicker": {
    "time_options": [],
    "refresh_intervals": []
  },
  "templating": {
    "list": []
  },
  "annotations": {
    "list": []
  },
  "refresh": "5s",
  "schemaVersion": 17,
  "version": 0,
  "links": []
}
`

func TestDecompress(t *testing.T) {
	var expected map[string]interface{}
	var actual map[string]interface{}
	decoded, err := ioutil.ReadAll(base64.NewDecoder(base64.StdEncoding, strings.NewReader(encodedCompressedDashboard)))
	if err != nil {
		t.Log("Failed to decode", err)
		t.Fail()
	}
	decompressed, err := Gunzip(decoded)
	if err != nil {
		t.Log("Failed to decompress", err)
		t.Fail()
	}
	err = json.Unmarshal([]byte(decompressed), &actual)
	if err != nil {
		t.Log("Failed to parse JSON from decoded", err)
		t.Fail()
	}
	err = json.Unmarshal([]byte(decodedDashboard), &expected)
	if err != nil {
		t.Log("Failed to parse JSON from ground truth", err)
		t.Fail()
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Log("Decoded JSONs were not the same")
		t.Fail()
	}
}

func TestGrafanaDashboardStatus_getContentCache(t *testing.T) {
	timestamp := metav1.Time{Time: time.Now().Add(-1 * time.Hour)}
	infinite := 0 * time.Second
	dashboardJSON := `{"dummyField": "dummyData"}`

	cachedDashboard, err := Gzip(dashboardJSON)
	assert.Nil(t, err)

	url := "http://127.0.0.1:8080/1.json"

	// Correctly populated cache
	status := GrafanaDashboardStatus{
		ContentCache:     cachedDashboard,
		ContentTimestamp: timestamp,
		ContentUrl:       url,
	}

	// Corrupted cache
	statusCorrupted := GrafanaDashboardStatus{
		ContentCache:     []byte("abc"),
		ContentTimestamp: timestamp,
		ContentUrl:       url,
	}

	tests := []struct {
		name     string
		status   GrafanaDashboardStatus
		url      string
		duration time.Duration
		want     string
	}{
		{
			name:     "no cache: fields are not populated",
			url:      status.ContentUrl,
			duration: infinite,
			status:   GrafanaDashboardStatus{},
			want:     "",
		},
		{
			name:     "no cache: url is different",
			url:      "http://another-url/2.json",
			duration: infinite,
			status:   status,
			want:     "",
		},
		{
			name:     "no cache: expired",
			url:      status.ContentUrl,
			duration: 1 * time.Minute,
			status:   status,
			want:     "",
		},
		{
			name:     "no cache: corrupted gzip",
			url:      statusCorrupted.ContentUrl,
			duration: infinite,
			status:   statusCorrupted,
			want:     "",
		},
		{
			name:     "valid cache: not expired yet",
			url:      status.ContentUrl,
			duration: 24 * time.Hour,
			status:   status,
			want:     dashboardJSON,
		},
		{
			name:     "valid cache: not expired yet (infinite)",
			url:      status.ContentUrl,
			duration: infinite,
			status:   status,
			want:     dashboardJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.status.getContentCache(tt.url, tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}
