package content

import "github.com/grafana/grafana-operator/v5/api/v1beta1"

func HasChanged(cr v1beta1.GrafanaContentResource, hash string) bool {
	return !Unchanged(cr, hash)
}

// Unchanged checks if the stored content hash on the status matches the input
func Unchanged(cr v1beta1.GrafanaContentResource, hash string) bool {
	status := cr.GrafanaContentStatus()

	// This indicates an implementation error
	if status == nil {
		return true
	}

	return status.Hash == hash
}
