package grafananotificationchannel

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func NewClient() *http.Client {
	return &http.Client{
		Timeout: time.Duration(30) * time.Second,
	}
}

func TestGrafanaClient_CreateNotificationChannel_Positive(t *testing.T) {
	r := require.New(t)
	type grafanaClient struct {
		url      string
		user     string
		password string
		client   *http.Client
	}
	type args struct {
		channel []byte
	}

	c := `{
      "uid": "PD-alert-notification",
      "name": "PD alert notification",
      "type":  "pagerduty",
      "isDefault": true,
      "sendReminder": true,
      "frequency": "15m",
	  "disableResolveMessage": true,	
      "settings": {
        "integrationKey": "put key here",
        "autoResolve": true,
        "uploadImage": true
     }
    }`
	id := uint(1)
	uid := "PD-alert-notification"
	name := "PD alert notification"
	nType := "pagerduty"
	isDefault := true
	sendReminder := true
	disableResolveMessage := true
	created := "2020-05-25 00:00"
	updated := "2020-05-25 00:00"
	message := "success"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/alert-notifications/" {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{
								"id": 1,
								"uid": "PD-alert-notification",
      							"name": "PD alert notification",
      							"type":  "pagerduty",
      							"isDefault": true,
      							"sendReminder": true,
      							"frequency": "15m",
								"disableResolveMessage": true, 
      							"settings": {
											"integrationKey": "put key here",
        									"autoResolve": true,
        									"uploadImage": true
     										},
								"created": "2020-05-25 00:00",
								"updated": "2020-05-25 00:00",
								"message" : "success"
    							}`))
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	tests := []struct {
		name          string
		grafanaClient grafanaClient
		args          args
		want          GrafanaResponse
	}{
		{name: "Create channel test",
			grafanaClient: grafanaClient{
				url:      ts.URL,
				user:     "testUser",
				password: "testPassword",
				client:   NewClient(),
			},
			args: args{channel: []byte(c)},
			want: GrafanaResponse{
				ID:                    &id,
				UID:                   &uid,
				Name:                  &name,
				Type:                  &nType,
				IsDefault:             &isDefault,
				SendReminder:          &sendReminder,
				DisableResolveMessage: &disableResolveMessage,
				Created:               &created,
				Updated:               &updated,
				Message:               &message,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GrafanaClientImpl{
				url:      tt.grafanaClient.url,
				user:     tt.grafanaClient.user,
				password: tt.grafanaClient.password,
				client:   tt.grafanaClient.client,
			}
			got, err := gc.CreateNotificationChannel(tt.args.channel)
			r.NoError(err, "CreateNotificationChannel() error = %v", err)
			r.Equal(tt.want, got)
		})
	}
}

func TestGrafanaClient_CreateNotificationChannel_Negative(t *testing.T) {
	r := require.New(t)
	type grafanaClient struct {
		url      string
		user     string
		password string
		client   *http.Client
	}
	type args struct {
		channel []byte
	}
	c := `{
      "uid": "PD-alert-notification",
      "name": "PD alert notification",
      "type":  "pagerduty",
      "isDefault": true,
      "sendReminder": true,
      "frequency": "15m",
	  "disableResolveMessage": true,	
      "settings": {
        "integrationKey": "put key here",
        "autoResolve": true,
        "uploadImage": true
     }
    }`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/alert-notifications/" {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message" : "error creating notificationChannel, expected status 200 but got 500"}`))
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	tests := []struct {
		name          string
		grafanaClient grafanaClient
		args          args
		want          GrafanaResponse
		wantErr       bool
	}{
		{name: "Create channel negative test",
			grafanaClient: grafanaClient{
				url:      ts.URL,
				user:     "testUser",
				password: "testPassword",
				client:   NewClient(),
			},
			args: args{channel: []byte(c)},
			want: GrafanaResponse{
				Message: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GrafanaClientImpl{
				url:      tt.grafanaClient.url,
				user:     tt.grafanaClient.user,
				password: tt.grafanaClient.password,
				client:   tt.grafanaClient.client,
			}
			_, err := gc.CreateNotificationChannel(tt.args.channel)
			r.Error(err, "CreateNotificationChannel() error = %v", err)
		})
	}
}

func TestGrafanaClient_UpdateNotificationChannel(t *testing.T) {
	r := require.New(t)
	type grafanaClient struct {
		url      string
		user     string
		password string
		client   *http.Client
	}
	type args struct {
		channel []byte
	}

	c := `{
      "uid": "PD-alert-notification",
      "name": "PD alert notification",
      "type":  "pagerduty",
      "isDefault": true,
      "sendReminder": true,
      "frequency": "15m",
	  "disableResolveMessage": true,	
      "settings": {
        "integrationKey": "put key here",
        "autoResolve": true,
        "uploadImage": true
     }
    }`
	id := uint(1)
	uid := "PD-alert-notification"
	name := "PD alert notification"
	nType := "pagerduty"
	isDefault := true
	sendReminder := true
	disableResolveMessage := true
	created := "2020-05-25 00:00"
	updated := "2020-05-25 00:00"
	message := "success"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/alert-notifications/uid/"+uid {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{
								"id": 1,
								"uid": "PD-alert-notification",
      							"name": "PD alert notification",
      							"type":  "pagerduty",
      							"isDefault": true,
      							"sendReminder": true,
      							"frequency": "15m",
								"disableResolveMessage": true, 
      							"settings": {
											"integrationKey": "put key here",
        									"autoResolve": true,
        									"uploadImage": true
     										},
								"created": "2020-05-25 00:00",
								"updated": "2020-05-25 00:00",
								"message" : "success"
    							}`))
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	tests := []struct {
		name          string
		grafanaClient grafanaClient
		args          args
		want          GrafanaResponse
	}{
		{name: "Update channel test",
			grafanaClient: grafanaClient{
				url:      ts.URL,
				user:     "testUser",
				password: "testPassword",
				client:   NewClient(),
			},
			args: args{channel: []byte(c)},
			want: GrafanaResponse{
				ID:                    &id,
				UID:                   &uid,
				Name:                  &name,
				Type:                  &nType,
				IsDefault:             &isDefault,
				SendReminder:          &sendReminder,
				DisableResolveMessage: &disableResolveMessage,
				Created:               &created,
				Updated:               &updated,
				Message:               &message,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GrafanaClientImpl{
				url:      tt.grafanaClient.url,
				user:     tt.grafanaClient.user,
				password: tt.grafanaClient.password,
				client:   tt.grafanaClient.client,
			}
			got, err := gc.UpdateNotificationChannel(tt.args.channel, uid)
			r.NoError(err, "UpdateNotificationChannel() error = %v", err)
			r.Equal(tt.want, got)
		})
	}
}

func TestGrafanaClient_DeleteNotificationChannel(t *testing.T) {
	r := require.New(t)
	type grafanaClient struct {
		url      string
		user     string
		password string
		client   *http.Client
	}
	uid := "PD-alert-notification"
	message := "Notification deleted"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/alert-notifications/uid/"+uid {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{
								"message" : "Notification deleted"
    							}`))
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	tests := []struct {
		name          string
		grafanaClient grafanaClient
		want          GrafanaResponse
	}{
		{name: "Delete channel test",
			grafanaClient: grafanaClient{
				url:      ts.URL,
				user:     "testUser",
				password: "testPassword",
				client:   NewClient(),
			},
			want: GrafanaResponse{
				Message: &message,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GrafanaClientImpl{
				url:      tt.grafanaClient.url,
				user:     tt.grafanaClient.user,
				password: tt.grafanaClient.password,
				client:   tt.grafanaClient.client,
			}
			got, err := gc.DeleteNotificationChannelByUID(uid)
			r.NoError(err, "DeleteNotificationChannelByUID() error = %v", err)
			r.Equal(tt.want.Message, got.Message)
		})
	}
}

func TestGrafanaClient_GetNotificationChannel(t *testing.T) {
	r := require.New(t)
	type grafanaClient struct {
		url      string
		user     string
		password string
		client   *http.Client
	}
	id := uint(1)
	uid := "PD-alert-notification"
	name := "PD alert notification"
	nType := "pagerduty"
	isDefault := true
	sendReminder := true
	disableResolveMessage := true
	created := "2020-05-25 00:00"
	updated := "2020-05-25 00:00"
	message := "success"

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/alert-notifications/uid/"+uid {
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{
								"id": 1,
								"uid": "PD-alert-notification",
      							"name": "PD alert notification",
      							"type":  "pagerduty",
      							"isDefault": true,
      							"sendReminder": true,
      							"frequency": "15m",
								"disableResolveMessage": true, 
      							"settings": {
											"integrationKey": "put key here",
        									"autoResolve": true,
        									"uploadImage": true
     										},
								"created": "2020-05-25 00:00",
								"updated": "2020-05-25 00:00",
								"message" : "success"
    							}`))
		}
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	tests := []struct {
		name          string
		grafanaClient grafanaClient
		want          GrafanaResponse
	}{
		{name: "Delete channel test",
			grafanaClient: grafanaClient{
				url:      ts.URL,
				user:     "testUser",
				password: "testPassword",
				client:   NewClient(),
			},
			want: GrafanaResponse{
				ID:                    &id,
				UID:                   &uid,
				Name:                  &name,
				Type:                  &nType,
				IsDefault:             &isDefault,
				SendReminder:          &sendReminder,
				DisableResolveMessage: &disableResolveMessage,
				Created:               &created,
				Updated:               &updated,
				Message:               &message,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gc := &GrafanaClientImpl{
				url:      tt.grafanaClient.url,
				user:     tt.grafanaClient.user,
				password: tt.grafanaClient.password,
				client:   tt.grafanaClient.client,
			}
			got, err := gc.GetNotificationChannel(uid)
			r.NoError(err, "GetNotificationChannel() error = %v", err)
			r.Equal(tt.want, got)
		})
	}
}
