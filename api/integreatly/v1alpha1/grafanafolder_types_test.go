/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/stretchr/testify/require"
	"testing"
)

var folderPermissions = []GrafanaPermissionItem{
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

func TestGetPermissions(t *testing.T) {
	folder := new(GrafanaFolder)
	folder.Spec.FolderName = "TEST"
	folder.Spec.FolderPermissions = folderPermissions

	result := folder.GetPermissions()
	require.NotNil(t, result)
	require.Equal(t, 2, len(result))
	require.Equal(t, "Viewer", result[0].PermissionTarget)
	require.Equal(t, "Editor", result[1].PermissionTarget)
}

func TestHash(t *testing.T) {
	folder := new(GrafanaFolder)
	folder.Spec.FolderName = "TEST"
	folder.Spec.FolderPermissions = folderPermissions

	result := folder.Hash()
	require.Equal(t, "c44659960c5741f3ee2f949e3df5a41c04a03acac15dbeb05f6c2a7423232b6c", result)
}
