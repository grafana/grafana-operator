local grafonnet = import "github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet";
local datasources = import "datasources.libsonnet";
local panels = import "panels.libsonnet";
local dashboard = grafonnet.dashboard;

function(title="Your Dashboard", timezone="browser")
  dashboard.new(title)
    + dashboard.withDescription("My fancy Jsonnet dashboard")
    + dashboard.withEditable(false)
    + dashboard.withTags(["tag1", "tag2", "tag3"])
    + dashboard.withVariables([datasources.prometheus])
    + dashboard.withPanels([panels.cpuRequested])
    + dashboard.withTimezone(timezone)
    + dashboard.graphTooltip.withSharedCrosshair()
