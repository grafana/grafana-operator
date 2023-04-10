package util

import (
	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	grapi "github.com/grafana/grafana-api-golang-client"
)

func GetOrCreateFolder(client *grapi.Client, dashboard *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
	if dashboard.Spec.FolderTitle == "" {
		return nil, nil
	}

	folder, err := GetFolder(client, dashboard)
	if err != nil {
		return nil, err
	}
	if folder != nil {
		return folder, nil
	}

	// Folder wasn't found, let's create it
	resp, err := client.NewFolder(dashboard.Spec.FolderTitle)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func GetFolder(client *grapi.Client, dashboard *v1beta1.GrafanaDashboard) (*grapi.Folder, error) {
	folders, err := client.Folders()
	if err != nil {
		return nil, err
	}

	for _, folder := range folders {
		if folder.Title == dashboard.Spec.FolderTitle {
			return &folder, nil
		}
		continue
	}
	return nil, nil
}
