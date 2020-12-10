# Grafana Operator

A Kubernetes Operator based on the Operator SDK for creating and managing Grafana instances.

## Grafana Operator on the Kubernetes community Slack
We have set up a channel dedicated to this operator on the Kubernetes community Slack, this is an easier way to address 
more immediate issues and facilitate discussion around development/bugs etc. as well as providing support for questions
about the operator.
    
1: Join the Kubernetes Slack (if you have not done so already) [Kubernetes Slack](https://slack.k8s.io/).
    
2: You will receive an email with an invitation link, follow that link and enter your desired username and password for the workspace(it might be easier if you use your Github username for our channel).
    
3: Once registered and able to see the Kubernetes community Slack workspace and channels follow this link to the [grafana-operator channel](https://kubernetes.slack.com/messages/grafana-operator/ ).
    
Alternatively:
If you're already a member of that workspace then just follow this link to the [grafana-operator channel](https://kubernetes.slack.com/messages/grafana-operator/ ) or search for "grafana-operator" in the browse channels option .

![image](https://user-images.githubusercontent.com/35736504/90978105-0b195300-e543-11ea-86ee-1825da0e3b75.png)
    


# Current status

The Operator is available on [Operator Hub](https://operatorhub.io/operator/grafana-operator).

It can deploy and manage a Grafana instance on Kubernetes and OpenShift. The following features are supported:

* Install Grafana to a namespace
* Import Grafana dashboards from the same or other namespaces
* Assign Grafana dashboards to custom or namespace-named grafana folders
* Import Grafana datasources from the same namespace
* Install Plugins (panels) defined as dependencies of dashboards 

# Summary of benefits
A list of benefits of using the operator over a single grafana deployment.

[The benefits of using an operator over standalone products as outlined by the people that created them](https://operatorframework.io/) 

Grafana-operator specific benefits over standalone Grafana:

* The ability to configure and manage grafana deployments using kubernetes resources such as CRDs,configMaps etc
* Incoming changes to allow for multi-namespace support for the operator, meaning that just one instance of the operator can manage every instance of grafana on the cluster!
* Reducing the need for manual steps, allowing for quicker setups for things such as (and not limited to):
    * Ingresses
    * Grafana product versions
    * Grafana dashboard plugins
    * many others!
* Manage dashboards more efficiently through jsonnet, plugins and folder assignment, which can all be done through .yamls!
* Periodical reconciliation of resources, ensuring that the desired state is maintained, so nothing will be broken for too long




# Operator flags

The operator supports the following flags on startup.
See [the documentation](./documentation/deploy_grafana.md) for a full list.
Flags can be passed as `args` to the container.

# Supported Custom Resources

The following Grafana resources are supported:

* Grafana
* GrafanaDashboard
* GrafanaDatasource

all custom resources use the api group `integreatly.org` and version `v1alpha1`.

## Grafana

Represents a Grafana instance. See [the documentation](./documentation/deploy_grafana.md) for a description of properties supported in the spec.

## GrafanaDashboard

Represents a Grafana dashboard and allows specifying required plugins. See [the documentation](./documentation/dashboards.md) for a description of properties supported in the spec.

## GrafanaDatasource

Represents a Grafana datasource. See [the documentation](./documentation/datasources.md) for a description of properties supported in the spec.

# Building the operator image

Init the submodules first to obtain grafonnet:

```sh
$ git submodule update --init
```

Then build the image using the operator-sdk:

```sh
$ operator-sdk build <registry>/<user>/grafana-operator:<tag>
```

# Running locally

You can run the Operator locally against a remote namespace using the operator-sdk:

Prerequisites:

* [operator-sdk](https://github.com/operator-framework/operator-sdk) installed
* kubectl pointing to the local context. [minikube](https://github.com/kubernetes/minikube) automatically sets the context to the local VM. If not you can use `kubectl config use <context>` or (if using the OpenShift CLI) `oc login -u <user> <url>`
* make sure to deploy the custom resource definition using the command ```kubectl create -f deploy/crds```

```sh
$ operator-sdk run local --namespace=<namespace> --operator-flags="<flags to pass>"
```

# Grafana features not yet supported in the operator

## Notifier provisioning

Grafana has provisioning support for multiple channels (notifiers) of alerts. The operator does currently not support this type of provisioning. An empty directory is mounted at the expected location to prevent a warning in the grafana log. This feature might be supported in the future. 
