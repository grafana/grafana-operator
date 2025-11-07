package client

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetValueFromSecretKey(ctx context.Context, ref *corev1.SecretKeySelector, c client.Client, namespace string) ([]byte, error) {
	if ref == nil {
		return nil, errors.New("empty secret key selector")
	}

	secret := &corev1.Secret{}
	selector := client.ObjectKey{
		Name:      ref.Name,
		Namespace: namespace,
	}

	err := c.Get(ctx, selector, secret)
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("empty credential secret: %v/%v", namespace, ref.Name)
	}

	if val, ok := secret.Data[ref.Key]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("credentials not found in secret: %v/%v", namespace, ref.Name)
}
