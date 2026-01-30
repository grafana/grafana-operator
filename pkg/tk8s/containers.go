package tk8s

import (
	"context"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func GetGrafanaTestContainer(t tHelper, ctx context.Context, image string) testcontainers.Container {
	t.Helper()

	c, err := testcontainers.Run(ctx, image,
		testcontainers.WithExposedPorts(
			"3000/tcp",
		),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForHTTP("/").WithStartupTimeout(2*time.Minute),
		),
	)

	require.NoError(t, err)

	return c
}
