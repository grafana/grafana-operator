package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LogMsgParsingPatches    = "failed to parse patch scripts"
	LogMsgResolvingPatchEnv = "failed to resolve patch environment"
)

func ParsePatches(p *v1beta1.Patch) ([]*gojq.Query, error) {
	if p == nil {
		return []*gojq.Query{}, nil
	}

	patches := make([]*gojq.Query, len(p.Scripts))
	for idx, s := range p.Scripts {
		q, err := gojq.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("script %d failed to parse: %w", idx, err)
		}

		patches[idx] = q
	}

	return patches, nil
}

type patchEnvResolver func(*v1beta1.Grafana) (string, error)

func CollectPatchEnv(ctx context.Context, cl client.Client, namespace string, refs []v1beta1.PatchEnvVar) ([]patchEnvResolver, error) {
	out := make([]patchEnvResolver, len(refs))
	for idx, r := range refs {
		switch {
		case r.ValueFrom.SecretKeyRef != nil:
			value, _, err := getSecretValue(ctx, cl, namespace, r.ValueFrom.SecretKeyRef)
			if err != nil {
				return nil, err
			}

			out[idx] = func(g *v1beta1.Grafana) (string, error) {
				return fmt.Sprintf("%s=%s", r.Name, value), nil
			}
		case r.ValueFrom.ConfigMapKeyRef != nil:
			value, _, err := getConfigMapValue(ctx, cl, namespace, r.ValueFrom.ConfigMapKeyRef)
			if err != nil {
				return nil, err
			}

			out[idx] = func(g *v1beta1.Grafana) (string, error) {
				return fmt.Sprintf("%s=%s", r.Name, value), nil
			}
		case r.ValueFrom.GrafanaRef != nil:
			out[idx] = func(g *v1beta1.Grafana) (string, error) {
				value, _, err := getGrafanaRefValue(g, r.ValueFrom.GrafanaRef)
				if err != nil {
					return "", fmt.Errorf("getting grafana field value: %w", err)
				}

				return fmt.Sprintf("%s=%s", r.Name, value), nil
			}
		}
	}

	return out, nil
}

func ApplyPatch(patches []*gojq.Query, resource map[string]any, env []string) (map[string]any, error) {
	var work map[string]any

	work = resource

	for _, q := range patches {
		code, err := gojq.Compile(q, gojq.WithEnvironLoader(func() []string {
			return env
		}))
		if err != nil {
			return nil, fmt.Errorf("compiling query: %w", err)
		}

		iter := code.Run(work)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}

			if err, ok := v.(error); ok {
				haltError := &gojq.HaltError{}
				if errors.As(err, &haltError) {
					break
				}

				return nil, fmt.Errorf("evaluating expression: %w", err)
			}

			typed, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid return type: %t", v)
			}

			work = typed
		}
	}

	return work, nil
}
