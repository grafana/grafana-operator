package controllers

import (
	"context"
	"fmt"

	"github.com/grafana/grafana-operator/v5/api/v1beta1"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CollectPatchEnv(ctx context.Context, cl client.Client, namespace string, instance *v1beta1.Grafana, refs []v1beta1.PatchEnvVar) ([]string, error) {
	out := make([]string, len(refs))
	for idx, r := range refs {
		val, _, err := getReferencedPatchValue(ctx, cl, namespace, r.ValueFrom, instance)
		if err != nil {
			return []string{}, fmt.Errorf("resolving patch env: %w", err)
		}

		out[idx] = fmt.Sprintf("%s=%s", r.Name, val)
	}

	return out, nil
}

func ApplyPatch(patch *v1beta1.Patch, resource map[string]any, env []string) (map[string]any, error) {
	if patch == nil {
		return resource, nil
	}

	var work any

	envLoader := func() []string {
		return env
	}
	work = resource

	for _, s := range patch.Scripts {
		q, err := gojq.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("parsing query: %w", err)
		}

		code, err := gojq.Compile(q, gojq.WithEnvironLoader(envLoader))
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
				if err, ok := err.(*gojq.HaltError); ok && err.Value() == nil {
					break
				}

				return nil, fmt.Errorf("evaluating expression: %w", err)
			}

			if _, ok := v.(map[string]any); !ok {
				return nil, fmt.Errorf("invalid return type: %t", v)
			}

			work = v
		}
	}

	return work.(map[string]any), nil
}
