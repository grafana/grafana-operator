package cache

import (
	"time"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
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
