package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAdminURL(t *testing.T) {
	tests := []struct {
		name      string
		adminURL  string
		wantHost  string
		wantPath  string
		wantError bool
	}{
		{
			name:      "No Path",
			adminURL:  "https://grafana.example.com",
			wantHost:  "grafana.example.com",
			wantPath:  "api",
			wantError: false,
		},
		{
			name:      "Root as Path",
			adminURL:  "https://grafana.example.com/",
			wantHost:  "grafana.example.com",
			wantPath:  "/api",
			wantError: false,
		},
		{
			name:      "Custom Port",
			adminURL:  "https://grafana.example.com:3000/",
			wantHost:  "grafana.example.com:3000",
			wantPath:  "/api",
			wantError: false,
		},
		{
			name:      "Invalid URL",
			adminURL:  "%",
			wantError: true,
		},
		{
			name:      "No Path and no Scheme",
			adminURL:  "grafana.example.com",
			wantError: true,
		},
		{
			name:      "No Scheme",
			adminURL:  "grafana.example.com/path",
			wantError: true,
		},
		{
			name:      "Custom Path",
			adminURL:  "https://grafana.example.com/instances/1",
			wantHost:  "grafana.example.com",
			wantPath:  "/instances/1/api",
			wantError: false,
		},
		{
			name:      "Relative Custom Path",
			adminURL:  "https://grafana.example.com/../test",
			wantHost:  "grafana.example.com",
			wantPath:  "/test/api",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseAdminURL(tt.adminURL)
			if tt.wantError {
				assert.Error(t, err, "This should be an invalid url input")
			} else {
				require.NoError(t, err, "This should be a valid url")
				assert.Equal(t, tt.wantPath, got.Path, "Path does not match")
				assert.Equal(t, tt.wantHost, got.Host, "Host does not match")
				assert.Contains(t, got.Path, "api", "/api is not appended to path correctly")
			}
		})
	}
}
