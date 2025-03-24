#!/usr/bin/lua

local datasource_template = [[
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDatasource
metadata:
  name: grafanadatasource-sample-%d
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana-a"
  plugins:
    - name: grafana-clock-panel
      version: 1.3.0
  datasource:
    name: prometheus-%d
    type: prometheus
    access: proxy
    url: http://prometheus-service:9090
    isDefault: true
    jsonData:
      'tlsSkipVerify': true
      'timeInterval': "5s"

]]

local dashboard_template = [[
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaDashboard
metadata:
  name: grafanadashboard-sample-%d
spec:
  instanceSelector:
    matchLabels:
      dashboards: "grafana-a"
  json: >
    {
      "id": null,
      "title": "Simple Dashboard %d",
      "tags": [],
      "style": "dark",
      "timezone": "browser",
      "editable": true,
      "hideControls": false,
      "graphTooltip": 1,
      "panels": [],
      "time": {
        "from": "now-6h",
        "to": "now"
      },
      "timepicker": {
        "time_options": [],
        "refresh_intervals": []
      },
      "templating": {
        "list": []
      },
      "annotations": {
        "list": []
      },
      "refresh": "5s",
      "schemaVersion": 17,
      "version": 0,
      "links": []
    }
]]


for i=1,100 do
    local dashboard_yaml = string.format(dashboard_template, i, i)
    local dashboard_file = io.open(string.format("./dashboards/dashboard-%d.yaml", i), "w")
    io.output(dashboard_file)
    io.write(dashboard_yaml)
    io.close(dashboard_file)

    local datasource_yaml = string.format(datasource_template, i, i)
    local datasource_file = io.open(string.format("./datasources/datasource-%d.yaml", i), "w")
    io.output(datasource_file)
    io.write(datasource_yaml)
    io.close(datasource_file)
end
