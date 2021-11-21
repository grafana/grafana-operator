# Design proposal 001 organization managment

## Summary

Add one or multiple CRD to the grafana-opreator that enables developers to manage:

- Organizations
- Teams
- Users

## Info

status: Draft

## Motivation

Many big organiaztions are using grafana-opertor to manage there grafana instance and they want to be able to manage organiaztions inside grafana.
There is also a need to be able to create local user account to be able to run your grafana on a TV.

## Goals

Manage organizations, teams and users using CRD:s

## Verification

- Create integration tests for the new CRD:s
- Add new e2e tests to cover the new CR.

## Proposal

### Pros

- Users can be added to the organization in the CR.
- Rename of a organization is doable
- Easier to use since it's clear that a team/user is under a specific organiaztion.

### Cons

- Can't reuse users/teams in multiple organiaztions with a simple labelSelector.

### CRD

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaOrganization # Not a good name, need a better one.
metadata:
  name: organization-example
spec:
  orgniaztions:
    - name: "Main Org."
      grafanaLabelSelector: # Requiered
      dashboardNamespaceSelector: # Will assume same ns unless specefied
      users:
        - loginOrEmail: user1 # I don't think we need to define user1 in this org since it's users1 default org, should it even be allowed
          role: admin # If the user is defined it have to match the same setting that is defined in user1.
        - loginOrEmail: user2
          role: Viewer
      teams:
        - name: MyTestTeam
          email: email@test.com
          theme: dark
          homeDashboardId: 39
          timezone: utc
          memebers:
            - user1
            - user2
    - name: "MegaOrg"
      grafanaLabelSelector:
      dashboardNamespaceSelector:
      users:
        - loginOrEmail: user3
          role: Viewer
  users:
    - login: user1 # Requiered
      email: user1@github.com
      role: admin # Requiered
      name: user1
      theme: light
      isGrafanaAdmin: true
      isDisabled: false
      orgName: "Main Org." # Requiered or we will assume that it's Main Org. by default
      password: # Requiered
        secretName: usersecret
        key: password
    - login: user2
      email: noreply2@grafana.com
      role: Viewer
      orgName: "Main Org."
      password:
        secretName: user2secret
        key: password
    - login: user3
      email: noreply3@grafana.com
      role: Viewer
      orgName: "Main Org."
      password:
        secretName: user3secret
        key: password
```

## Context

We need to consider the [multi-namespace support](https://github.com/grafana-operator/grafana-operator/pull/599) that is currently getting worked on.
The operator will start to manage multiple grafana instances in a single cluster.

In the future it might be intresting to [manage external grafana instances](https://github.com/grafana-operator/grafana-operator/issues/402) even though it's currently not being prioratized.
We haven't had any discussion about this but that might also be off the table considering [Avoiding scope creep](https://github.com/grafana-operator/grafana-operator/issues/497).

When creating the new controller we should use [CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362) from the begining.

We woulden't consider multiple replicas of the grafana instance, to be sure that the grafana CRD does what it should the users will have to use something like postgres
to store state to be able to have multiple grafana instances. For example this is the same way when it comes to LDAP.
We need to document the assumptions that we are making.

We would have to add retry logic to readd organiaztion/teams/users, just like we did recently in the [dashboard controller](https://github.com/grafana-operator/grafana-operator/pull/488).
Another way could be to use [status.condition](https://github.com/grafana-operator/grafana-operator/issues/571) and see if the grafana instance got a status.condition saying that the organiaztion exists.
The bad thing with just relying on that would be that it woulden't recreate if someone deleted the organiaztion manually inside grafana.
The status.condition would still be usefull to implement as well.

## Alternatives

Instead of having a single CRD to manage organiaztions, teams and users we could have multiple ones.

If teams, users or orgs get more "abilities" they are separated and can be changed independently.

Also if upstream grafana decides to do a breaking change around one the CR:s we only need to bump one of the API versions,
assuming that they are not to hard coupled.

We could also reuse the same team/user definition in multiple organizations.

Thanks to all the extra config that is needed it will be easier to do an error.

### Thoughts

How should we find the organiaztion where we should create the user/team? Just define a name or use a selector?

### Multiple CRD

```org.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaOrganization
metadata:
  name: org1
spec:
  grafanaLabelSelector:
  dashboardNamespaceSelector:
  - name: organiaztionX
```

```team.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaTeam
metadata:
  name: team1
spec:
  orgName: "Main Org."
  grafanaLabelSelector:
  dashboardNamespaceSelector:
  teams:
    - name: MyTestTeam
      email: email@test.com
      theme: dark
      homeDashboardId: 39
      timezone: utc
      organization: organiaztionX
      memebers: # Could use labelSelectors here instead.
        - user1
        - user2
```

```user.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaUser
metadata:
  name: user1
spec:
  orgName: "Main Org."
  grafanaLabelSelector:
  dashboardNamespaceSelector:
  - login: user1
    email: user1@github.com
    role: admin
    name: user1
    theme: light
    isGrafanaAdmin: true
    isDisabled: false
    password:
      secretName: usersecret
      key: password
  - login: user2
    email: noreply2@grafana.com
    role: Viewer
    password:
      secretName: user2secret
      key: password
  - login: user3
    email: noreply3@grafana.com
    role: Viewer
    password:
      secretName: user3secret
      key: password
```

## Work Plan

Implement the new CR, add each API one by one.

- Organiaztion
- User
- Team

## Open questions

- We need to come up with a better name for kind: GrafanaOrganization.
- The default organiaztion is called `Main Org.` I think most of our users will use this organiaztion. How should we help them use it,
  it's a rather strange name and in the current design they will have to define that name to use it.
- Should we define the user under a organiaztion if it's already defined as it's default organiaztion when creating the user?
- How do we manage each of these new CRs? would it be through a higher level operator that manages organization related resources? or would it be a controller per resource?

## Related issues

- [408 Manage organizations](https://github.com/grafana-operator/grafana-operator/issues/408)
- [525 Support for orgId in GrafanaDashboard](https://github.com/grafana-operator/grafana-operator/issues/525)
- [174 Cannot create multiple Grafana in the same namespace](https://github.com/grafana-operator/grafana-operator/issues/174)
- [362 Refactor controllers to use controllerruntime CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362)
- [571 Add status.condition to grafandashboards and grafanadatsources](https://github.com/grafana-operator/grafana-operator/issues/571)
- [432 Grafanadashboards aren't recreated if deleted through the Grafana UI](https://github.com/grafana-operator/grafana-operator/issues/432)

## References

- [Grafana organiaztion API](https://grafana.com/docs/grafana/latest/http_api/org/)
- [Grafana team API](https://grafana.com/docs/grafana/latest/http_api/team/)
- [Grafana user API](https://grafana.com/docs/grafana/latest/http_api/user/)
- [Grafana config auto_assign_org](https://grafana.com/docs/grafana/latest/administration/configuration/#auto_assign_org)
