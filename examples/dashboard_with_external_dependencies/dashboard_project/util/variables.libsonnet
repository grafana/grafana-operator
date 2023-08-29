local g = import './g.libsonnet';
local datasources = import './datasources.libsonnet';
local envs = import './envs.libsonnet';

local var = g.dashboard.variable;

{
   namespace:
    var.constant.new('namespace', [envs.K8S_NAMESPACE]),

   promteheus_metrics_aggegation_time:
    var.interval.new('agg_time', ['1m', '5m', '10m', '15m', '30m', '1h', '3h', '6h', '12h', '1d', '3d', '1w', '1m'])
    + var.interval.generalOptions.withLabel('Metrics aggregation time')
}
