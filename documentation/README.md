## Grafana Operator Documentation

* [Installing Grafana](./deploy_grafana.md)
* [Dashboards](./dashboards.md)
* [Data Sources](./datasources.md)
* [Multi namespace support](./multi_namespace_support.md)
* [Mounting extra config files](./extra_files.md)

## Examples

The following example CRs are provided:

### Grafana deployments

* [Grafana.yaml](../deploy/examples/Grafana.yaml): Installs Grafana using the default configuration and an Ingress if the `--openshift` flag is not provided. Suitable for Kubernetes and OpenShift.
* [GrafanaWithRoute.yaml](../deploy/examples/GrafanaWithRoute.yaml): Installs Grafana using the default configuration and a Route. Only suitable for OpenShift.  
* [GrafanaWithIngressHost.yaml](../deploy/examples/GrafanaWithIngressHost.yaml): Installs Grafana using the default configuration and an Ingress where the host is set for outside access. Only suitable for Kubernetes, setting the host on Ingresses is not permitted on OpenShift. 
* [ldap/Grafana.yaml](../deploy/examples/ldap/Grafana.yaml): Installs Grafana and sets up LDAP authentication. LDAP configuration is mounted from the configmap [ldap/ldap-config.yaml](../deploy/examples/ldap/ldap-config.yaml)

### Dashboards

* [SimpleDashboard.yaml](../deploy/examples/dashboards/SimpleDashboard.yaml): Minimal empty dashboard.
* [DashboardWithPlugins.yaml](../deploy/examples/dashboards/DashboardWithPlugins.yaml): Minimal empty dashboard with plugin dependencies.

### Data sources

* [Prometheus.yaml](../deploy/examples/datasources/Prometheus.yaml): Prometheus data source, expects a service named `prometheus-service` listening on port 9090 in the same namespace.
