package gtime

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseInterval(t *testing.T) {
	daysInMonth, daysInYear := calculateDays()

	tcs := []struct {
		inp      string
		duration time.Duration
		err      *regexp.Regexp
	}{
		{inp: "1d", duration: 24 * time.Hour},
		{inp: "1w", duration: 168 * time.Hour},
		{inp: "2w", duration: 2 * 168 * time.Hour},
		{inp: "1M", duration: time.Duration(daysInMonth * 24 * int(time.Hour))},
		{inp: "1y", duration: time.Duration(daysInYear * 24 * int(time.Hour))},
		{inp: "invalid-duration", err: regexp.MustCompile(`^time: invalid duration "?invalid-duration"?$`)},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res, err := ParseInterval(tc.inp)
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

func TestFormatInterval(t *testing.T) {
	testCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"61s", time.Second * 61, "1m"},
		{"30ms", time.Millisecond * 30, "30ms"},
		{"23h", time.Hour * 23, "23h"},
		{"24h", time.Hour * 24, "1d"},
		{"367d", time.Hour * 24 * 367, "1y"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, FormatInterval(tc.duration))
		})
	}
}

func TestGetIntervalFrom(t *testing.T) {
	testCases := []struct {
		name            string
		queryInterval   string
		queryIntervalMs int64
		defaultInterval time.Duration
		expected        time.Duration
	}{
		{"45s", "45s", 0, time.Second * 15, time.Second * 45},
		{"45", "45", 0, time.Second * 15, time.Second * 45},
		{"2m", "2m", 0, time.Second * 15, time.Minute * 2},
		{"1d", "1d", 0, time.Second * 15, time.Hour * 24},
		{"intervalMs", "", 45000, time.Second * 15, time.Second * 45},
		{"intervalMs sub-seconds", "", 45200, time.Second * 15, time.Millisecond * 45200},
		{"defaultInterval when interval empty", "", 0, time.Second * 15, time.Second * 15},
		{"defaultInterval when intervalMs 0", "", 0, time.Second * 15, time.Second * 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GetIntervalFrom(tc.queryInterval, "", tc.queryIntervalMs, tc.defaultInterval)
			require.Nil(t, err)
			require.Equal(t, tc.expected, actual)
		})
	}
}

func TestParseIntervalStringToTimeDuration(t *testing.T) {
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
		// ParseIntervalStringToTimeDuration specific conditions
		{inp: "10", duration: 10 * time.Second},
		{inp: "<10s>", duration: 10 * time.Second},
		{inp: "10s>", duration: 10 * time.Second},
		{inp: "<10s", duration: 10 * time.Second},
		{inp: "", err: regexp.MustCompile(`invalid interval`)},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res, err := ParseIntervalStringToTimeDuration(tc.inp)
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

func TestRoundInterval(t *testing.T) {
	tcs := []struct {
		input    time.Duration
		expected time.Duration
	}{
		{input: 9 * time.Millisecond, expected: time.Millisecond * 1},
		{input: 14 * time.Millisecond, expected: time.Millisecond * 10},
		{input: 34 * time.Millisecond, expected: time.Millisecond * 20},
		{input: 74 * time.Millisecond, expected: time.Millisecond * 50},
		{input: 140 * time.Millisecond, expected: time.Millisecond * 100},
		{input: 320 * time.Millisecond, expected: time.Millisecond * 200},
		{input: 740 * time.Millisecond, expected: time.Millisecond * 500},
		{input: 1400 * time.Millisecond, expected: time.Millisecond * 1000},
		{input: 3200 * time.Millisecond, expected: time.Millisecond * 2000},
		{input: 7400 * time.Millisecond, expected: time.Millisecond * 5000},
		{input: 12400 * time.Millisecond, expected: time.Millisecond * 10000},
		{input: 17250 * time.Millisecond, expected: time.Millisecond * 15000},
		{input: 23000 * time.Millisecond, expected: time.Millisecond * 20000},
		{input: 42000 * time.Millisecond, expected: time.Millisecond * 30000},
		{input: 85000 * time.Millisecond, expected: time.Millisecond * 60000},
		{input: 200000 * time.Millisecond, expected: time.Millisecond * 120000},
		{input: 420000 * time.Millisecond, expected: time.Millisecond * 300000},
		{input: 720000 * time.Millisecond, expected: time.Millisecond * 600000},
		{input: 1000000 * time.Millisecond, expected: time.Millisecond * 900000},
		{input: 1250000 * time.Millisecond, expected: time.Millisecond * 1200000},
		{input: 2500000 * time.Millisecond, expected: time.Millisecond * 1800000},
		{input: 5200000 * time.Millisecond, expected: time.Millisecond * 3600000},
		{input: 8500000 * time.Millisecond, expected: time.Millisecond * 7200000},
		{input: 15000000 * time.Millisecond, expected: time.Millisecond * 10800000},
		{input: 30000000 * time.Millisecond, expected: time.Millisecond * 21600000},
		{input: 85000000 * time.Millisecond, expected: time.Millisecond * 43200000},
		{input: 150000000 * time.Millisecond, expected: time.Millisecond * 86400000},
		{input: 600000000 * time.Millisecond, expected: time.Millisecond * 86400000},
		{input: 1500000000 * time.Millisecond, expected: time.Millisecond * 604800000},
		{input: 3500000000 * time.Millisecond, expected: time.Millisecond * 2592000000},
		{input: 40000000000 * time.Millisecond, expected: time.Millisecond * 2592000000},
		{input: 70000000000 * time.Millisecond, expected: time.Millisecond * 31536000000},
	}
	for i, tc := range tcs {
		t.Run(fmt.Sprintf("testcase %d", i), func(t *testing.T) {
			res := RoundInterval(tc.input)
			require.Equal(t, tc.expected, res, "input %q", tc.input)
		})
	}
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
