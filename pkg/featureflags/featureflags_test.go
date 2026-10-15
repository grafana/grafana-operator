package featureflags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFeatureFlagParsing(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		valid   map[featureFlag]string
		want    map[featureFlag]bool
		wantErr string
	}{
		{
			name: "Empty arg",
			arg:  "",
			valid: map[featureFlag]string{
				"flag1": "",
			},
			want: map[featureFlag]bool{},
		},
		{
			name: "Valid feature flags",
			arg:  "flag1,flag2",
			valid: map[featureFlag]string{
				"flag1": "",
				"flag2": "",
				"flag3": "",
			},
			want: map[featureFlag]bool{
				"flag1": true,
				"flag2": true,
			},
		},
		{
			name: "Invalid feature flag",
			arg:  "flag1,flag2",
			valid: map[featureFlag]string{
				"flag1": "",
			},
			wantErr: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(tt.arg, tt.valid)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, "unknown feature flag")
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
