package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	DefaultTLSConfiguration  = &tls.Config{MinVersion: tls.VersionTLS12}
	InsecureTLSConfiguration = &tls.Config{MinVersion: tls.VersionTLS12, InsecureSkipVerify: true} // #nosec G402 - Linter disabled because InsecureSkipVerify is the wanted behavior for this variable
)

// build the tls.Config object based on the content of the Grafana CR object
func buildTLSConfiguration(ctx context.Context, c client.Client, grafana *v1beta1.Grafana) (*tls.Config, error) {
	var tlsConfigBlock *v1beta1.TLSConfig

	switch {
	case grafana.Spec.Client != nil && grafana.Spec.Client.TLS != nil:
		// prefer top level if set, fall back to deprecated field
		tlsConfigBlock = grafana.Spec.Client.TLS
	case grafana.Spec.External != nil && grafana.Spec.External.TLS != nil:
		// fall back to external tls field if set
		tlsConfigBlock = grafana.Spec.External.TLS
	default:
		// if nothing is specified, ignore tls settings
		return nil, nil
	}

	if tlsConfigBlock.InsecureSkipVerify {
		return InsecureTLSConfiguration, nil
	}

	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	secretName := tlsConfigBlock.CertSecretRef.Name

	secretNamespace := grafana.Namespace
	if tlsConfigBlock.CertSecretRef.Namespace != "" {
		secretNamespace = tlsConfigBlock.CertSecretRef.Namespace
	}

	secret := &v1.Secret{}
	selector := client.ObjectKey{
		Name:      secretName,
		Namespace: secretNamespace,
	}

	err := c.Get(ctx, selector, secret)
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("empty credential secret: %v/%v", grafana.Namespace, tlsConfigBlock.CertSecretRef.Name)
	}

	crt, crtPresent := secret.Data["tls.crt"]
	key, keyPresent := secret.Data["tls.key"]

	if (crtPresent && !keyPresent) || (keyPresent && !crtPresent) {
		return nil, fmt.Errorf("invalid secret %v/%v. tls.crt and tls.key needs to be present together when one of them is declared", tlsConfigBlock.CertSecretRef.Namespace, tlsConfigBlock.CertSecretRef.Name)
	} else if crtPresent && keyPresent {
		loadedCrt, err := tls.X509KeyPair(crt, key)
		if err != nil {
			return nil, fmt.Errorf("certificate from secret %v/%v cannot be parsed : %w", grafana.Namespace, tlsConfigBlock.CertSecretRef.Name, err)
		}

		tlsConfig.Certificates = []tls.Certificate{loadedCrt}
	}

	if ca, ok := secret.Data["ca.crt"]; ok {
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to add ca.crt from the secret %s/%s", tlsConfigBlock.CertSecretRef.Namespace, tlsConfigBlock.CertSecretRef.Name)
		}

		tlsConfig.RootCAs = caCertPool
	}

	return tlsConfig, nil
}
