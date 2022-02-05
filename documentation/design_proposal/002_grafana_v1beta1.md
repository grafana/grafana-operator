# Design proposal 002 grafana CRD v1beta1

## Summary

Update the grafana CRD to be easier to use and provide more customization opportunities.

## Info

status: Draft

## Motivation

The grafana CRD API is currently a bit sprawling and it can be hard to find how to
overwrite a number of default values, especially when it comes to deployment configuration.

As a part of the v1beta1 release we want to make this better and at the same time
give our users more power to customize there deployments.

## Verification

- Create integration tests for the new CRD:s
- Add new e2e tests to cover the new CR.

## Current

Currently the grafana spec looks like this.

```.go
type GrafanaSpec struct {
	Config                     GrafanaConfig            `json:"config"`
	Containers                 []v1.Container           `json:"containers,omitempty"`
	DashboardLabelSelector     []*metav1.LabelSelector  `json:"dashboardLabelSelector,omitempty"`
	Ingress                    *GrafanaIngress          `json:"ingress,omitempty"`
	InitResources              *v1.ResourceRequirements `json:"initResources,omitempty"`
	Secrets                    []string                 `json:"secrets,omitempty"`
	ConfigMaps                 []string                 `json:"configMaps,omitempty"`
	Service                    *GrafanaService          `json:"service,omitempty"`
	Deployment                 *GrafanaDeployment       `json:"deployment,omitempty"`
	Resources                  *v1.ResourceRequirements `json:"resources,omitempty"`
	ServiceAccount             *GrafanaServiceAccount   `json:"serviceAccount,omitempty"`
	Client                     *GrafanaClient           `json:"client,omitempty"`
	DashboardNamespaceSelector *metav1.LabelSelector    `json:"dashboardNamespaceSelector,omitempty"`
	DataStorage                *GrafanaDataStorage      `json:"dataStorage,omitempty"`
	Jsonnet                    *JsonnetConfig           `json:"jsonnet,omitempty"`
	BaseImage                  string                   `json:"baseImage,omitempty"`
	InitImage                  string                   `json:"initImage,omitempty"`
	LivenessProbeSpec          *LivenessProbeSpec       `json:"livenessProbeSpec,omitempty"`
	ReadinessProbeSpec         *ReadinessProbeSpec      `json:"readinessProbeSpec,omitempty"`
}
```

## Proposal

More and more users have asked us for more possibility to customize the grafana deployment but also
other objects that grafana-operator manages.

To be able to provide this I think that we should provide the entire object when possible while at the same time
be opinionated how the default object should look like just like we do today.

To get rid of required fields we could use a similar strategy as [Banzai Cloud](https://github.com/banzaicloud/operator-tools/blob/2189d8efc3856efd4a7c7fbb28b7cba9a977d0bd/pkg/typeoverride/override.go) does.

I know much of these things is already done in our repo. Would it be better to break out separately?
There have been problems with merging specific volumes and such inside the deployment and giving the option to override makes this complex.
We might just be happy with the current solution that we have and just make sure that all resources part of the object is defined in our code.

```.go
type GrafanaSpec struct {
	Config                     GrafanaConfig            `json:"config"`
	DashboardLabelSelector     []*metav1.LabelSelector  `json:"dashboardLabelSelector,omitempty"`
	Ingress                    *netv1.Ingress           `json:"ingress,omitempty"`
	RouteOpenShift             *v12.Route               `json:"routeOpenShift,omitempty"`
	Service                    *v1.Service              `json:"service,omitempty"`
	Deployment                 *typeoverride.Deployment `json:"deployment,omitempty"`
	ServiceAccount             *v1.ServiceAccount       `json:"serviceAccount,omitempty"`
	Client                     *GrafanaClient           `json:"client,omitempty"`
	DashboardNamespaceSelector *metav1.LabelSelector    `json:"dashboardNamespaceSelector,omitempty"`
	Jsonnet                    *JsonnetConfig           `json:"jsonnet,omitempty"`
	GrafanaContainer           *GrafanaContainer        `json:"grafanaContainer,omitempty"`
}

// GrafanaContainer provides a means to configure the grafana container
type GrafanaContainer struct {
	BaseImage         string                   `json:"baseImage,omitempty"`
	Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
	ReadinessProbe    *v1.Probe                `json:"readinessProbe,omitempty"`
	LivenessProbeSpec *v1.Probe                `json:"livenessProbe,omitempty"`
}
```

Instead of setting `grafana.spec.baseImage` you would have to do `grafana.spec.grafanaContainer.baseImage`.

Checkout the [grafana_types](002_grafana_types.go) file for a better example. Just overwrite `api/integreatly/v1alpha1/grafana_types.go` with this file.

## Alternatives

Instead of keeping things like baseImage we could provide the user with the entire deployment object with some tweaks.
This is for example how [Banzai Cloud](https://github.com/banzaicloud/operator-tools) does it.

```.go
type GrafanaSpec struct {
	Config                     GrafanaConfig            `json:"config"`
	DashboardLabelSelector     []*metav1.LabelSelector  `json:"dashboardLabelSelector,omitempty"`
	DeploymentOverrides        *typeoverride.Deployment `json:"deploymentOverrides,omitempty"`
	Ingress                    *GrafanaIngress          `json:"ingress,omitempty"`
	Service                    *GrafanaService          `json:"service,omitempty"`
	ServiceAccount             *GrafanaServiceAccount   `json:"serviceAccount,omitempty"`
	Client                     *GrafanaClient           `json:"client,omitempty"`
	DashboardNamespaceSelector *metav1.LabelSelector    `json:"dashboardNamespaceSelector,omitempty"`
	Jsonnet                    *JsonnetConfig           `json:"jsonnet,omitempty"`
}

// Deployment is a subset of [Deployment in k8s.io/api/apps/v1](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.19/#deployment-v1-apps), with [DeploymentSpec replaced by the local variant](#deployment-spec).
type Deployment struct {
	ObjectMeta `json:"metadata,omitempty"`
	// The desired behavior of [this deployment](#deploymentspec).
	Spec DeploymentSpec `json:"spec,omitempty"`
}
```

[DeploymentSpec](https://github.com/banzaicloud/operator-tools/blob/2189d8efc3856efd4a7c7fbb28b7cba9a977d0bd/pkg/typeoverride/override.go#L137-L176)

The **good** thing in designing the CRD this way is that you can do everything you want.

The **bad** thing in designing the CRD this way is that you can do everything you want.

It will be harder to do basic things like setting baseImages.

Instead of setting `grafana.spec.baseImage` you would have to do `grafana.spec.deploymentOverrides.DeploymentSpec.templates.containers.grafana.image`.

## Work Plan

Provide this PR with `grafana_types.go` containing all the type definitions.
Create a PR that implements this solution.

## Open questions

- How many of the templates should we provide in grafanaContainer? You can always override them in the deployment file

## Related issues

- [667 5.0 road map](https://github.com/grafana-operator/grafana-operator/pull/667)

## References

[Banzai Cloud operator-tools](https://github.com/banzaicloud/operator-tools)
