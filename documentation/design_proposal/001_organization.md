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

### New CRD prod

- Users can be added to the organization in the CR.
- Rename of a organization is doable
- IF the organization gets new functionality it could be added easier.

### New CRD CONS

- Needs to be applied after the deployment is done.
- API-url needs to be provided to the CR.

### CRD

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: grafana1
spec:
  - name: organiaztionX
      users:
        - login: user1
          email: user1@github.com
          role: admin
          name: user1
          theme: light
          isGrafanaAdmin: true
          isDisabled: true
          password:
            secretName: mysecret
            key: password
        - login: user2
          email: noreply@grafana.com
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
```

## Context

We need to consider the [multi-namespace support](https://github.com/grafana-operator/grafana-operator/pull/599) that is currently getting worked on.
The operator will start to manage multiple grafana instances in a single cluster.

In the future it might be intresting to [manage external grafana instances](https://github.com/grafana-operator/grafana-operator/issues/402) even though it's currently not being prioratized.
We haven't had any discussion about this but that might also be off the table considering [Avoiding scope creep](https://github.com/grafana-operator/grafana-operator/issues/497).

When creating the new controller we should use [CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362) from the begining.

## Alternatives

### Multiple CR:s

Instead of having a single CR to manage organiaztions, teams and users we could have multiple ones.

If teams, users or orgs get more "abilities" they are separated and can be changed independently.
I just feel it might assist in organizing all parts of the user/right management.

Also if upstream grafana decides to do a breaking change around one the CR:s we only need to bump one of the API versions,
assuming that they are not to hard coupled.

We could also reuse the same team/user definition in multiple organizations.

#### Multiple CRD:s

```org.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: org1
spec:
  - name: organiaztionX
```

```team.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: team1
spec:
    teams:
        - name: MyTestTeam
          email: email@test.com
          theme: dark
          homeDashboardId: 39
          timezone: utc
          memebers:
            - user1
            - user2

```

```user.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: user1
spec:
    users:
    - login: user1
        email: user1@github.com
        role: admin
        name: user1
        theme: light
        isGrafanaAdmin: true
        isDisabled: true
        password:
        secretName: mysecret
        key: password
        organization: organiaztionX
    - login: user2
        email: noreply@grafana.com
        role: Viewer
        organization: organiaztionX
```

### Adding org to Grafana CR

#### Adding org to Grafana PROS

- Orgs are part of a single Grafana instance. Therefore they should be added to the Grafana CR.
- One less CRD that needs to be maintained and if users need to be added to an Org this could be done in another CRD i.e. GrafanaUser.
- No issue regarding "timing", since the orgs could be applied instantly after the deployment is ready.
- URL for API-calls is already present

#### Adding org to Grafana CONS

- If there is a name change of an Org, it can not be updated, but pre-change-Org must be deleted and new one added afterwards.

## Work Plan

Implement the new CR one at the time.

## Open questions

### Same username

Need to find out if i can have the same username in different organiaztions.

### Ldap/openid users

I assume a user that is managed through ldap or similar can be a part of a team.
If logged in through LDAP my guess is that a local user gets created when the user logins in for the first time. Need to verify this.

## Related issues

- [408 Manage organizations](https://github.com/grafana-operator/grafana-operator/issues/408)
- [525 Support for orgId in GrafanaDashboard](https://github.com/grafana-operator/grafana-operator/issues/525)
- [174 Cannot create multiple Grafana in the same namespace](https://github.com/grafana-operator/grafana-operator/issues/174)
- [362 Refactor controllers to use controllerruntime CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362)

## References

- [Grafana organiaztion API](https://grafana.com/docs/grafana/latest/http_api/org/)
- [Grafana team API](https://grafana.com/docs/grafana/latest/http_api/team/)
- [Grafana user API](https://grafana.com/docs/grafana/latest/http_api/user/)
