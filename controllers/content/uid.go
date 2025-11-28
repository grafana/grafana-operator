package content

import "github.com/grafana/grafana-operator/v5/api/v1beta1"

func IsUpdatedUID(cr v1beta1.GrafanaContentResource, uid string) bool {
	status := cr.GrafanaContentStatus()
	// This indicates an implementation error
	if status == nil {
		return false
	}

	// Resource has just been created, status is not yet updated
	if status.UID == "" {
		return false
	}

	uid = GetGrafanaUID(cr, uid)

	return status.UID != uid
}

// GetGrafanaUID selects a UID to be used for Grafana API requests (preference: spec.CustomUID -> contentUID -> metadata.uid)
func GetGrafanaUID(cr v1beta1.GrafanaContentResource, contentUID string) string {
	if spec := cr.GrafanaContentSpec(); spec != nil {
		if spec.CustomUID != "" {
			return spec.CustomUID
		}
	}

	if contentUID != "" {
		return contentUID
	}

	return string(cr.GetUID())
}
