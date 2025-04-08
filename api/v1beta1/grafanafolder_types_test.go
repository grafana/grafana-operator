package v1beta1

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func TestGrafanaFolder_GetUID(t *testing.T) {
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
			got := tt.cr.CustomUIDOrUID()
			assert.Equal(t, tt.want, got)
		})
	}
}

func newFolder(name string, uid string) *GrafanaFolder {
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
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			folder := newFolder("missing-uid", "")
			By("Create new Folder without uid")
			Expect(k8sClient.Create(ctx, folder)).To(Succeed())

			By("Adding a uid")
			folder.Spec.CustomUID = "new-folder-uid"
			Expect(k8sClient.Update(ctx, folder)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			folder := newFolder("existing-uid", "existing-uid")
			By("Creating Folder with existing UID")
			Expect(k8sClient.Create(ctx, folder)).To(Succeed())

			By("And setting UID to ''")
			folder.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, folder)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			folder := newFolder("removing-uid", "existing-uid")
			By("Create new Folder with existing UID")
			Expect(k8sClient.Create(ctx, folder)).To(Succeed())

			By("Changing the existing UID")
			folder.Spec.CustomUID = "new-folder-uid"
			Expect(k8sClient.Update(ctx, folder)).To(HaveOccurred())
		})
	})
})
