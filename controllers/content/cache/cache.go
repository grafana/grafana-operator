package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetContentCache(cr v1beta1.GrafanaContentResource) []byte {
	spec := cr.GrafanaContentSpec()
	if spec == nil {
		return nil
	}

	status := cr.GrafanaContentStatus()
	if status == nil {
		return nil
	}

	return getContentCache(status, spec.URL, spec.ContentCacheDuration.Duration)
}

// getContentCache returns content cache when the following conditions are met: url is the same, data is not expired, gzipped data is not corrupted
func getContentCache(in *v1beta1.GrafanaContentStatus, url string, cacheDuration time.Duration) []byte {
	if in.ContentURL != url {
		return []byte{}
	}

	notExpired := cacheDuration <= 0 || in.ContentTimestamp.Add(cacheDuration).After(time.Now())
	if !notExpired {
		return []byte{}
	}

	cache, err := Gunzip(in.ContentCache)
	if err != nil {
		return []byte{}
	}

	return cache
}

func SetContentCache(cr v1beta1.GrafanaContentResource, data map[string]any) error {
	spec := cr.GrafanaContentSpec()
	status := cr.GrafanaContentStatus()

	return setContentCache(status, spec.URL, data, spec.ContentCacheDuration.Duration)
}

func setContentCache(in *v1beta1.GrafanaContentStatus, url string, data map[string]any, cacheDuration time.Duration) error {
	notExpired := cacheDuration <= 0 || in.ContentTimestamp.Add(cacheDuration).After(time.Now())
	if notExpired && in.ContentURL == url {
		return nil
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshaling content: %w", err)
	}

	gz, err := Gzip(encoded)
	if err != nil {
		return fmt.Errorf("compressing content: %w", err)
	}

	in.ContentCache = gz
	in.ContentTimestamp = metav1.Time{Time: time.Now()}
	in.ContentURL = url

	return nil
}
