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

	assert.Equal(t, "namespace", ns)
	assert.Equal(t, "name", n)
	assert.Equal(t, "identifier", i)
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
	list := NamespacedResourceList{
		NamespacedResource("default/folder0/aaaa"),
		NamespacedResource("default/folder1/bbbb"),
		NamespacedResource("default/folder2/cccc"),
	}

	tests := []struct {
		name       string
		rNamespace string
		rName      string
		want       int
	}{
		{
			name:       "Not found",
			rNamespace: "default",
			rName:      "not-found",
			want:       -1,
		},
		{
			name:       "Found at 0",
			rNamespace: "default",
			rName:      "folder0",
			want:       0,
		},
		{
			name:       "Found at 2",
			rNamespace: "default",
			rName:      "folder2",
			want:       2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := list.IndexOf(tt.rNamespace, tt.rName)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveEntries(t *testing.T) {
	r1 := NamespacedResource("1/1/1")
	r2 := NamespacedResource("1/1/2")
	r3 := NamespacedResource("3/3/3")
	r4 := NamespacedResource("3/3/4")

	tests := []struct {
		name     string
		list     NamespacedResourceList
		toRemove NamespacedResourceList
		want     NamespacedResourceList
	}{
		{
			name:     "Remove 'missing' entry from list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r4},
			want:     NamespacedResourceList{r1, r2, r3},
		},
		{
			name:     "Remove first entry from the list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r1},
			want:     NamespacedResourceList{r2, r3},
		},
		{
			name:     "Remove middle entry from the list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r2},
			want:     NamespacedResourceList{r1, r3},
		},
		{
			name:     "Remove last entry from the list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r3},
			want:     NamespacedResourceList{r1, r2},
		},
		{
			name:     "Remove multiple entries from the list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r1, r2},
			want:     NamespacedResourceList{r3},
		},
		{
			name:     "Remove all entries from the list",
			list:     NamespacedResourceList{r1, r2, r3},
			toRemove: NamespacedResourceList{r1, r2, r3},
			want:     NamespacedResourceList{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.list.RemoveEntries(&tt.toRemove)

			assert.Equal(t, tt.want, got)
			for _, r := range tt.toRemove {
				assert.NotContainsf(t, got, r, "Resources should have removed from the source list")
			}
		})
	}
}
