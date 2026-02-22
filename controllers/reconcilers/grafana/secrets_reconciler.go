package grafana

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"sort"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/grafana/grafana-operator/v5/controllers/reconcilers"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type SecretsReconciler struct {
	client client.Client
}

func NewSecretsReconciler(cl client.Client) reconcilers.OperatorGrafanaReconciler {
	return &SecretsReconciler{
		client: cl,
	}
}

// Reconcile collects the ResourceVersions of all Secrets and ConfigMaps referenced in the
// Grafana CR's deployment spec and external config, hashes them, and stores the result in
// vars.SecretsHash. The deployment reconciler then injects this as a SECRETS_HASH env var,
// so any secret rotation causes a pod template change and triggers a rolling restart.
func (r *SecretsReconciler) Reconcile(ctx context.Context, cr *v1beta1.Grafana, vars *v1beta1.OperatorReconcileVars, scheme *runtime.Scheme) (v1beta1.OperatorStageStatus, error) {
	log := logf.FromContext(ctx).WithName("SecretsReconciler")

	secretNames, configMapNames := ReferencedSecretsAndConfigMaps(cr)

	var resourceVersions []string

	for _, name := range secretNames {
		secret := &corev1.Secret{}

		err := r.client.Get(ctx, types.NamespacedName{Namespace: cr.Namespace, Name: name}, secret)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("referenced secret not found, skipping", "secret", name)

				continue
			}

			return v1beta1.OperatorStageResultFailed, fmt.Errorf("fetching secret %s: %w", name, err)
		}

		resourceVersions = append(resourceVersions, fmt.Sprintf("secret/%s=%s", name, secret.ResourceVersion))
	}

	for _, name := range configMapNames {
		cm := &corev1.ConfigMap{}

		err := r.client.Get(ctx, types.NamespacedName{Namespace: cr.Namespace, Name: name}, cm)
		if err != nil {
			if apierrors.IsNotFound(err) {
				log.Info("referenced configmap not found, skipping", "configmap", name)

				continue
			}

			return v1beta1.OperatorStageResultFailed, fmt.Errorf("fetching configmap %s: %w", name, err)
		}

		resourceVersions = append(resourceVersions, fmt.Sprintf("configmap/%s=%s", name, cm.ResourceVersion))
	}

	vars.SecretsHash = hashResourceVersions(resourceVersions)

	return v1beta1.OperatorStageResultSuccess, nil
}

// ReferencedSecretsAndConfigMaps returns deduplicated, sorted lists of Secret and ConfigMap names
// referenced in the Grafana CR's deployment spec (containers env/envFrom/volumes) and
// external/client TLS configuration.
func ReferencedSecretsAndConfigMaps(cr *v1beta1.Grafana) ([]string, []string) {
	secretSet := make(map[string]struct{})
	configMapSet := make(map[string]struct{})

	addSecret := func(name string) {
		if name != "" {
			secretSet[name] = struct{}{}
		}
	}

	addConfigMap := func(name string) {
		if name != "" {
			configMapSet[name] = struct{}{}
		}
	}

	collectDeploymentRefs(cr, addSecret, addConfigMap)
	collectExternalRefs(cr, addSecret)
	collectClientRefs(cr, addSecret)

	secretNames := make([]string, 0, len(secretSet))
	for name := range secretSet {
		secretNames = append(secretNames, name)
	}

	configMapNames := make([]string, 0, len(configMapSet))
	for name := range configMapSet {
		configMapNames = append(configMapNames, name)
	}

	sort.Strings(secretNames)
	sort.Strings(configMapNames)

	return secretNames, configMapNames
}

func collectDeploymentRefs(cr *v1beta1.Grafana, addSecret, addConfigMap func(string)) {
	if cr.Spec.Deployment == nil ||
		cr.Spec.Deployment.Spec.Template == nil ||
		cr.Spec.Deployment.Spec.Template.Spec == nil {
		return
	}

	podSpec := cr.Spec.Deployment.Spec.Template.Spec

	allContainers := make([]corev1.Container, 0, len(podSpec.Containers)+len(podSpec.InitContainers))
	allContainers = append(allContainers, podSpec.Containers...)
	allContainers = append(allContainers, podSpec.InitContainers...)

	for _, c := range allContainers {
		collectContainerEnvRefs(c, addSecret, addConfigMap)
	}

	for _, vol := range podSpec.Volumes {
		if vol.Secret != nil {
			addSecret(vol.Secret.SecretName)
		}

		if vol.ConfigMap != nil {
			addConfigMap(vol.ConfigMap.Name)
		}
	}
}

func collectContainerEnvRefs(c corev1.Container, addSecret, addConfigMap func(string)) {
	for _, env := range c.Env {
		if env.ValueFrom == nil {
			continue
		}

		if env.ValueFrom.SecretKeyRef != nil {
			addSecret(env.ValueFrom.SecretKeyRef.Name)
		}

		if env.ValueFrom.ConfigMapKeyRef != nil {
			addConfigMap(env.ValueFrom.ConfigMapKeyRef.Name)
		}
	}

	for _, envFrom := range c.EnvFrom {
		if envFrom.SecretRef != nil {
			addSecret(envFrom.SecretRef.Name)
		}

		if envFrom.ConfigMapRef != nil {
			addConfigMap(envFrom.ConfigMapRef.Name)
		}
	}
}

func collectExternalRefs(cr *v1beta1.Grafana, addSecret func(string)) {
	if cr.Spec.External == nil {
		return
	}

	ext := cr.Spec.External

	if ext.APIKey != nil {
		addSecret(ext.APIKey.Name)
	}

	if ext.AdminUser != nil {
		addSecret(ext.AdminUser.Name)
	}

	if ext.AdminPassword != nil {
		addSecret(ext.AdminPassword.Name)
	}

	if ext.TLS != nil && ext.TLS.CertSecretRef != nil {
		addSecret(ext.TLS.CertSecretRef.Name)
	}
}

func collectClientRefs(cr *v1beta1.Grafana, addSecret func(string)) {
	if cr.Spec.Client == nil || cr.Spec.Client.TLS == nil || cr.Spec.Client.TLS.CertSecretRef == nil {
		return
	}

	addSecret(cr.Spec.Client.TLS.CertSecretRef.Name)
}

// hashResourceVersions computes a stable SHA-256 hash over a sorted list of
// "kind/name=resourceVersion" strings. An empty list produces an empty string
// so that Grafana instances with no referenced secrets don't get a spurious hash.
func hashResourceVersions(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	sort.Strings(versions)

	h := sha256.New()

	for _, v := range versions {
		io.WriteString(h, v) //nolint:errcheck
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
