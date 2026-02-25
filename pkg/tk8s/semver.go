package tk8s

import (
	"github.com/blang/semver/v4"
	"github.com/stretchr/testify/require"
)

func SkipTestIfInRange(t tHelper, semverRange, version string) {
	v, err := semver.Parse(version)
	require.NoError(t, err)

	rangeChecker, err := semver.ParseRange(semverRange)
	require.NoError(t, err)

	isInRange := rangeChecker(v)

	if isInRange {
		t.Skip()
	}
}
