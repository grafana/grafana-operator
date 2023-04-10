package util

import (
	"strings"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	grapi "github.com/grafana/grafana-api-golang-client"
)

func ExistingDashboardVersionMatches(client *grapi.Client, instanceStatus v1beta1.GrafanaDashboardInstanceStatus) (bool, error) {
	existing, err := client.DashboardByUID(instanceStatus.UID)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, err
	}

	if float64(instanceStatus.Version) == existing.Model["version"].(float64) {
		return true, nil
	}

	return false, nil
}
