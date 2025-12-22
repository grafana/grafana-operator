package v1beta1

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("GrafanaCommonSpec#AllowCrossNamespaceImport Validation rule tests", func() {
	t := GinkgoT()

	undefinedCrossImportFolder := &GrafanaFolder{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaFolder",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
		},
		Spec: GrafanaFolderSpec{
			GrafanaCommonSpec: GrafanaCommonSpec{
				InstanceSelector: &metav1.LabelSelector{},
			},
		},
	}

	ctx := context.Background()
	Context("Enabling allowCrossNamespaceImport", func() {
		It("Allows setting allowCrossNamespaceImport after creation", func() {
			copyOfundefinedCrossImportFolder := undefinedCrossImportFolder.DeepCopy()
			copyOfundefinedCrossImportFolder.Name = "disabled-from-undefined"
			By("Creating a Folder without allowCrossNamespaceImport")
			err := cl.Create(ctx, copyOfundefinedCrossImportFolder)
			require.NoError(t, err)

			By("Setting allowCrossNamespaceImport false")
			copyOfundefinedCrossImportFolder.Spec.AllowCrossNamespaceImport = false
			err = cl.Update(ctx, copyOfundefinedCrossImportFolder)
			require.NoError(t, err)
		})

		It("Allows enabling allowCrossNamespaceImport from undefined", func() {
			secondUndfinedCrossImportFolder := undefinedCrossImportFolder.DeepCopy()
			secondUndfinedCrossImportFolder.Name = "enabled-from-undefined"

			By("Creating a Folder with false allowCrossNamespaceImport")
			err := cl.Create(ctx, secondUndfinedCrossImportFolder)
			require.NoError(t, err)

			By("Setting allowCrossNamespaceImport true")
			secondUndfinedCrossImportFolder.Spec.AllowCrossNamespaceImport = true
			err = cl.Update(ctx, secondUndfinedCrossImportFolder)
			require.NoError(t, err)
		})

		It("Allows enabling allowCrossNamespaceImport when false", func() {
			explicitNoCrossImportFolder := undefinedCrossImportFolder.DeepCopy()
			explicitNoCrossImportFolder.Name = "enabled-from-false"
			explicitNoCrossImportFolder.Spec.AllowCrossNamespaceImport = false
			By("Creating a Folder with allowCrossNamespaceImport false")
			err := cl.Create(ctx, explicitNoCrossImportFolder)
			require.NoError(t, err)

			By("Setting allowCrossNamespaceImport true")
			explicitNoCrossImportFolder.Spec.AllowCrossNamespaceImport = true
			err = cl.Update(ctx, explicitNoCrossImportFolder)
			require.NoError(t, err)
		})
	})

	Context("Disabling allowCrossNamespaceImport is blocked", func() {
		It("Blocks disabling allowCrossNamespaceImport after creation with false", func() {
			crossImportFolder := undefinedCrossImportFolder.DeepCopy()
			crossImportFolder.Name = "disabled-from-true"
			crossImportFolder.Spec.AllowCrossNamespaceImport = true
			By("Creating a Folder with allowCrossNamespaceImport")
			err := cl.Create(ctx, crossImportFolder)
			require.NoError(t, err)

			By("Setting allowCrossNamespaceImport false")
			crossImportFolder.Spec.AllowCrossNamespaceImport = false
			err = cl.Update(ctx, crossImportFolder)
			require.Error(t, err)
		})

		It("Blocks disabling allowCrossNamespaceImport after creation with undefined", func() {
			secondCrossImportFolder := undefinedCrossImportFolder.DeepCopy()
			secondCrossImportFolder.Name = "unset-from-true"
			secondCrossImportFolder.Spec.AllowCrossNamespaceImport = true
			By("Creating a Folder with allowCrossNamespaceImport")
			err := cl.Create(ctx, secondCrossImportFolder)
			require.NoError(t, err)

			By("Setting allowCrossNamespaceImport false")
			unsetCrossImportFolder := undefinedCrossImportFolder.DeepCopy()
			unsetCrossImportFolder.Name = "unset-from-true" // Needs the same name as above
			err = cl.Update(ctx, unsetCrossImportFolder)
			require.Error(t, err)
		})
	})
})

func TestGetPluginConfigMapKey(t *testing.T) {
	longName := strings.Repeat("a", 100)
	longNameHash := "2816597888e4a0d3a36b82b83316ab32680eb8f00f8cd3b904d681246d285a0e"

	tests := []struct {
		name   string
		prefix string
		meta   metav1.ObjectMeta
		want   string
	}{
		{
			name:   "short-name",
			prefix: "dashboard",
			meta: metav1.ObjectMeta{
				Name:      "short-name",
				Namespace: "default",
			},
			want: "dashboard_default_short-name",
		},
		{
			name:   "name over 63 characters",
			prefix: "datasource",
			meta: metav1.ObjectMeta{
				Name:      longName,
				Namespace: "default",
			},
			want: fmt.Sprintf("datasource_default_%s-%s", longName[:63], longNameHash),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := GrafanaDashboard{
				ObjectMeta: tt.meta,
			}

			got := GetPluginConfigMapKey(tt.prefix, &cr.ObjectMeta)
			assert.Equal(t, tt.want, got)
		})
	}
}
