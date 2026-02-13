package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGrafanaStatusListFolder(t *testing.T) {
	t.Run("&Folder{} maps to NamespacedResource list", func(t *testing.T) {
		g := &Grafana{}
		arg := &GrafanaFolder{}
		_, _, err := g.Status.StatusList(arg)
		assert.NoError(t, err, "Folder does not have a case in Grafana.Status.StatusList")
	})
}

func TestGrafanaFolder_GetTitle(t *testing.T) {
	tests := []struct {
		name string
		cr   GrafanaFolder
		want string
	}{
		{
			name: "No custom title",
			cr: GrafanaFolder{
				ObjectMeta: metav1.ObjectMeta{Name: "cr-name"},
			},
			want: "cr-name",
		},
		{
			name: "Custom title",
			cr: GrafanaFolder{
				ObjectMeta: metav1.ObjectMeta{Name: "cr-name"},
				Spec: GrafanaFolderSpec{
					Title: "custom-title",
				},
			},
			want: "custom-title",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cr.GetTitle()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGrafanaFolder_GetGrafanaUID(t *testing.T) {
	tests := []struct {
		name string
		cr   GrafanaFolder
		want string
	}{
		{
			name: "No custom UID",
			cr: GrafanaFolder{
				ObjectMeta: metav1.ObjectMeta{UID: "92fd2e0a-ad63-4fcf-9890-68a527cbd674"},
			},
			want: "92fd2e0a-ad63-4fcf-9890-68a527cbd674",
		},
		{
			name: "Custom UID",
			cr: GrafanaFolder{
				ObjectMeta: metav1.ObjectMeta{UID: "92fd2e0a-ad63-4fcf-9890-68a527cbd674"},
				Spec: GrafanaFolderSpec{
					CustomUID: "custom-uid",
				},
			},
			want: "custom-uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cr.GetGrafanaUID()
			assert.Equal(t, tt.want, got)
		})
	}
}

func newFolder(name, uid string) *GrafanaFolder {
	return &GrafanaFolder{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaFolder",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaFolderSpec{
			CustomUID: uid,
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"test": "folder",
					},
				},
			},
		},
	}
}

var _ = Describe("Folder type", func() {
	Context("Ensure Folder spec.uid is immutable", func() {
		t := GinkgoT()

		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			folder := newFolder("missing-uid", "")

			By("Create new Folder without uid")

			err := cl.Create(ctx, folder)
			require.NoError(t, err)

			By("Adding a uid")

			folder.Spec.CustomUID = "new-folder-uid"
			err = cl.Update(ctx, folder)
			require.Error(t, err)
		})

		It("Should block removing uid field when set", func() {
			folder := newFolder("existing-uid", "existing-uid")

			By("Creating Folder with existing UID")

			err := cl.Create(ctx, folder)
			require.NoError(t, err)

			By("And setting UID to ''")

			folder.Spec.CustomUID = ""
			err = cl.Update(ctx, folder)
			require.Error(t, err)
		})

		It("Should block changing value of uid", func() {
			folder := newFolder("removing-uid", "existing-uid")

			By("Create new Folder with existing UID")

			err := cl.Create(ctx, folder)
			require.NoError(t, err)

			By("Changing the existing UID")

			folder.Spec.CustomUID = "new-folder-uid"
			err = cl.Update(ctx, folder)
			require.Error(t, err)
		})
	})
})
