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
			// grafana 13.3 runs its fresh-database migrations for well over 15s before /api/health responds
			wait.ForHTTP("/api/health").WithStartupTimeout(90*time.Second),
		),
	)

	require.NoError(t, err)

	return c
}
