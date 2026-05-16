package featureflags

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var ErrUknownFeatureFlag = errors.New("unknown feature flag")

type FeatureFlag struct {
	Name         string
	IsActive     bool
	IsDeprecated bool
	Description  string
}

func (ff FeatureFlag) String() string {
	return fmt.Sprintf("%s: %t", ff.Name, ff.IsActive)
}

type FeatureFlags map[string]*FeatureFlag

func (ffs FeatureFlags) SetActive(name string) error {
	if _, ok := ffs[name]; !ok {
		return ErrUknownFeatureFlag
	}

	ffs[name].IsActive = true

	return nil
}

func (ffs FeatureFlags) SetActiveFromArg(arg string) error {
	toActivate := []string{}
	unknown := []string{}

	for name := range strings.SplitSeq(arg, ",") {
		if name == "" {
			continue
		}

		if _, ok := ffs[name]; !ok {
			unknown = append(unknown, name)
			continue
		}

		toActivate = append(toActivate, name)
	}

	if len(unknown) > 0 {
		return fmt.Errorf("unknown feature flags: %s", strings.Join(unknown, ","))
	}

	for _, name := range toActivate {
		err := ffs.SetActive(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ffs FeatureFlags) String() string {
	withStatus := make([]string, 0, len(ffs))

	for _, v := range ffs {
		withStatus = append(withStatus, v.String())
	}

	sort.Strings(withStatus)

	return strings.Join(withStatus, ", ")
}
