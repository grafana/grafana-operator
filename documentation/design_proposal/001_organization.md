# Design proposal 001 organization managment

There are a number of requests for managing organizations in grafana.

This document will try to consider future needs when it comes to

- Organizations
- Teams
- Users

We see possible ways. Either we add the Org to the Grafana CR or create a new CRD.

## Adding to Grafana CR

### Adding to Grafana PROS

Orgs are part of a single Grafana instance. Therefore they should be added to the Grafana CR.
One less CRD that needs to be maintained and if users need to be added to an Org this could be done in another CRD i.e. GrafanaUser.
No issue regarding "timing", since the orgs could be applied instantly after the deployment is ready.
URL for API-calls is already present

### Adding to Grafana CONS

If there is a name change of an Org, it can not be updated, but pre-change-Org must be deleted and new one added afterwards.

## Creating new CRD

### New CRD prod

Users can be added to the Org in the CR.
Rename is doable
IF the Orgs get new functionality it could be added easier

### New CRD CONS

Needs to be applied after the deployment is done.
API-url needs to be provided to the CR

### CRD

```.yaml
apiVersion: integreatly.org/v1alpha1
kind: Organization
metadata:
  name: grafana1
spec:
  - organiaztionX:
      users:
        - user1
          email: user1@github.com
          role: admin
          login: user1
          theme: light
          isGrafanaAdmin: true
          isDisabled: true
          password:
            secretName: mysecret
            key: password
        - user2:
          email: noreply@grafana.com
          role: Viewer
      teams:
        - MyTestTeam:
          email: email@test.com
          theme: dark
          homeDashboardId: 39
          timezone: utc
          memebers:
            - user1
            - user2
```

#### Thoughts

Should the Grafana CRD find the organization or should the organiazton find the grafana instance?
For example by using label selectors.

How should we manage the default organization? Can we manage it without making breaking chnages?

## TODO

Need to find out if i can have the same username in different organiaztions.

I assume a user that is managed through ldap or similar can be a part of a team.
If logged in through LDAP my guess is that a local user gets created when the user logins in for the first time. Need to verify this.

## Context

We need to consider the [multi-namespace support](https://github.com/grafana-operator/grafana-operator/pull/599) that is currently getting worked on.
The operator will start to manage multiple grafana instances in a single cluster.

In the future it might be intresting to [manage external grafana instances](https://github.com/grafana-operator/grafana-operator/issues/402) even though it's currently not being prioratized.
We haven't had any discussion about this but that might also be off the table considering [Avoiding scope creep](https://github.com/grafana-operator/grafana-operator/issues/497).

When creating the new controller we should use [CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362) from the begining.

## Related issues

- [408 Manage organizations](https://github.com/grafana-operator/grafana-operator/issues/408)
- [525 Support for orgId in GrafanaDashboard](https://github.com/grafana-operator/grafana-operator/issues/525)
- [174 Cannot create multiple Grafana in the same namespace](https://github.com/grafana-operator/grafana-operator/issues/174)
- [362 Refactor controllers to use controllerruntime CreateOrUpdate](https://github.com/grafana-operator/grafana-operator/issues/362)

## References

- [Grafana organiaztion API](https://grafana.com/docs/grafana/latest/http_api/org/)
- [Grafana team API](https://grafana.com/docs/grafana/latest/http_api/team/)
- [Grafana user API](https://grafana.com/docs/grafana/latest/http_api/user/)
