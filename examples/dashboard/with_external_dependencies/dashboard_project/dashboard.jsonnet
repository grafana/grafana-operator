local datasources = import './util/datasources.libsonnet';
local envs = import './util/envs.libsonnet';
local g = import './util/g.libsonnet';
local panels = import './util/panels.libsonnet';
local queries = import './util/queries.libsonnet';
local styles = import './util/styles.libsonnet';
local vars = import './util/variables.libsonnet';

local dashboard = g.dashboard;
local annotation = g.dashboard.annotation;


dashboard.new('Overview dashboard')
+ dashboard.withUid('customer-specified-uid')
+ dashboard.withDescription('Description')
+ dashboard.withTimezone('')
+ dashboard.time.withFrom('now-150m')
+ dashboard.time.withTo('now')
+ dashboard.withVariables([
  vars.namespace,
  vars.promteheus_metrics_aggegation_time,
])
+ dashboard.withAnnotations([
  annotation.withBuiltIn(1)
  + annotation.withDatasource(datasources.dashboardDatasource)
  + annotation.withDatasource(datasources.dashboardDatasource)
  + annotation.withEnable(1)
  + annotation.withHide(1)
  + annotation.withType('dashboard')
  + annotation.withTarget(
    annotation.target.withLimit(100)
    + annotation.target.withMatchAny(0)
    + annotation.target.withTags([])
    + annotation.target.withType('dashboard')
  ),
])
+ dashboard.withEditable(1)
+ dashboard.withPanels([
  panels.heatMap.tsBucketsHeatMap('Ready Pods', [
    queries.readyPods,
  ], datasources.prometheusDatasource, styles.heatMap.tsBucket.color, { gridPos: { h: 7, w: 12, x: 0, y: 12 } }),
  panels.timeSeries.base('Total Restarts Per Container', [
    queries.totalRestartsPerContainer,
  ], datasources.prometheusDatasource, styles.timeSeries.shorts, { gridPos: { h: 7, w: 12, x: 0, y: 19 } }),
])
