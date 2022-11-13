#!/usr/bin/lua

local template = [[
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
    local yaml = string.format(template, i, i)
    local file = io.open(string.format("./dashboards/dashboard-%d.yaml", i), "w")
    io.output(file)
    io.write(yaml)
    io.close(file)
end