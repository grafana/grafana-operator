package featureflags

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type FeatureFlag string

var (
	ErrActiveNotInitialized = errors.New("field with active flags is not initialized")
	ErrUknownFeatureFlag    = errors.New("unknown feature flag")
)

type FeatureFlags struct {
	// available is the source of truth of flags available to the user. The value in the map should contain documentation for the feature flag
	available map[FeatureFlag]string
	active    map[FeatureFlag]bool
}

func (ffs *FeatureFlags) IsActive(ff FeatureFlag) bool {
	// TODO: handle not available
	return ffs.active[ff]
}

func (ffs *FeatureFlags) IsAvailable(ff FeatureFlag) bool {
	// TODO: handle not available
	_, isAvailable := ffs.available[ff]

	return isAvailable
}

func (ffs *FeatureFlags) SetActive(ff FeatureFlag) error {
	if ffs.active == nil {
		return ErrActiveNotInitialized
	}

	if ffs.IsAvailable(ff) {
		// TODO: add safety?
		ffs.active[ff] = true

		return nil
	}

	return ErrUknownFeatureFlag
}

func (ffs *FeatureFlags) SetActiveFromArg(arg string) error {
	toActivate := []FeatureFlag{}
	unknown := []string{}

	for f := range strings.SplitSeq(arg, ",") {
		if f == "" {
			continue
		}

		ff := FeatureFlag(f)

		if ffs.IsAvailable(ff) {
			toActivate = append(toActivate, ff)
		} else {
			unknown = append(unknown, f)
		}
	}

	if len(unknown) > 0 {
		return fmt.Errorf("unknown feature flags: %s", strings.Join(unknown, ","))
	}

	for _, ff := range toActivate {
		err := ffs.SetActive(ff)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ffs *FeatureFlags) String() string {
	available := slices.Sorted(maps.Keys(ffs.available))

	withStatus := make([]string, 0, len(available))

	for _, ff := range available {
		withStatus = append(
			withStatus, fmt.Sprintf("%s: %t", ff, ffs.IsActive(ff)),
		)
	}

	return strings.Join(withStatus, ", ")
}

func NewFeatureFlags(availableFlags map[FeatureFlag]string) FeatureFlags {
	if availableFlags == nil {
		availableFlags = map[FeatureFlag]string{}
	}

	ffs := FeatureFlags{
		available: availableFlags,
		active:    map[FeatureFlag]bool{},
	}

	return ffs
}
