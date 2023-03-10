local grafana = import 'grafonnet/grafana.libsonnet';
local dashboard = grafana.dashboard;
local template = grafana.template;
local singlestat = grafana.singlestat;
local graphPanel = grafana.graphPanel;
local prometheus = grafana.prometheus;

local buildInfo =
  singlestat.new(
    title='Version',
    datasource='Prometheus',
    format='none',
    valueName='name',
  ).addTarget(
    prometheus.target(
      'prometheus_build_info{instance="$instance"}',
      legendFormat='{{ version }}',
    )
  );

local systemLoad =
  singlestat.new(
    title='5m system load',
    datasource='Prometheus',
    format='none',
    valueName='current',
    decimals=2,
    sparklineShow=true,
    colorValue=true,
    thresholds='4,6',
  ).addTarget(
    prometheus.target(
      'node_load5{instance="$instance"}',
    )
  );

local networkTraffic =
  graphPanel.new(
    title='Network traffic on eth0',
    datasource='Prometheus',
    linewidth=2,
    format='Bps',
    aliasColors={
      Rx: 'light-green',
      Tx: 'light-red',
    },
  ).addTarget(
    prometheus.target(
      'rate(node_network_receive_bytes_total{instance="$instance",device="eth0"}[1m])',
      legendFormat='Rx',
    )
  ).addTarget(
    prometheus.target(
      'irate(node_network_transmit_bytes_total{instance="$instance",device="eth0"}[1m]) * (-1)',
      legendFormat='Tx',
    )
  );

dashboard.new(
  'Prometheus test',
  tags=['prometheus'],
  schemaVersion=18,
  editable=true,
  time_from='now-1h',
  refresh='1m',
)
.addTemplate(
  template.datasource(
    'PROMETHEUS_DS',
    'prometheus',
    'Prometheus',
    hide='label',
  )
)
.addTemplate(
  template.new(
    'instance',
    '$PROMETHEUS_DS',
    'label_values(prometheus_build_info, instance)',
    label='Instance',
    refresh='time',
  )
)

.addPanels(
  [
    buildInfo { gridPos: { h: 4, w: 3, x: 0, y: 0 } },

    systemLoad { gridPos: { h: 4, w: 4, x: 3, y: 0 } },

    networkTraffic { gridPos: { h: 8, w: 7, x: 0, y: 4 } },
  ]
)
