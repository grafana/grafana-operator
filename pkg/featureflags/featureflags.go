package featureflags

import (
	"fmt"
	"log/slog"
	"strings"
)

var activeFlags map[featureFlag]bool = make(map[featureFlag]bool)

type featureFlag string

const FoldersUseNewAPI featureFlag = "FoldersUseNewAPI"

// availableFlags is the source of truth of flags available to the user. The value in the map should contain documentation for the feature flag
var availableFlags = map[featureFlag]string{
	FoldersUseNewAPI: "Use Grafana 13+ API-server style APIs for the folder controller",
}

func SetActiveFromArg(arg string) error {
	active, err := parseFlags(arg, availableFlags)
	if err != nil {
		return err
	}

	activeFlags = active

	return nil
}

func parseFlags(arg string, existing map[featureFlag]string) (map[featureFlag]bool, error) {
	active := make(map[featureFlag]bool)

	for f := range strings.SplitSeq(arg, ",") {
		if f == "" {
			continue
		}

		ff := featureFlag(f)
		if _, ok := existing[ff]; !ok {
			return nil, fmt.Errorf("unknown feature flag: '%s'", ff)
		}

		slog.Info("activating feature flag", "flag", ff)
		active[ff] = true
	}

	return active, nil
}

func IsEnabled(flag featureFlag) bool {
	return activeFlags[flag]
}
