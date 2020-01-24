package handlers

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"fmt"
	"html/template"

	"github.com/integr8ly/grafana-operator/v3/pkg/api/models"
	"github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultProxyNameFormat = "%sis"
	cryptNonce             = "bb8ef84243d2ee95a41c6c57"
)

func grafanaFromCRD(g *v1alpha1.Grafana) *models.Grafana {

	return &models.Grafana{
		Name: &g.Name,
		Config: models.GrafanaConfig{
			AdminUser:          g.Spec.Config.Security.AdminUser,
			IngressHost:        g.Spec.Ingress.Hostname,
			DisableSignoutMenu: *g.Spec.Config.Auth.DisableSignoutMenu,
			DisableLoginForm:   *g.Spec.Config.Auth.DisableLoginForm,
		},
	}
}

func getProxyHost(ingresshost, name string) (p string, err error) {
	if ingresshost == "" {
		return p, fmt.Errorf("No Grafana ingresshost provided")
	}
	n := fmt.Sprintf(defaultProxyNameFormat, name)
	p = fmt.Sprintf(ingresshost, n)
	return
}

func yamlUnmarshalHandler(in interface{}, out interface{}) error {
	var tpl bytes.Buffer
	h, err := yaml.Marshal(in)
	if err != nil {
		return err
	}

	t, err := template.New("config").Parse(string(h))
	err = t.Execute(&tpl, nil)

	return yaml.Unmarshal(h, out)
}

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

func createSecret(key, plaintext []byte) (ciphertext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	nonce, _ := hex.DecodeString(cryptNonce)
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}
	seal := aesgcm.Seal(nil, nonce, plaintext, nil)
	return hex.EncodeToString(seal), nil
}

func createNamespaceIfNotExists(client client.Client, ns string) (err error) {
	n := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}
	if err = client.Create(context.Background(), &n); err != nil {
		if errors.IsAlreadyExists(err) {
			return nil
		}
		return
	}
	return
}

func decrypteSecret(key []byte, secret string) (plaintext string, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	nonce, _ := hex.DecodeString(cryptNonce)
	ciphertext, err := hex.DecodeString(secret)
	if err != nil {
		return
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	plainbyte, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return
	}

	return string(plainbyte), nil
}
