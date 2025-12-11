local grafonnet = import "github.com/grafana/grafonnet/gen/grafonnet-latest/main.libsonnet";
local dashboard = grafonnet.dashboard;
local datasource = dashboard.variable.datasource;
local timeSeries = grafonnet.panel.timeSeries;
local prometheus = grafonnet.query.prometheus;

local prometheusDatasource = datasource.new('my-datasource', 'prometheus')
    + datasource.generalOptions.withLabel('Prometheus')
    + datasource.generalOptions.withDescription('Some prometheus instance you want to use')
    + datasource.generalOptions.showOnDashboard.withNothing();

dashboard.new("Your Dashboard")
    + dashboard.withDescription("My fancy Jsonnet dashboard")
    + dashboard.withEditable(false)
    + dashboard.withTags(["tag1", "tag2", "tag3"])
    + dashboard.withVariables([prometheusDatasource])
    + dashboard.withPanels([
        timeSeries.new("CPU Requested")
        + timeSeries.panelOptions.withDescription("Container CPU resource request")
        + timeSeries.queryOptions.withTargets([
            prometheus.new('$%s' % prometheusDatasource.name, |||
              kube_pod_container_resource_requests{resource="cpu",exported_namespace="some-namespace",exported_pod="some-pod"}
            |||)
            + prometheus.withLegendFormat("Requested by {{ exported_pod }}/{{ exported_container }}"),
        ])
        + timeSeries.standardOptions.withNoValue("N/A")
        + timeSeries.options.tooltip.withMode("multi")
        + timeSeries.options.tooltip.withSort("desc"),
    ])
    + dashboard.withTimezone("browser")
    + dashboard.graphTooltip.withSharedCrosshair()
