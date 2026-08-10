---
title: "GrafanaSilence CRD"
linkTitle: "GrafanaSilence CRD"
---

## Summary

Add a new GrafanaSilence CRD so that the operator can provision
[Grafana Silences](https://grafana.com/docs/grafana/latest/alerting/configure-notifications/create-silence/)
automatically to the Grafana instance(s) it is managing. Currently silences
must be manually configured via the UI

## Info

status: Accepted

## Motivation

The Grafana Operator can declaratively manage most of the alerting lifecycle — alert rules, contact points, and notification policies — but alert silences can only be created and managed by hand through the Grafana UI (or the alert-generated silence URL). This is the one remaining alerting object with no Infrastructure-as-Code representation.

As a result, silences live only in the live state of a single Grafana instance. They are not version controlled, not peer
reviewed, not reproducible across rebuilds or across multiple instances, and cannot be produced by automation. This is fine for a human silencing a single alert reactively, but it breaks down for planned, recurring, or fleet-wide silences.

A GrafanaSilence CRD closes this gap by letting silences be defined declaratively, reviewed in git, and reconciled by the operator like every other Grafana resource.

## Use Case

* A team runs planned maintenance (node pool upgrades, database migrations, cluster patching) and knows in advance which alerts will fire as expected noise.
* The silence must exist before the maintenance begins — so there is no alert-generated URL to start from, and the UI flow (which assumes a human present at alert time) does not apply.
* The team wants the silence committed alongside the change that causes it, applied by the same pipeline, and removed automatically when the window closes.
* The same silence may need to apply identically across many Grafana instances / clusters, and must be auditable — who silenced what, why, and when — via git history and PR review.
* An organization enforces all monitoring changes require approval (changing alert thresholds or dashboards)
  * Ex. don't want an engineer silencing a production alert without approval!
  * Ex. Approval/consensus is a hard requirement for important monitoring changes
* An organization has robust rules in place and enforces business rules on the alerts
  * Shift-left type validation/business rules can be enforced on silence entries
  * Ex. Restrict teams from creating overly broad silence entries or shooting themselves in the foot, etc.
* Organizations which support multi-tenant/team grafana
  * It's important to have these in place as one team could accidentally silence another team's alerts by adding too broad a silence entry
  * Ex. Apply business logic above to restrict teams to only creating silence entries which match the labels or alerts they own
* Enables grafana owners to enforce stricter RBAC rules for silences in the UI - ex. `alert.silences:read`

Currently none of this is possible: silences must be clicked into each instance by hand and leave no reviewable record. A GrafanaSilence CRD lets the silence travel with the rest of the alerting config as code.

## Additional Use Case
* Enhances 3rd party alerting rules (kubernetes-mixin, prometheus-operator, etc.)
  Many out-of-the-box alerts - example is kubernetes-monitoring/kubernetes-mixin where the default is to alert on everything and there may not be facilities to customize the rule filters using jsonnet. We don't want to change the thresholds, but we'd like to exclude namespaces from some of the alerts. Being able to programmatically define silence entries would compliment these mixins and 3rd party alerts (default prometheus/mimir/loki/node/kubernetes/etc) by being able to use the off-the-shelf alerts but override them with exclusions without forking or modifying their default configurations. Right now we do not have a good way to customize these other than using jsonnet or custom generation scripts. Being able to deploy those and add our own silence rules for some of them would be helpful.
  This adds flexibility and makes using 3rd party alert rules even more powerful and practical.

## Verification

- Create e2e tests for the operator creating GrafanaSilence from baseline definition


## Proposal

Add a GrafanaSilence CRD that declaratively manages a silence on a target Grafana instance, using the same
instanceSelector/matcher patterns already established by the other Grafana Operator CRDs.

### Example Custom Resource

```yaml
---
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaSilence
metadata:
  name: silence-alert-example
  namespace: grafana
spec:
  instanceSelector:
    matchLabels:
      grafana: my-grafana-instance
  comment: "silencing my alert"
  createdBy: "DevOps Team"
  startsAt: "2026-06-24T20:00:00Z"
  endsAt: "2027-06-27T22:00:00Z"
  matchers:
    - name: alertname
      value: HighCPUUsage
      isEqual: true
      isRegex: false
    - name: env
      value: "prod|staging"
      isEqual: true
      isRegex: true
    - name: __alert_rule_uid__
      value: "my-alert-rule-uid" # TBD could handle this with a referencer
      isEqual: true
      isRegex: false
```

The operator reconciles the CR into a silence via the Grafana Alertmanager API and keeps it in sync.
Deleting the CR removes the silence; time-expired silences are reconciled to match Grafana's state.

### Associating Silences with Alert Rules
In order to associate a silence with an alert rule, the GrafanaSilence CRD will need to reference the alert rule by its UID. This is because alert rules can have the same name but different UIDs, and the UID is guaranteed to be unique.
The GrafanaSilence CRD's `matchers` field which allows users to specify matchers for the silence must define `__alert_rule_uid__` label.

It is possible to implement a reference field for this purpose if the alert rule is managed by the operator (not implemented as part of this proposal.)
This may be slightly tricky as the `GrafanaAlertRuleGroup` contains multiple alert rules, and the operator would need to find the correct alert rule by name and then use its UID for the silence. This is a potential future enhancement.

Note that if a user defines a `GrafanaSilence` without a `__alert_rule_uid__` this will still work as expected, they will just not appear "linked" to existing alerts.
As an alternative to the operator supporting a reference field, it may be preferable for end users to assign well-known `uid` for their alert rules in `GrafanaAlertRuleGroup` definitions

### Important Implementation Detail Regarding Silence IDs

Unlike other grafana entities which are assigned a UID by Grafana, silences are assigned a random ID by Grafana when created. This means that the operator must track the ID of the silence it created in order to update or delete it later. The operator will need to store this ID in the status of the GrafanaSilence CR, so that it can reconcile updates and deletions correctly.
An annotation has been declared for this purpose - `grafana.integreatly.org/silence-id`
The reasoning for an annotation instead of a status field is for the operator to be able to adopt/own existing silence entries.  This is not possible with a status field, but with an annotation, users can add the annotation for an existing silence entry and the operator will begin managing it.
