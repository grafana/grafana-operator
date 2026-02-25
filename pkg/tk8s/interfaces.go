package tk8s

import "github.com/stretchr/testify/require"

type tHelper interface {
	Helper()
	Skip(args ...any)
	require.TestingT
}
