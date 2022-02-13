# Design proposal 002 grafana CRD v1beta1

## Summary

Update the grafana CRD to be easier to use and provide more customization opportunities.

This document only contain design changes for the grafana crd, not grafanadashboard, grafanasource, etc.

The changes in this document contain allot of breaking changes.

## Info

status: Draft

## Motivation

More and more users have asked us for more possibility to customize the grafana deployment but also
other objects that grafana-operator manages.

The new version of the CRD should provide the entire deployment/containers object when possible while at the same time
be opinionated how the default object that we generate should look like just like even more then we do today.

The grafana CRD API is currently a bit sprawling and it can be hard to find how to
overwrite a number of default values, especially when it comes to deployment configuration.

As a part of the v1beta1 release we want to make this better and at the same time
give our users more power to customize there deployments.

## Out of scope

In the roadmap for 5.0 there are mentions about moving `dashboardLabelSelector` and `dashboardNamespaceSelector` from the grafana object
and instead have the grafanadashboard crd to find the grafana it's selected instances.
This design document don't tale this in to consideration and only focuses on configuration of the grafana instance.

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

In short the proposal is about moving the grafana container specific config from the main GrafanaSpec to GrafanaContainer.
GrafanaContainer should contain [V1.Container](https://pkg.go.dev/k8s.io/api@v0.20.2/core/v1?utm_source=gopls#Container) but with removal of a number of required values like name that will be hard coded to `grafana`.

This way we could provide a shortcut to our users to be able to set all grafana related config inside one part of the crd.

Create specific structs for all the objects that we manage in the operator like deployment, configmaps, serviceaccounts, etc.
and make them as generic as possible to give our users the opportunity to tweak the deployments in any way they want.

To minimize the amount of fields needed to be defined when overwriting custom config we should get rid of required fields for many objects. We could use a similar strategy as [Banzai Cloud](https://github.com/banzaicloud/operator-tools/blob/2189d8efc3856efd4a7c7fbb28b7cba9a977d0bd/pkg/typeoverride/override.go) does.
This is something that we already do in many cases in our repository, the question is if it would be good to use a third party package to help out with this to minimize the amount of code in our repo.
The bad thing is of course the lower the amount of control we get over it. But we can always fork repos or setup a similar solution on our own.

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
```

For example instead of setting `grafana.spec.baseImage` you would have to do `grafana.spec.grafanaContainer.image`.

### Change in defaults

Today we have a number of defaults that is a bit strange. It would be good change a number of defaults, both on the
operator it self and the grafana spec to make it as secure as possible out of the box.

#### Default Security Context

Set better default security contexts on deployments and containers created by the operator.

Openshift got a special way of managing security context called [SCC](https://docs.openshift.com/container-platform/4.9/authentication/managing-security-context-constraints.html)
and we have to add some check to see if we are deploying in a Openshift environment or not.

The securityContext that we should aspire to is something like this:
Openshift don't support providing `runAsGroup` and `runAsUser` since it's automatically set per namespace and the numbers provided is random.
`runAsNonRoot` might work since it gets user and group from Openshift by default.

> These settings should also apply to our plugin initContainer.

```.yaml
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        drop:
        - NET_RAW
      privileged: false
      readOnlyRootFilesystem: true
      runAsGroup: 1000
      runAsNonRoot: true
      runAsUser: 1000
```

These settings also forces us to add some default volumes else grafana won't start. TODO add the volumes needed.

#### Default routes/ingress

Today it's not possible to define none https route for Openshift users. By separating routes and ingress in to separate configs in the crd we should be able to separate the defaults configs for each object easier. It would also become easier to provide the entire object to the users so they can set any config they want.

In Openshift a good default is TLS enabled by default since it's built in to Openshift.
While in ingress it's not as easy since we don't know how certificates is managed.

So what needs to change?

The default setting for routes should be [edge](https://docs.openshift.com/container-platform/4.9/networking/routes/secured-routes.html) `TLS enabled=true` but enable the possibility to run in http mode only.

The default ingress should keep on being TLS enabled=false

#### Default resource requests

To minimize the risk of our operator and the grafana deployment to be evicted during high load in your clusters we should provide basic resource requests for both the operator and the grafana deployment/init container.

These settings will be rather low by default.

## Alternatives1

Just like the proposal but instead of giving out the entire container object we could keep the custom resources that we have today.

```.go
// GrafanaContainer provides a means to configure the grafana container
type GrafanaContainer struct {
 BaseImage         string                   `json:"baseImage,omitempty"`
 Resources         *v1.ResourceRequirements `json:"resources,omitempty"`
 ReadinessProbe    *v1.Probe                `json:"readinessProbe,omitempty"`
 LivenessProbeSpec *v1.Probe                `json:"livenessProbe,omitempty"`
}
```

This would make it easier for our current users when migrating to the new version since they would know how it was written in 4.0.

But this would also create confusion how can i configure other grafana container related config?

You would set baseImage with `grafana.spec.grafanaContainer.baseImage` but if you want to change the securityContext you would have to change in
`grafana.spec.deploymentOverrides.DeploymentSpec.templates.containers.grafana.securityContext`. This is rather confusing since they perform changes in the same object.

This would also make it harder from a code point of view since we would have to write some special merge logic
instead of just checking if the user tries to do any changes to the grafana container.
This would increase the risk of creating merge issues

## Alternatives2

Just like the proposal but instead of having a custom object for grafanaContainer we could have the user
configure all deployment related settings through the `DeploymentOverrides` object.

This would make it harder to use but probably allot easier from a development point of view since we would have to do less advanced merges between configs.

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

It will be harder to do basic things like setting baseImages.

Instead of setting `grafana.spec.baseImage` you would have to do `grafana.spec.deploymentOverrides.DeploymentSpec.templates.containers.grafana.image`.

## Open questions

- ~~Should we do a similar solution for the init pluigin container like the proposal is for the grafana container?~~
  - There are very few people that have asked us for changes in the initContainer in general. It's probably not needed to have a easy access config for the init container.
- Banzai clouds operator-tools don't provide Openshift routes. Don't know how keen they would be to change this upstream. We probably want to keep this config in a repo that we own to make it easier to change if we need to.
- What securityContext settings can we use in Openshift without any issues?

## Related issues

- [667 5.0 road map](https://github.com/grafana-operator/grafana-operator/pull/667)
- [649 Route TLS=false](https://github.com/grafana-operator/grafana-operator/issues/649)

## References

- [Banzai Cloud operator-tools](https://github.com/banzaicloud/operator-tools)
- [Openshift scc](https://docs.openshift.com/container-platform/4.9/authentication/managing-security-context-constraints.html)
- [edge TLS Openshift route](https://docs.openshift.com/container-platform/4.9/networking/routes/secured-routes.html)
