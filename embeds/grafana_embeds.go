package embeds

import "embed"

//go:embed grafonnet-lib
var GrafonnetEmbed embed.FS

//go:embed testing/dashboard.jsonnet
var TestDashboardEmbed []byte

//go:embed testing/dashboard.json
var TestDashboardEmbedExpectedJSON []byte
