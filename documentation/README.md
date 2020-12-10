## Grafana Operator Documentation

* [Installing Grafana](./deploy_grafana.md)
* [Dashboards](./dashboards.md)
* [Data Sources](./datasources.md)
* [Multi namespace support](./multi_namespace_support.md)
* [Mounting extra config files](./extra_files.md)
* [Jsonnet support](./jsonnet.md)
* [Env vars](./env_vars.md)

## Examples

The following example CRs are provided:

### Grafana deployments

* [Grafana.yaml](../deploy/examples/Grafana.yaml): Installs Grafana using the default configuration and an Ingress or Route.
* [GrafanaWithIngressHost.yaml](../deploy/examples/GrafanaWithIngressHost.yaml): Installs Grafana using the default configuration and an Ingress where the host is set for external access. 
* [ldap/Grafana.yaml](../deploy/examples/ldap/Grafana.yaml): Installs Grafana and sets up LDAP authentication. LDAP configuration is mounted from the configmap [ldap/ldap-config.yaml](../deploy/examples/ldap/ldap-config.yaml)
* [oauth/Grafana.yaml](../deploy/examples/oauth/Grafana.yaml): Installs Grafana and enable OAuth authentication using the OpenShift OAuthProxy. 
* [ha/Grafana.yaml](../deploy/examples/oauth/Grafana.yaml): Installs Grafana in high availability mode with Postgres as a database. 
* [persistentvolume/Grafana.yaml](../deploy/examples/persistentvolume/Grafana.yaml): Installs Grafana but provides a dedicated PVC for the database.
* [env/Grafana.yaml](../deploy/examples/env/Grafana.yaml): Shows how to provide env vars including admin credentials from a secret.

### Dashboards

* [SimpleDashboard.yaml](../deploy/examples/dashboards/SimpleDashboard.yaml): Minimal empty dashboard.
* [DashboardWithPlugins.yaml](../deploy/examples/dashboards/DashboardWithPlugins.yaml): Minimal empty dashboard with plugin dependencies.
* [DashboardFromURL.yaml](../deploy/examples/dashboards/DashboardFromURL.yaml): A dashboard that downloads its contents from a URL and falls back to embedded json if the URL cannot be resolved.
* [KeycloakDashboard.yaml](../deploy/examples/dashboards/KeycloakDashboard.yaml): A dashboard that shows keycloak metrics and demonstrates how to use datasource inputs.

### Data sources

* [Prometheus.yaml](../deploy/examples/datasources/Prometheus.yaml): Prometheus data source, expects a service named `prometheus-service` listening on port 9090 in the same namespace.
* [SimpleJson.yaml](../deploy/examples/datasources/SimpleJson.yaml): Simple JSON data source, requires the [grafana-simple-json-datasource](https://grafana.com/grafana/plugins/grafana-simple-json-datasource) plugin to be installed.
