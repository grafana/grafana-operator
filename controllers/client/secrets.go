package client

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetValueFromSecretKey(ctx context.Context, cl client.Client, namespace string, keySelector *corev1.SecretKeySelector) ([]byte, error) {
	if keySelector == nil {
		return nil, errors.New("empty secret key selector")
	}

	secret := &corev1.Secret{}

	selector := client.ObjectKey{
		Name:      keySelector.Name,
		Namespace: namespace,
	}

	err := cl.Get(ctx, selector, secret)
	if err != nil {
		return nil, err
	}

	if secret.Data == nil {
		return nil, fmt.Errorf("empty credential secret: %v/%v", namespace, keySelector.Name)
	}

	if val, ok := secret.Data[keySelector.Key]; ok {
		return val, nil
	}

	return nil, fmt.Errorf("credentials not found in secret: %v/%v", namespace, keySelector.Name)
}
