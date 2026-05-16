package config

import "github.com/grafana/grafana-operator/v5/pkg/featureflags"

// var FoldersUseNewAPI = featureflags.FeatureFlag{
// 	Name:         "FoldersUseNewAPI",
// 	IsActive:     false,
// 	IsDeprecated: false,
// 	Description:  "Use Grafana 13+ API-server style APIs for the folder controller",
// }

var FeatureFlags = featureflags.FeatureFlags{
	// FoldersUseNewAPI.Name: &FoldersUseNewAPI,
}
