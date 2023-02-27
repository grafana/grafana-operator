# Grafana-operator

## Cross namespace grafana instances

As described in [#44](https://github.com/grafana-operator/grafana-operator-experimental/issues/44) we didn't want it
to be to easy to get access to a grafana datasource that wasn't defined the same namespace as the grafana instance.

To solve this we introduced `spec.allowCrossNamespaceImport` option to, dashboards, datasources and folders to be false by default.
This setting makes it so a grafana instance in another namespace don't get the grafana resources applied to it even if the label matches.

This is because especially the data sources contain secret information and we don't want another team to be able to use your datasource unless defined to do so in both CR:s.
