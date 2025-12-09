local grafonnet = import "github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet";
local datasources = import "datasources.libsonnet";
local panels = import "panels.libsonnet";
local dashboard = grafonnet.dashboard;

function(title="Your Dashboard", timezone="browser")
  dashboard.new(title)
    + dashboard.withDescription(std.extVar('description'))
    + dashboard.withEditable(std.parseJson(std.extVar('editable')))
    + dashboard.withTags([std.extVar('some-tag')])
    + dashboard.withVariables([datasources.prometheus])
    + dashboard.withPanels([panels.cpuRequested])
    + dashboard.withTimezone(timezone)
    + dashboard.graphTooltip.withSharedCrosshair()
