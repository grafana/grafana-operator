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

- Needs to be applied after the deployment is done.
- API-url needs to be provided to the CR together with username & secret.

### CRD

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: organization-example
spec:
  - name: organiaztionX
      grafana:
        - url: http://grafana/
          username: admin # default value
          password:
            secretName: mysecret
              key: password
        - url: https://external-grafana/
          username: admin # default value
          password:
            secretName: admin-secret2
              key: password
          tls:
            secretName: external-https-cert
      users:
        - login: user1
          email: user1@github.com
          role: admin
          name: user1
          theme: light
          isGrafanaAdmin: true
          isDisabled: true
          password:
            secretName: usersecret
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

### Options

The organaiztion CRD could also use a label selector just like we do for dashboardSelectors but the other way around.
Instead of the grafana instance finding which organizations to add the organaization will find which grafana instances to apply to.

This way we woulden't need to define organiaztion.spec.grafana.

## Context

We need to consider the [multi-namespace support](https://github.com/grafana-operator/grafana-operator/pull/599) that is currently getting worked on.
The operator will start to manage multiple grafana instances in a single cluster.

In the future it might be intresting to [manage external grafana instances](https://github.com/grafana-operator/grafana-operator/issues/402) even though it's currently not being prioratized.
We haven't had any discussion about this but that might also be off the table considering [Avoiding scope creep](https://github.com/grafana-operator/grafana-operator/issues/497).

When creating the new controller we should use [CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362) from the begining.

## Alternatives

Instead of having a single CRD to manage organiaztions, teams and users we could have multiple ones.

If teams, users or orgs get more "abilities" they are separated and can be changed independently.
I just feel it might assist in organizing all parts of the user/right management.

Also if upstream grafana decides to do a breaking change around one the CR:s we only need to bump one of the API versions,
assuming that they are not to hard coupled.

We could also reuse the same team/user definition in multiple organizations.

Thanks to all the extra config that is needed it will be easier to do an error.

### Thoughts

Should we assume that the oeprator owns the organiaztion? If not we will have to define the grafana-url & username & secret
in the team and user CRD as well. Unless the grafana instances find organiaztion, team, users.

How should we find the organiaztion where we should create the user/team? Just define a name or use a selector?

In the example below organiaztion is part of spec.teams.organiaztion. Another option could be to put organiaztion under spec.organiaztion.
This way a team definition can only be used in one organiaztion. It would most likley lower the amount of miss configuration.
We need to decide if we should support multiple organiaztions in one team CRD, same thing with user.

### Multiple CRD

```org.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaOrganization
metadata:
  name: org1
spec:
  - name: organiaztionX
      grafana:
        - url: http://grafana/
          username: admin # default value
          password:
            secretName: mysecret
              key: password
```

```team.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaTeam
metadata:
  name: team1
spec:
    teams:
    - name: MyTestTeam
        email: email@test.com
        theme: dark
        homeDashboardId: 39
        timezone: utc
        organization: organiaztionX
        memebers:
          - user1
          - user2
```

```user.yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaUser
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

## Work Plan

Implement the new CR one at the time.

## Open questions
- How do we manage each of these new CRs? would it be through a higher level operator that manages organization related resources? or would it be a controller per resource?
### Same username

Need to find out if i can have the same username in different organiaztions.

### Ldap/openid users

I assume a user that is managed through ldap or similar can be a part of a team.
If logged in through LDAP my guess is that a local user gets created when the user logins in for the first time. Need to verify this.

### Default organaiztion

Does the default organaiztion have a name? The id is 1.
Is there any config in grafana.ini that might create issues with default organiaztion vs not?
For example we probably need to test a bit how [auto_assign_org](https://grafana.com/docs/grafana/latest/administration/configuration/#auto_assign_org) works.

## Related issues

- [408 Manage organizations](https://github.com/grafana-operator/grafana-operator/issues/408)
- [525 Support for orgId in GrafanaDashboard](https://github.com/grafana-operator/grafana-operator/issues/525)
- [174 Cannot create multiple Grafana in the same namespace](https://github.com/grafana-operator/grafana-operator/issues/174)
- [362 Refactor controllers to use controllerruntime CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362)

## References

- [Grafana organiaztion API](https://grafana.com/docs/grafana/latest/http_api/org/)
- [Grafana team API](https://grafana.com/docs/grafana/latest/http_api/team/)
- [Grafana user API](https://grafana.com/docs/grafana/latest/http_api/user/)
- [Grafana config auto_assign_org](https://grafana.com/docs/grafana/latest/administration/configuration/#auto_assign_org)
