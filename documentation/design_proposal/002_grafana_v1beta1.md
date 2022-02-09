# Design proposal 002 grafana CRD v1beta1

## Summary

Update the grafana CRD to be easier to use and provide more customization opportunities.

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

To get rid of required fields we could use a similar strategy as [Banzai Cloud](https://github.com/banzaicloud/operator-tools/blob/2189d8efc3856efd4a7c7fbb28b7cba9a977d0bd/pkg/typeoverride/override.go) does.

I know much of these things is already done in our repo.
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

### Change in defaults

#### Default Security Context

Today we have a number of defaults that would be good to apply to the operator it self and the grafana deployment
to make it as secure as possible out of the box. One of the problems with this in openshift [SCC](https://docs.openshift.com/container-platform/4.9/authentication/managing-security-context-constraints.html)
and we have to add some check to see if we are deploying in a openshift environment or not.

One of the configs we should asspire to is something like this in our container seucirty context.
Openshift don't support providing `runAsGroup` and `runAsUser`, `runAsNonRoot`might work since it gets user and group from openshift by default.

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

Today it's not possible to define none https route. By seperating routes and ingress in to seperate configs in the crd we should be able to seperate these defaults easier.
In openshift a good default is TLS enabled by default. While in ingress it's not as easy.

So what needs to change?

The default setting for routes should be [edge](https://docs.openshift.com/container-platform/4.9/networking/routes/secured-routes.html) `TLS enabled=true` but enable the possability to run in http mode only.

The default ingress should keep on being TLS enabled=false

#### Default resource requests

To miniamize the risk of our operator and the grafana deployment to be evicted during high load in your clusters we should provide basic resource reuqests for both the operator and the grafana deployment/init container.

## Alternatives

Instead of having a custom object for grafanaContainer we could provide the user with the entire deployment object with some tweaks and have them configure everything from inside the deployment.

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

## Work Plan

Provide this PR with `grafana_types.go` containing all the type definitions.
Create a PR that implements this solution.

## Open questions

- Should we do a similar solution for the init pluigin container?
- Banzai clouds operator-tools don't support openshift routes. Don't know how keen they would be to change this upstream. We probably want to keep this config in a repo that we own to make it easier to change if we need to.
- What securityContext settings can we use in openshift without any issues?
- If we think that the proposal is the best way forward how do we make sure that the developers only configure the grafana container in one location? Do we need to make sure of this? Can we manage this through some intresting merge logic?

## Related issues

- [667 5.0 road map](https://github.com/grafana-operator/grafana-operator/pull/667)
- [649 Route TLS=false](https://github.com/grafana-operator/grafana-operator/issues/649)

## References

- [Banzai Cloud operator-tools](https://github.com/banzaicloud/operator-tools)
- [openshift scc](https://docs.openshift.com/container-platform/4.9/authentication/managing-security-context-constraints.html)
- [edge TLS openshift route](https://docs.openshift.com/container-platform/4.9/networking/routes/secured-routes.html)
