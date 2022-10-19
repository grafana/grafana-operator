package grafanadashboardfolder

import (
	"github.com/grafana-operator/grafana-operator/v4/api/integreatly/v1alpha1"
	"github.com/stretchr/testify/require"
	"testing"
)

const expectedFolderPermissionRequestBody = `{ "items": [ {"role": "Viewer", "permission": 1},{"role": "Editor", "permission": 2} ]}`

var folderPermissions = []*v1alpha1.GrafanaPermissionItem{
	{
		PermissionTargetType: "role",
		PermissionTarget:     "Viewer",
		PermissionLevel:      1,
	},
	{
		PermissionTargetType: "role",
		PermissionTarget:     "Editor",
		PermissionLevel:      2,
	},
}

func TestBuildFolderPermissionRequestBody(t *testing.T) {
	result := buildFolderPermissionRequestBody(folderPermissions)
	require.Equal(t, expectedFolderPermissionRequestBody, result)
}

func TestBuildFolderUidFromName(t *testing.T) {
	result := buildFolderUidFromName("Test Folder")
	require.Equal(t, "ced3f5903fc61238671a36457c61fc81", result)
}
