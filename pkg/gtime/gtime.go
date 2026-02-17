package gtime

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var dateUnitPattern = regexp.MustCompile(`^(\d+)([dwMy])$`)

// ParseDuration parses a duration with support for all units that Grafana uses.
// Durations are independent of wall time.
func ParseDuration(inp string) (time.Duration, error) {
	dur, period, err := parse(inp)
	if err != nil {
		return 0, err
	}
	if period == "" {
		return dur, nil
	}

	// The average number of days in a year, using the Julian calendar
	const daysInAYear = 365.25
	const day = 24 * time.Hour
	const week = 7 * day
	const year = time.Duration(float64(day) * daysInAYear)
	const month = time.Duration(float64(year) / 12)

	switch period {
	case "d":
		return dur * day, nil
	case "w":
		return dur * week, nil
	case "M":
		return dur * month, nil
	case "y":
		return dur * year, nil
	}

	return 0, fmt.Errorf("invalid duration %q", inp)
}

func parse(inp string) (time.Duration, string, error) {
	if inp == "" {
		return 0, "", errors.New("empty input")
	}

	// Fast path for simple duration formats (no date units)
	lastChar := inp[len(inp)-1]
	if lastChar != 'd' && lastChar != 'w' && lastChar != 'M' && lastChar != 'y' {
		dur, err := time.ParseDuration(inp)
		return dur, "", err
	}

	// Check if the rest is a number for date units
	numPart := inp[:len(inp)-1]
	isNum := true
	for _, c := range numPart {
		if c < '0' || c > '9' {
			isNum = false
			break
		}
	}
	if isNum {
		num, err := strconv.Atoi(numPart)
		if err != nil {
			return 0, "", err
		}
		return time.Duration(num), string(lastChar), nil
	}

	// Fallback to regex for complex cases
	result := dateUnitPattern.FindStringSubmatch(inp)
	if len(result) != 3 {
		dur, err := time.ParseDuration(inp)
		return dur, "", err
	}

	num, err := strconv.Atoi(result[1])
	if err != nil {
		return 0, "", err
	}

	return time.Duration(num), result[2], nil
}
