# Mounting extra files into the Grafana Pod

Sometimes Grafana needs to access extra configuration files. One example is LDAP authentication where the main configuration is stored in a separate file.

The Grafana operator supports mounting of Secrets and ConfigMaps into the Grafana Pod.

## Mounting Secrets and ConfigMaps

Consider the following config map containing LDAP configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ldap-config
data:
  ldap.toml: |
    [[servers]]
    host = "127.0.0.1"
    port = 389
    use_ssl = false
    start_tls = false
    ssl_skip_verify = false
    bind_dn = "cn=admin,dc=grafana,dc=org"
    bind_password = 'grafana'
    search_filter = "(cn=%s)"
    search_base_dns = ["dc=grafana,dc=org"]
    [servers.attributes]
    name = "givenName"
    surname = "sn"
    username = "cn"
    member_of = "memberOf"
    email =  "email"
```

The goal is to make `ldap.toml` available inside the Pod. Use the `spec.configMaps` property of the Grafana CR to automatically add a `Volume` and a `VolumeMount` to the Grafana Pod:

```yaml
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: example-grafana
spec:
  configMaps:
    - ldap-config
  secrets:
    ...
...
```

When mounting Secrets or ConfigMaps in Kubernetes, the keys become the file names and the values the contents of the file. For every config map specified in this way the Grafana operator will create a volume with the name `configmap-<name>` (the prefix will be `secret-` for Secrets) and add it to the Grafana deployment.
It will also create a volume mount with the same name and add it to all containers in the deployment. This includes the Grafana container and all extra containers specified via the `spec.containers` property. Config maps are mounted inside the containers under `/etc/grafana-configmaps/<configmap name>/`, secrets under `/etc/grafana-secrets/<secret name>/`.

The missing piece for the LDAP example is to tell Grafana about the location of the configuration file. This can be done in the config section of the Grafana CR:

```yaml
apiVersion: integreatly.org/v1alpha1
kind: Grafana
metadata:
  name: example-grafana
spec:
  configMaps:
    - ldap-config
  config:
    auth.ldap:
      enabled: true
      config_file: /etc/grafana-configmaps/ldap-config/ldap.toml
...
```

The full example can be found under `deploy/examples/ldap`.

*NOTE*: The Grafana Pod will not be able to start until all specified Secrets and Config maps exist.