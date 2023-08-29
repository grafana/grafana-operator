local datasources = import './datasources.libsonnet';
local g = import './g.libsonnet';
local envs = import './envs.libsonnet';

local query = g.query;
local prometheusQuery = query.prometheus;

{
  readyPods:
    prometheusQuery.new(
      envs.PROMETHEUS_DS_UID,
      |||
        avg by (container) (100 * kube_pod_container_status_ready{namespace="$namespace", pod=~"pod-.*", job="kube-state-metrics"})[1m]
      |||
    ) + prometheusQuery.withLegendFormat('{{`{{ pod }}`}}'),

  totalRestartsPerContainer:
    prometheusQuery.new(
      envs.PROMETHEUS_DS_UID,
      |||
        sum by (container) (kube_pod_container_status_restarts_total{namespace="$namespace", pod=~"pod-.*", job="kube-state-metrics"})
      |||
    ) + prometheusQuery.withLegendFormat('{{`{{ container }}`}}')
    + prometheusQuery.withEditorMode('code'),
}
