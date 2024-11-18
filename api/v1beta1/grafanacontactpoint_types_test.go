package v1beta1

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	// apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newContactPoint(name string, uid string) *GrafanaContactPoint {
	settings := new(apiextensionsv1.JSON)
	json.Unmarshal([]byte("{}"), settings) //nolint:errcheck

	return &GrafanaContactPoint{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       "GrafanaContactPoint",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: GrafanaContactPointSpec{
			CustomUID: uid,
			InstanceSelector: &v1.LabelSelector{
				MatchLabels: map[string]string{
					"test": "datasource",
				},
			},
			Settings: settings,
		},
	}
}

var _ = Describe("ContactPoint type", func() {
	Context("Ensure ContactPoint spec.uid is immutable", func() {
		ctx := context.Background()

		It("Should block adding uid field when missing", func() {
			contactpoint := newContactPoint("missing-uid", "")
			By("Create new ContactPoint without uid")
			Expect(k8sClient.Create(ctx, contactpoint)).To(Succeed())

			By("Adding a uid")
			contactpoint.Spec.CustomUID = "new-contactpoint-uid"
			Expect(k8sClient.Update(ctx, contactpoint)).To(HaveOccurred())
		})

		It("Should block removing uid field when set", func() {
			contactpoint := newContactPoint("existing-uid", "existing-uid")
			By("Creating ContactPoint with existing UID")
			Expect(k8sClient.Create(ctx, contactpoint)).To(Succeed())

			By("And setting UID to ''")
			contactpoint.Spec.CustomUID = ""
			Expect(k8sClient.Update(ctx, contactpoint)).To(HaveOccurred())
		})

		It("Should block changing value of uid", func() {
			contactpoint := newContactPoint("removing-uid", "existing-uid")
			By("Create new ContactPoint with existing UID")
			Expect(k8sClient.Create(ctx, contactpoint)).To(Succeed())

			By("Changing the existing UID")
			contactpoint.Spec.CustomUID = "new-contactpoint-uid"
			Expect(k8sClient.Update(ctx, contactpoint)).To(HaveOccurred())
		})
	})
})
