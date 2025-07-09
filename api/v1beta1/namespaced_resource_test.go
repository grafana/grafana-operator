package v1beta1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func strP(s string) *string {
	return &s
}

func mockNamespacedResourceList() NamespacedResourceList {
	return NamespacedResourceList{
		NamespacedResource("default/folder0/aaaa"),
		NamespacedResource("default/folder1/bbbb"),
		NamespacedResource("default/folder2/cccc"),
		NamespacedResource("default/folder3/dddd"),
	}
}

func TestSplit(t *testing.T) {
	r := NamespacedResource("namespace/name/identifier")
	ns, n, i := r.Split()

	assert.Equal(t, ns, "namespace")
	assert.Equal(t, n, "name")
	assert.Equal(t, i, "identifier")
}

func TestFind(t *testing.T) {
	in := mockNamespacedResourceList()

	tests := []struct {
		testName   string
		rNamespace string
		rName      string
		found      bool
		wantIdent  *string
	}{
		{
			testName:   "Missing from list",
			rNamespace: "default",
			rName:      "not-found",
			found:      false,
			wantIdent:  nil,
		},
		{
			testName:   "Present in list",
			rNamespace: "default",
			rName:      "folder2",
			found:      true,
			wantIdent:  strP("cccc"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			found, gotIdent := in.Find(tt.rNamespace, tt.rName)

			assert.Equal(t, tt.found, found)
			assert.Equal(t, tt.wantIdent, gotIdent)
		})
	}
}

func TestIndexOf(t *testing.T) {
	in := mockNamespacedResourceList()

	tests := []struct {
		testName   string
		rNamespace string
		rName      string
		wantIdx    int
	}{
		{
			testName:   "Missing from list",
			rNamespace: "default",
			rName:      "not-found",
			wantIdx:    -1,
		},
		{
			testName:   "Present in list",
			rNamespace: "default",
			rName:      "folder2",
			wantIdx:    2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			gotIdx := in.IndexOf(tt.rNamespace, tt.rName)

			assert.Equal(t, tt.wantIdx, gotIdx)
		})
	}
}

func TestRemoveEntries(t *testing.T) {
	in := mockNamespacedResourceList()

	tests := []struct {
		name     string
		toRemove NamespacedResourceList
		wantLen  int
	}{
		{
			name:     "Remove 'missing' entry from list",
			toRemove: NamespacedResourceList{},
			wantLen:  len(in),
		},
		{
			name:     "Remove first entry from the list",
			toRemove: NamespacedResourceList{"default/folder0/aaaa"},
			wantLen:  len(in) - 1,
		},
		{
			name:     "Remove middle entry from the list",
			toRemove: NamespacedResourceList{"default/folder2/cccc"},
			wantLen:  len(in) - 1,
		},
		{
			name:     "Remove last entry from the list",
			toRemove: NamespacedResourceList{"default/folder3/dddd"},
			wantLen:  len(in) - 1,
		},
		{
			name:     "Remove multiple entries from the list",
			toRemove: NamespacedResourceList{"default/folder1/bbbb", "default/folder2/cccc", "default/folder3/dddd"},
			wantLen:  len(in) - 3,
		},
		{
			name:     "Remove all entries from the list",
			toRemove: NamespacedResourceList{"default/folder0/aaaa", "default/folder1/bbbb", "default/folder2/cccc", "default/folder3/dddd"},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		list := mockNamespacedResourceList()

		t.Run(tt.name, func(t *testing.T) {
			gotList := list.RemoveEntries(&tt.toRemove)

			assert.Equal(t, tt.wantLen, len(gotList))
			for _, r := range tt.toRemove {
				assert.NotContainsf(t, gotList, r, "Resources should have removed from the source list")
			}
		})
	}
}
