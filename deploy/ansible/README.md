# Grafana Operator | Ansible

Ansible scripts for deploying and configuring the Grafana Operator.

# Playbooks

## grafana-operator-cluster-resources.yaml

Playbook for deploying the cluster level resources required for deploying the Grafana Operator.

### Required K8s Permisions 
These permissions are reuqied by the K8s user who runs this playbook. Typically this would be a cluster administrator.

* create CRDs
* create ClusterRoles

### Parameters

| Parameter                                        | Choices / **Defaults** | Comments
|--------------------------------------------------|------------------------|---------
| k8s\_host                                        |                        | K8s API to run this playbook against
| k8s\_validate\_certs                             | **True** / False       | Whether to validate K8s API certificate
| k8s\_api\_key                                  |                        | K8s API token to authenticate with. Mutually exclusive with `k8s_username` and `k8s_password`.
| k8s\_username                                    |                        | K8s username to authenticate with. Mutually exclusive with `k8s_api_key`.
| k8s\_password                                    |                        | K8s password to authenticate with. Mutually exclusive with `k8s_api_key`.
| grafana\_operator\_install                       | **True** / False       | `True` to install the resources, `False` to uninstall
| grafana\_operator\_allow\_project\_admin\_deploy | **True** / False       | `True` to install the aggregated ClusterRoles to allow project admins to be able to deploy Grafana Operator and associated Grafana custom resources. Ignored if `grafana_operator_install` is `False`.

### Examples

#### Install cluster resources
```
ansible-playbook grafana-operator-cluster-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret
```

#### Uninstall cluster resources
```
ansible-playbook grafana-operator-cluster-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e grafana_operator_install=false
```

## grafana-operator-namespace-resources.yaml

Creates the Grafana Operator in a given namespace.

This playbook will only work if a cluster administrator has first run the [grafana-operator-cluster-resources.yaml](#grafana-operator-cluster-resourcesyaml) playbook.

### Required K8s Permisions 
These permissions are reuqied by the K8s user who runs this playbook.

* create Namespace/Project
* list Namespaces/Projects
* admin of destination Namespace/Project
* [Grafana custom resources CRUD](../cluster_roles/cluster_role_aggregate_grafana_admin_edit.yaml)

### Parameters

| Parameter                                                         | Choices / **Defaults** | Comments
|-------------------------------------------------------------------|------------------------|---------
| k8s\_host                                                         |                        | K8s API to run this playbook against
| k8s\_validate\_certs                                              | **True** / False       | Whether to validate K8s API certificate
| k8s\_api\_key                                                   |                        | K8s API token to authenticate with. Mutually exclusive with `k8s_username` and `k8s_password`.
| k8s\_username                                       `             |                        | K8s username to authenticate with. Mutually exclusive with `k8s_api_key`.
| k8s\_password                                                     |                        | K8s password to authenticate with. Mutually exclusive with `k8s_api_key`.
| grafana\_operator\_install                                        | **True** / False       | `True` to install the resources, `False` to uninstall
| grafana\_operator\_namespace                                      | **grafana**	     | Namespace to create the Grafana Operator in.
| grafana\_operator\_delete\_namespace                              | True / **False**       | If `grafana_operator_install` is `False` and this is `True` then the Namespace specified in `grafana_operator_namespace` will be deleted. Else if this is `False` only the Grafana Operator resources in `grafana_operator_namespace` will be deleted. Ignored if `grafana_operator_install` is `True`.
| grafana\_operator\_args\_grafana\_image                           |                        | Value to pass to `--grafana-image` Grafana Operator flag.
| grafana\_operator\_args\_grafana\_image\_tag                      |                        | Value to pass to `--grafana-image-tag` Grafana Operator flag.
| grafana\_operator\_args\_grafana\_plugins\_init\_container\_image |                        | Value to pass to `--grafana-plugins-init-container-image` Grafana Operator flag.
| grafana\_operator\_args\_grafana\_plugins\_init\_container\_tag   |                        | Value to pass to `--grafana-plugins-init-container-tag` Grafana Operator flag.
| grafana\_operator\_args\_scan\_all                                | True / **False**       | `True` to pass `--scan-all` Grafana Operator flag. If `True` is passed then a cluster administrator must also run [grafana-operator-cluster-dashboards-scan.yaml](#grafana-operator-cluster-dashboards-scan.yaml) to grant the required permissions for the Grafana Operator to scan other namespaces for GrafanaDashboards custom resources. Mutually exclusive with `grafan_operator_args_namespaces`.
| grafana\_operator\_args\_namespaces                               |                        | Comma speerated list of namespaces for the Grafana Operator to scan for GrafanaDashboards custom resources. For this to work the ServiceAccount created by this playbook must be manually given permissions to read GrafanaDashboard resources in the listed namespaces or a cluster administrator must also run [grafana-operator-cluster-dashboards-scan.yaml](#grafana-operator-cluster-dashboards-scan.yaml) to grant the required permissions for the Grafana Operator to scan any namespace for GrafanaDashboard resources. Mutually exlusive with `grafana_operator_args_scan_all`.

### Examples

#### Install namespace resources
```
ansible-playbook grafana-operator-namespace-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=project_creator \
  -e k8s_password=secret
```

#### Install namespace resources - scan all namespaces for GrafanaDashboards
```
ansible-playbook grafana-operator-namespace-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=project_creator \
  -e k8s_password=secret \
  -e grafana_operator_args_scan_all=true
```

#### Uninstall namespace resources - keep namespace
```
ansible-playbook grafana-operator-namespace-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=project_creator \
  -e k8s_password=secret \
  -e grafan_operator_install=False
```

#### Uninstall namespace resources - delete namespace
```
ansible-playbook grafana-operator-namespace-resources.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=project_creator \
  -e k8s_password=secret \
  -e grafan_operator_install=False \
  -e grafana_operator_delete_namespace=True
```

## grafana-operator-cluster-dashboards-scan.yaml

Playbook for granting a Grafana Operator permissions to scan all cluster namespaces for GlusterDashboards.

### Required K8s Permisions 
These permissions are reuqied by the K8s user who runs this playbook. Typically this would be a cluster administrator.

* create ClusterRoles
* create ClusterRoleBindings

### Parameters

| Parameter                    | Choices / **Defaults** | Comments
|------------------------------|------------------------|---------
| k8s\_host                    |                        | K8s API to run this playbook against
| k8s\_validate\_certs         | **True** / False       | Whether to validate K8s API certificate
| k8s\_api\_key              |                        | K8s API token to authenticate with. Mutually exclusive with `k8s_username` and `k8s_password`.
| k8s\_username                |                        | K8s username to authenticate with. Mutually exclusive with `k8s_api_key`.
| k8s\_password                |                        | K8s password to authenticate with. Mutually exclusive with `k8s_api_key`.
| grafana\_operator\_install   | **True** / False       | `True` to grant the permissions. `False` to revoke.
| grafana\_operator\_namespace | **grafana**            | Namespace the `grafana-operator` ServiceAccount is in to grant cluster wide read to GrafanaDashboad resources.

### Examples

#### Grant permissions
```
ansible-playbook grafana-operator-cluster-dashboards-scan.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret
```

#### Grant permissions - custom namespace
```
ansible-playbook grafana-operator-cluster-dashboards-scan.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e grafan_operator_namespace=custom-monitoring
```

#### Revoke permissoins
```
ansible-playbook grafana-operator-cluster-dashboards-scan.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e grafana_operator_install=false
```

#### Revoke permissoins - custom namespace
```
ansible-playbook grafana-operator-cluster-dashboards-scan.yaml \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e grafan_operator_namespace=custom-monitoring \
  -e grafana_operator_install=false
```

## openshift-monitoring-update-prometheus-authentication.yaml

Playbook for updateing the authentication for prometehus so as to create accounts for other dashbaords to consume the data prometheus data through the proxy.

### Required K8s Permisions
These permissions are reuqied by the K8s user who runs this playbook. Typically this would be a cluster administrator.

* edit access to the `openshift-monitoring` project

### Parameters

| Parameter                   | Choices / **Defaults** | Comments
|-----------------------------|------------------------|---------
| k8s\_host                   |                        | K8s API to run this playbook against
| k8s\_validate\_certs        | **True** / False       | Whether to validate K8s API certificate
| k8s\_api\_key               |                        | K8s API token to authenticate with. Mutually exclusive with `k8s_username` and `k8s_password`.
| k8s\_username               |                        | K8s username to authenticate with. Mutually exclusive with `k8s_api_key`.
| k8s\_password               |                        | K8s password to authenticate with. Mutually exclusive with `k8s_api_key`.
| prometheus\_htpasswd\_users |                        | Array of hashes of user names and passwords to add or remove from the prometheus authentication list. Expected hash keys are `name`, `password`, and optionally `state` with value of `present` or `absent` with a default of `present`.

#### prometheus\_htpasswd\_users
YAML example
```yaml
prometheus_htpasswd_users:
  - name: custom-monitoring
    password: secret
  - name: custom-monitoring-old
    password: secret
    state: absent
```

JSON example
```json
'[{"name":"custom-monitoring","password":"secret"},{"name":"custom-monitoring-old","password":"secret","state":"absent"}]'
```

### Examples
These are all "one lineer" examples. If following "true" "git-ops" principles recomended you make an inventory file and group\_vars or host\_vars file with the parameters specified there.

### Add new prometheus user
```
ansible-playbook deploy/ansible/openshift-monitoring-expose-prometheus.yaml \
  -i deploy/ansible/localhost.ini \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e prometheus_htpasswd_users='[{"name":"custom-monitoring","password":"secret"}]'
```

### Remove prometheus user
```
ansible-playbook deploy/ansible/openshift-monitoring-expose-prometheus.yaml \
  -i deploy/ansible/localhost.ini \
  -e k8s_host=https://ocp.example.xyz \
  -e k8s_username=admin1 \
  -e k8s_password=secret \
  -e prometheus_htpasswd_users='[{"name":"custom-monitoring","password":"secret","state":"absent"}]'
```

# Tested With
These are the versions these playbooks have been tested with. This does not mean this wont work with other versions this is simply known working versions.

* Ansible
  * 2.9.2
* OpenShift Container Platform
  * v3.11.154 
