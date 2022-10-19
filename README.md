# Grafana Operator

An operator to provision and manage Grafana Instances, Dashboards, Datasources and notification channels. Based on the [Operator-SDK](https://sdk.operatorframework.io/)

## Companies and teams that trust and use the Grafana operator

| Company  | Logo | Company | Logo
| :---     |    :----:   | :---        |    :----:   |
| [Red Hat](https://www.redhat.com)| <img src= "./media/users/redhat.png" width=100 height=100 > | [Integreatly](https://www.redhat.com/en/products/integration)|<img src= "./media/users/integreatly.png" width=100 height=100> |
 [Continental](https://www.continental.com/)|<img src= "./media/users/continental.png" width=100 height=100 > | [Handelsbanken]("https://www.handelsbanken.se/en/")|<img src= "./media/users/handelsbanken.svg" width=100 height=100 >|
| [Xenit](https://xenit.se/contact/)|<img src= "./media/users/xenit.png" width=150 height=50 >| [Torqata](https://torqata.com)|<img src= "./media/users/torqata.jpg" width=150 height=50 > |
|[Operate-first](https://www.operate-first.cloud/)|<img src= "./media/users/operate-first.png" width=100 height=100 > | [iFood](https://www.ifood.com.br)|<img src= "./media/users/ifood.svg" width=150 height=50 > |

***If you find this operator useful in your product/deployment, feel free to send a pull request to add your company/team to be displayed here!***

## Grafana Operator on the Kubernetes community Slack

We have set up a channel dedicated to this operator on the Kubernetes community Slack, this is an easier way to address
more immediate issues and facilitate discussion around development/bugs etc. as well as providing support for questions
about the operator.

1: Join the Kubernetes Slack (if you have not done so already) [Kubernetes Slack](https://slack.k8s.io/).

2: You will receive an email with an invitation link, follow that link and enter your desired username and password for the workspace(it might be easier if you use your Github username for our channel).

3: Once registered and able to see the Kubernetes community Slack workspace and channels follow this link to the [grafana-operator channel](https://kubernetes.slack.com/messages/grafana-operator/ ).

Alternatively:
If you're already a member of that workspace then just follow this link to the [grafana-operator channel](https://kubernetes.slack.com/messages/grafana-operator/)
or search for "grafana-operator" in the browse channels option.

![image](https://user-images.githubusercontent.com/35736504/90978105-0b195300-e543-11ea-86ee-1825da0e3b75.png)

## Current status

All releases of the operator can be found on [Operator Hub](https://operatorhub.io/operator/grafana-operator).

***Sometimes a release may take a few hours (in rare cases, days) to land on Operatorhub, please be patient, it's out of our control.***

### Supported Versions

#### v3.x

***This version has known vulnerabilities present, rooted in the version of the operator-sdk that was used, please upgrade to v4(operator-sdk v1.3.0) to mitigate the risk***

This version of the operator will be deprecated in the near future, we recommend new users to install v4 and existing users to upgrade as soon as possible using the [upgrade guide](./documentation/upgrade.md).

We won't be accepting any new features for v3, the only releases made under this version will be either bug-fixes or security patches.

The operator-sdk is an exception to the security patch rule, it cannot be updated without introducing breaking changes, hence the recommendation to upgrade to v4, which mitigates these CVEs.

The documentation for this version can be found here: [https://github.com/grafana-operator/grafana-operator/tree/v3/documentation](https://github.com/grafana-operator/grafana-operator/tree/v3/documentation).

#### v4.x (master)

This is the current main branch of the project, all future development will take place here, any new features and improvements should be submitted against this branch.

Please use the following link to access documentation at any given release of the operator:

```txt
https://github.com/grafana-operator/grafana-operator/tree/<version>/documentation
```

## Summary of benefits

Why decide to go with the Grafana-operator over a standard standalone Grafana deployment for your monitoring stack?

If [the benefits of using an operator over standalone products as outlined by the people that created them](https://operatorframework.io/) and our current high-profile users aren't enough to convince you, here's some more:

* The ability to configure and manage your entire Grafana with the use Kubernetes resources such as CRDs, configMaps, Secrets etc.
* Automation of:
  * Ingresses.
  * Grafana product versions.
  * Grafana dashboard plugins.
  * Grafana datasources.
  * Grafana notification channel provisioning.
  * Oauth proxy.
  * many others!
* Efficient dashboard management through jsonnet, plugins, organizations and folder assignment, which can all be done through `.yamls`!
* Both Kubernetes and OpenShift supported out of the box.
* Multi-Arch builds and container images.
* Operatorhub/OLM support (Allows you to install the operator with a few clicks).

And the things on our roadmap:

* Multi-Namespace and Multi-Instance support, allowing the operator to manage not only your Grafana instance, but also any other grafana instance on the cluster, eg. for public facing customer instance.

## Operator flags

The operator supports the following flags on startup.
See [the documentation](./documentation/deploy_grafana.md) for a full list.
Flags can be passed as `args` to the container.

## Supported Custom Resources

The following Grafana resources are supported:

* Grafana
* GrafanaDashboard
* GrafanaDatasource
* GrafanaNotificationChannel

all custom resources use the api group `integreatly.org` and version `v1alpha1`.
To get an overview of the available grafana-operator CRD see api.md.

### Grafanas

Represents a Grafana instance. See [the documentation](./documentation/deploy_grafana.md) for a description of properties supported in the spec.

### Dashboards

Represents a Grafana dashboard and allows specifying required plugins. See [the documentation](./documentation/dashboards.md) for a description of properties supported in the spec.

### Datasources

Represents a Grafana datasource. See [the documentation](./documentation/datasources.md) for a description of properties supported in the spec.

### Notifiers

Represents a Grafana notifier. See [the documentation](./documentation/notifiers.md) for a description of properties supported in the spec.

## Development and Local Deployment

### Using the Makefile

If you want to develop/build/test the operator, here are some instructions how to set up your dev-environment: [follow me](./documentation/develop.md)

## Debug

We have documented a few steps to help you debug the [grafana-operator](documentation/debug.md).
