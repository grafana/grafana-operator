# Working with notifier channels

This document describes how to create notifiers in grafana.

## Notifier properties

Notifiers are represented by the `GrafanaNotificationChannel` resource

Examples can be found in `deploy/examples/notificationchannels`

Also look at the api documentation found [here](../documentation/api.md)

### Creating notifiers

*Note:* Thus far this is the only supported method of notifier provisioning.

The recommended procedure for creating notifiers is as follows:

1) Create a Grafana Operator deployment as per docs.
2) Log in using admin credentials, these can be found in either:
   - The `grafana-admin-credentials` secret.
   - User generated secret, and supplemented in `EnvFrom` in the Grafana CR.
3) Once logged in, create and test notifiers through the UI
4) Once tested and created, extract the raw JSON
5) Create a new `GrafanaNotifgicationChannel` CR and provide the JSON string in `spec.json` as in the example below:
6) Apply the resource to the cluster

The created notifier should now be provisioned and managed by the operator.

```yaml
apiVersion: integreatly.org/v1alpha1
kind: GrafanaNotificationChannel
metadata:
  name: pager-duty-channel
  labels:
    app: grafana
spec:
  name: pager-duty-channel.json
  json: >
    {
      "uid": "pager-duty-alert-notification",
      "name": "Pager Duty alert notification",
      "type":  "pagerduty",
      "isDefault": true,
      "sendReminder": true,
      "frequency": "15m",
      "disableResolveMessage": true,
      "settings": {
        "integrationKey": "put key here",
        "autoResolve": true,
        "uploadImage": true
     }
    }
```
