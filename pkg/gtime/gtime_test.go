package gtime

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseDuration(t *testing.T) {
	tcs := []struct {
		inp      string
		duration time.Duration
		err      *regexp.Regexp
	}{
		{inp: "1s", duration: time.Second},
		{inp: "1m", duration: time.Minute},
		{inp: "1h", duration: time.Hour},
		{inp: "1d", duration: 24 * time.Hour},
		{inp: "1w", duration: 7 * 24 * time.Hour},
		{inp: "2w", duration: 2 * 7 * 24 * time.Hour},
		{inp: "1M", duration: time.Duration(730.5 * float64(time.Hour))},
		{inp: "1y", duration: 365.25 * 24 * time.Hour},
		{inp: "5y", duration: 5 * 365.25 * 24 * time.Hour},
		{inp: "invalid-duration", err: regexp.MustCompile(`^time: invalid duration "?invalid-duration"?$`)},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res, err := ParseDuration(tc.inp)
			if tc.err == nil {
				require.NoError(t, err, "input %q", tc.inp)
				require.Equal(t, tc.duration, res, "input %q", tc.inp)
			} else {
				require.Error(t, err, "input %q", tc.inp)
				require.Regexp(t, tc.err, err.Error())
			}
		})
	}
}

func calculateDays() (int, int) {
	now := time.Now().UTC()
	currentYear, currentMonth, currentDay := now.Date()

	firstDayOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := firstDayOfMonth.AddDate(0, 1, -1).Day()

	t1 := time.Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0, time.UTC)
	t2 := t1.AddDate(1, 0, 0)

	daysInYear := int(t2.Sub(t1).Hours() / 24)

	return daysInMonth, daysInYear
}

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantDur      time.Duration
		wantPeriod   string
		wantErrRegex *regexp.Regexp
	}{
		{
			name:       "simple duration seconds",
			input:      "30s",
			wantDur:    30 * time.Second,
			wantPeriod: "",
		},
		{
			name:       "simple duration minutes",
			input:      "5m",
			wantDur:    5 * time.Minute,
			wantPeriod: "",
		},
		{
			name:       "simple duration minutes and seconds",
			input:      "1m30s",
			wantDur:    90 * time.Second,
			wantPeriod: "",
		},
		{
			name:       "complex duration",
			input:      "1h30m",
			wantDur:    90 * time.Minute,
			wantPeriod: "",
		},
		{
			name:       "days unit",
			input:      "7d",
			wantDur:    7,
			wantPeriod: "d",
		},
		{
			name:       "weeks unit",
			input:      "2w",
			wantDur:    2,
			wantPeriod: "w",
		},
		{
			name:       "months unit",
			input:      "3M",
			wantDur:    3,
			wantPeriod: "M",
		},
		{
			name:       "years unit",
			input:      "1y",
			wantDur:    1,
			wantPeriod: "y",
		},
		{
			name:         "invalid duration",
			input:        "invalid",
			wantErrRegex: regexp.MustCompile(`time: invalid duration "?invalid"?`),
		},
		{
			name:         "invalid number",
			input:        "abc1d",
			wantErrRegex: regexp.MustCompile(`time: invalid duration "?abc1d"?`),
		},
		{
			name:         "empty string",
			input:        "",
			wantErrRegex: regexp.MustCompile(`empty input`),
		},
		{
			name:       "fast path - pure number with date unit",
			input:      "30d",
			wantDur:    30,
			wantPeriod: "d",
		},
		{
			name:       "fast path - pure number with week unit",
			input:      "2w",
			wantDur:    2,
			wantPeriod: "w",
		},
		{
			name:       "fast path - pure number with month unit",
			input:      "6M",
			wantDur:    6,
			wantPeriod: "M",
		},
		{
			name:       "fast path - pure number with year unit",
			input:      "5y",
			wantDur:    5,
			wantPeriod: "y",
		},
		{
			name:         "non-numeric prefix with date unit",
			input:        "a5d",
			wantErrRegex: regexp.MustCompile(`time: invalid duration "?a5d"?`),
		},
		{
			name:         "mixed characters with date unit",
			input:        "5a3d",
			wantErrRegex: regexp.MustCompile(`time: unknown unit "a" in duration "5a3d"`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDur, gotPeriod, err := parse(tt.input)
			if tt.wantErrRegex != nil {
				require.Error(t, err)
				require.Regexp(t, tt.wantErrRegex, err.Error())
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantDur, gotDur)
			require.Equal(t, tt.wantPeriod, gotPeriod)
		})
	}
}
