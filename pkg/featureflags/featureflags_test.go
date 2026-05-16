package featureflags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getFF(t *testing.T, name string, isActive bool) *FeatureFlag {
	t.Helper()

	ff := &FeatureFlag{
		Name:     name,
		IsActive: isActive,
	}

	return ff
}

func TestFeatureFlagsSetActive(t *testing.T) {
	const name = "available"

	ffs := FeatureFlags{
		name: getFF(t, name, false),
	}

	err := ffs.SetActive(name)
	require.NoError(t, err)
	assert.True(t, ffs[name].IsActive)

	err = ffs.SetActive("unknown")
	require.ErrorIs(t, err, ErrUknownFeatureFlag)
}

func TestFeatureFlagsSetActiveFromArg(t *testing.T) {
	const name1 = "ff1"
	const name2 = "ff2"

	tests := []struct {
		name      string
		arg       string
		ffs       FeatureFlags
		want      FeatureFlags
		wantError bool
	}{
		{
			name:      "empty arg",
			arg:       "",
			ffs:       FeatureFlags{},
			want:      FeatureFlags{},
			wantError: false,
		},
		{
			name: "valid flag",
			arg:  fmt.Sprintf("%s", name1),
			ffs: FeatureFlags{
				name1: getFF(t, name1, false),
			},
			want: FeatureFlags{
				name1: getFF(t, name1, true),
			},
		},
		{
			name: "valid flags",
			arg:  fmt.Sprintf("%s,%s", name1, name2),
			ffs: FeatureFlags{
				name1: getFF(t, name1, false),
				name2: getFF(t, name2, false),
			},
			want: FeatureFlags{
				name1: getFF(t, name1, true),
				name2: getFF(t, name2, true),
			},
		},
		{
			name: "uknown flag",
			arg:  "unknown",
			ffs: FeatureFlags{
				name1: getFF(t, name1, false),
			},
			want: FeatureFlags{
				name1: getFF(t, name1, false),
			},
			wantError: true,
		},
		{
			name: "unknown flag amongst correct flags",
			arg:  fmt.Sprintf("%s,%s,unknown", name1, name2),
			ffs: FeatureFlags{
				name1: getFF(t, name1, false),
			},
			want: FeatureFlags{
				name1: getFF(t, name1, false),
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ffs.SetActiveFromArg(tt.arg)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			got := tt.ffs

			assert.Equal(t, tt.want, got)
		})
	}
}
