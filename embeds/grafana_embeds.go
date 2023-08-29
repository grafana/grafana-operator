package embeds

import "embed"

//go:embed grafonnet-lib
var GrafonnetEmbed embed.FS

//go:embed testing/dashboard.jsonnet
var TestDashboardEmbed []byte

//go:embed testing/dashboard.json
var TestDashboardEmbedExpectedJSON []byte

//go:embed testing/dashboard_with_envs.jsonnet
var TestDashboardEmbedWithEnv []byte

//go:embed testing/dashboard_with_provided_envs.json
var TestDashboardEmbedWithEnvExpectedJSON []byte

//go:embed testing/jsonnetProjectWithRuntimeRaw.tar.gz
var TestJsonnetProjectBuildFolderGzip []byte
