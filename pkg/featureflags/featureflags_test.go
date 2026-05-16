package featureflags

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagsIsActive(t *testing.T) {
	fActive := FeatureFlag("active")
	fInactive := FeatureFlag("inactive")

	ffs := FeatureFlags{
		active: map[FeatureFlag]bool{
			fActive:   true,
			fInactive: false,
		},
	}

	assert.True(t, ffs.IsActive(fActive))
	assert.False(t, ffs.IsActive(fInactive))
}

func TestFeatureFlagsIsAvailable(t *testing.T) {
	fAvailable := FeatureFlag("available")
	fNonAvailable := FeatureFlag("non-available")

	ffs := FeatureFlags{
		available: map[FeatureFlag]string{
			fAvailable: "test flag",
		},
	}

	assert.True(t, ffs.IsAvailable(fAvailable))
	assert.False(t, ffs.IsAvailable(fNonAvailable))
}

func TestFeatureFlagsSetActive(t *testing.T) {
	fAvailable := FeatureFlag("available")
	fNonAvailable := FeatureFlag("non-available")

	ffs := FeatureFlags{
		available: map[FeatureFlag]string{
			fAvailable: "test flag",
		},
		active: map[FeatureFlag]bool{},
	}

	err := ffs.SetActive(fAvailable)
	require.NoError(t, err)

	err = ffs.SetActive(fNonAvailable)
	require.ErrorIs(t, err, ErrUknownFeatureFlag)
}

func TestFeatureFlagsSetActiveFromArg(t *testing.T) {
	fAvailable1 := FeatureFlag("available1")
	fAvailable2 := FeatureFlag("available2")

	tests := []struct {
		name       string
		arg        string
		wantActive map[FeatureFlag]bool
		wantError  bool
	}{
		{
			name:       "empty arg",
			arg:        "",
			wantActive: map[FeatureFlag]bool{},
			wantError:  false,
		},
		{
			name: "valid flag",
			arg:  fmt.Sprintf("%s", fAvailable1),
			wantActive: map[FeatureFlag]bool{
				fAvailable1: true,
			},
		},
		{
			name: "valid flags",
			arg:  fmt.Sprintf("%s,%s", fAvailable1, fAvailable2),
			wantActive: map[FeatureFlag]bool{
				fAvailable1: true,
				fAvailable2: true,
			},
		},
		{
			name:       "invalid flag",
			arg:        "non-available",
			wantActive: map[FeatureFlag]bool{},
			wantError:  true,
		},
		{
			name:       "invalid flag amongst correct flags",
			arg:        fmt.Sprintf("%s,%s,non-available", fAvailable1, fAvailable2),
			wantActive: map[FeatureFlag]bool{},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ffs := FeatureFlags{
				available: map[FeatureFlag]string{
					fAvailable1: "available flag 1",
					fAvailable2: "available flag 2",
				},
				active: map[FeatureFlag]bool{},
			}

			err := ffs.SetActiveFromArg(tt.arg)
			if tt.wantError {
				require.Error(t, err)
			}

			assert.Equal(t, tt.wantActive, ffs.active)
		})
	}
}

func TestNewFeatureFlags(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		got := NewFeatureFlags(
			nil,
		)
		want := FeatureFlags{
			available: map[FeatureFlag]string{},
			active:    map[FeatureFlag]bool{},
		}

		assert.Equal(t, got, want)
	})

	t.Run("empty", func(t *testing.T) {
		got := NewFeatureFlags(
			map[FeatureFlag]string{},
		)
		want := FeatureFlags{
			available: map[FeatureFlag]string{},
			active:    map[FeatureFlag]bool{},
		}

		assert.Equal(t, got, want)
	})

	t.Run("defined", func(t *testing.T) {
		availableFlags := map[FeatureFlag]string{
			FeatureFlag("flag1"): "test flag",
		}

		want := FeatureFlags{
			available: availableFlags,
			active:    map[FeatureFlag]bool{},
		}

		got := NewFeatureFlags(
			availableFlags,
		)

		assert.Equal(t, got, want)
	})
}
