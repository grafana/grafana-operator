package client

import (
	"fmt"
	"net/url"
)

func ParseAdminURL(adminURL string) (*url.URL, error) {
	gURL, err := url.Parse(adminURL)
	if err != nil {
		return nil, fmt.Errorf("parsing url for client: %w", err)
	}

	if gURL.Host == "" {
		return nil, fmt.Errorf("invalid Grafana adminURL, url must contain protocol and host")
	}

	gURL = gURL.JoinPath("/api")

	return gURL, nil
}
