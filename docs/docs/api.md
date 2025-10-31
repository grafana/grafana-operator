---
title: API Reference
weight: 100
---

Packages:

- [grafana.integreatly.org/v1beta1](#grafanaintegreatlyorgv1beta1)

# grafana.integreatly.org/v1beta1

Resource Types:

- [GrafanaAlertRuleGroup](#grafanaalertrulegroup)

- [GrafanaContactPoint](#grafanacontactpoint)

- [GrafanaDashboard](#grafanadashboard)

- [GrafanaDatasource](#grafanadatasource)

- [GrafanaFolder](#grafanafolder)

- [GrafanaLibraryPanel](#grafanalibrarypanel)

- [GrafanaMuteTiming](#grafanamutetiming)

- [GrafanaNotificationPolicy](#grafananotificationpolicy)

- [GrafanaNotificationPolicyRoute](#grafananotificationpolicyroute)

- [GrafanaNotificationTemplate](#grafananotificationtemplate)

- [Grafana](#grafana)

- [GrafanaServiceAccount](#grafanaserviceaccount)




## GrafanaAlertRuleGroup
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaAlertRuleGroup is the Schema for the grafanaalertrulegroups API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaAlertRuleGroup</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaAlertRuleGroupSpec defines the desired state of GrafanaAlertRuleGroup<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))): Only one of FolderUID or FolderRef can be set and one must be defined</li><li>((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable))): spec.editable is immutable</li><li>((!has(oldSelf.folderUID) && !has(self.folderUID)) || (has(oldSelf.folderUID) && has(self.folderUID))): spec.folderUID is immutable</li><li>((!has(oldSelf.folderRef) && !has(self.folderRef)) || (has(oldSelf.folderRef) && has(self.folderRef))): spec.folderRef is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupstatus">status</a></b></td>
        <td>object</td>
        <td>
          The most recent observed state of a Grafana resource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec
<sup><sup>[↩ Parent](#grafanaalertrulegroup)</sup></sup>



GrafanaAlertRuleGroupSpec defines the desired state of GrafanaAlertRuleGroup

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaalertrulegroupspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>interval</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: duration<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspecrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editable</b></td>
        <td>boolean</td>
        <td>
          Whether to enable or disable editing of the alert rule group in Grafana UI<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: Value is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderRef</b></td>
        <td>string</td>
        <td>
          Match GrafanaFolders CRs to infer the uid<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: Value is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderUID</b></td>
        <td>string</td>
        <td>
          UID of the folder containing this rule group
Overrides the FolderSelector<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: Value is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the alert rule group. If not specified, the resource name will be used.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanaalertrulegroupspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaalertrulegroupspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaalertrulegroupspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.rules[index]
<sup><sup>[↩ Parent](#grafanaalertrulegroupspec)</sup></sup>



AlertRule defines a specific rule to be evaluated. It is based on the upstream model with some k8s specific type mappings

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>condition</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspecrulesindexdataindex">data</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>execErrState</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: OK, Alerting, Error, KeepLast<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>for</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: duration<br/>
            <i>Default</i>: 0s<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>noDataState</b></td>
        <td>enum</td>
        <td>
          <br/>
          <br/>
            <i>Enum</i>: Alerting, NoData, OK, KeepLast<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the alert rule. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>isPaused</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keepFiringFor</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: duration<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>missingSeriesEvalsToResolve</b></td>
        <td>integer</td>
        <td>
          The number of missing series evaluations that must occur before the rule is considered to be resolved.<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspecrulesindexnotificationsettings">notificationSettings</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspecrulesindexrecord">record</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.rules[index].data[index]
<sup><sup>[↩ Parent](#grafanaalertrulegroupspecrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasourceUid</b></td>
        <td>string</td>
        <td>
          Grafana data source unique identifier; it should be '__expr__' for a Server Side Expression operation.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>model</b></td>
        <td>JSON</td>
        <td>
          JSON is the raw JSON query and includes the above properties as well as custom properties.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>queryType</b></td>
        <td>string</td>
        <td>
          QueryType is an optional identifier for the type of query.
It can be used to distinguish different types of queries.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>refId</b></td>
        <td>string</td>
        <td>
          RefID is the unique identifier of the query, set by the frontend call.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaalertrulegroupspecrulesindexdataindexrelativetimerange">relativeTimeRange</a></b></td>
        <td>object</td>
        <td>
          relative time range<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.rules[index].data[index].relativeTimeRange
<sup><sup>[↩ Parent](#grafanaalertrulegroupspecrulesindexdataindex)</sup></sup>



relative time range

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>from</b></td>
        <td>integer</td>
        <td>
          from<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>to</b></td>
        <td>integer</td>
        <td>
          to<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.rules[index].notificationSettings
<sup><sup>[↩ Parent](#grafanaalertrulegroupspecrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>receiver</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group_by</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_interval</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_wait</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mute_time_intervals</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>repeat_interval</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.spec.rules[index].record
<sup><sup>[↩ Parent](#grafanaalertrulegroupspecrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>from</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>metric</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.status
<sup><sup>[↩ Parent](#grafanaalertrulegroup)</sup></sup>



The most recent observed state of a Grafana resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaalertrulegroupstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaAlertRuleGroup.status.conditions[index]
<sup><sup>[↩ Parent](#grafanaalertrulegroupstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaContactPoint
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaContactPoint is the Schema for the grafanacontactpoints API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaContactPoint</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanacontactpointspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaContactPointSpec defines the desired state of GrafanaContactPoint<br/>
          <br/>
            <i>Validations</i>:<li>((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid))): spec.uid is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanacontactpointstatus">status</a></b></td>
        <td>object</td>
        <td>
          The most recent observed state of a Grafana resource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec
<sup><sup>[↩ Parent](#grafanacontactpoint)</sup></sup>



GrafanaContactPointSpec defines the desired state of GrafanaContactPoint

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanacontactpointspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>settings</b></td>
        <td>JSON</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disableResolveMessage</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          Manually specify the UID the Contact Point is created with. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.uid is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanacontactpointspecvaluesfromindex">valuesFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanacontactpointspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanacontactpointspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanacontactpointspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.valuesFrom[index]
<sup><sup>[↩ Parent](#grafanacontactpointspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>targetPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanacontactpointspecvaluesfromindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          <br/>
          <br/>
            <i>Validations</i>:<li>(has(self.configMapKeyRef) && !has(self.secretKeyRef)) || (!has(self.configMapKeyRef) && has(self.secretKeyRef)): Either configMapKeyRef or secretKeyRef must be set</li>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.valuesFrom[index].valueFrom
<sup><sup>[↩ Parent](#grafanacontactpointspecvaluesfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanacontactpointspecvaluesfromindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanacontactpointspecvaluesfromindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.valuesFrom[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanacontactpointspecvaluesfromindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.spec.valuesFrom[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanacontactpointspecvaluesfromindexvaluefrom)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.status
<sup><sup>[↩ Parent](#grafanacontactpoint)</sup></sup>



The most recent observed state of a Grafana resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanacontactpointstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaContactPoint.status.conditions[index]
<sup><sup>[↩ Parent](#grafanacontactpointstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaDashboard
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaDashboard is the Schema for the grafanadashboards API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaDashboard</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaDashboardSpec defines the desired state of GrafanaDashboard<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID))): Only one of folderUID or folderRef can be declared at the same time</li><li>(has(self.folder) && !(has(self.folderRef) || has(self.folderUID))) || !(has(self.folder)): folder field cannot be set when folderUID or folderRef is already declared</li><li>((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid))): spec.uid is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardstatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaDashboardStatus defines the observed state of GrafanaDashboard<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec
<sup><sup>[↩ Parent](#grafanadashboard)</sup></sup>



GrafanaDashboardSpec defines the desired state of GrafanaDashboard

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          model from configmap<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentCacheDuration</b></td>
        <td>string</td>
        <td>
          Cache duration for models fetched from URLs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecdatasourcesindex">datasources</a></b></td>
        <td>[]object</td>
        <td>
          maps required data sources to existing ones<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          environments variables from secrets or config maps<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecenvsindex">envs</a></b></td>
        <td>[]object</td>
        <td>
          environments variables as a map<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folder</b></td>
        <td>string</td>
        <td>
          folder assignment for dashboard<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderRef</b></td>
        <td>string</td>
        <td>
          Name of a `GrafanaFolder` resource in the same namespace<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderUID</b></td>
        <td>string</td>
        <td>
          UID of the target folder for this dashboard<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecgrafanacom">grafanaCom</a></b></td>
        <td>object</td>
        <td>
          grafana.com/dashboards<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gzipJson</b></td>
        <td>string</td>
        <td>
          GzipJson the model's JSON compressed with Gzip. Base64-encoded when in YAML.<br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>json</b></td>
        <td>string</td>
        <td>
          model json<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jsonnet</b></td>
        <td>string</td>
        <td>
          Jsonnet<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecjsonnetlib">jsonnetLib</a></b></td>
        <td>object</td>
        <td>
          Jsonnet project build<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecpluginsindex">plugins</a></b></td>
        <td>[]object</td>
        <td>
          plugins<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          Manually specify the uid, overwrites uids already present in the json model.
Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40.<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.uid is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          model url<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecurlauthorization">urlAuthorization</a></b></td>
        <td>object</td>
        <td>
          authorization options for model from url<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanadashboardspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.configMapRef
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



model from configmap

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.datasources[index]
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



GrafanaResourceDatasource is used to set the datasource name of any templated datasources in
content definitions (e.g., dashboard JSON).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasourceName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>inputName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envFrom[index]
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecenvfromindexconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecenvfromindexsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envFrom[index].configMapKeyRef
<sup><sup>[↩ Parent](#grafanadashboardspecenvfromindex)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envFrom[index].secretKeyRef
<sup><sup>[↩ Parent](#grafanadashboardspecenvfromindex)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envs[index]
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Inline env value<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecenvsindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Reference on value source, might be the reference on a secret or config map<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envs[index].valueFrom
<sup><sup>[↩ Parent](#grafanadashboardspecenvsindex)</sup></sup>



Reference on value source, might be the reference on a secret or config map

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecenvsindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecenvsindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envs[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanadashboardspecenvsindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.envs[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanadashboardspecenvsindexvaluefrom)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.grafanaCom
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



grafana.com/dashboards

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>revision</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.jsonnetLib
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



Jsonnet project build

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fileName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>gzipJsonnetProject</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>jPath</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.plugins[index]
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.urlAuthorization
<sup><sup>[↩ Parent](#grafanadashboardspec)</sup></sup>



authorization options for model from url

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecurlauthorizationbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.urlAuthorization.basicAuth
<sup><sup>[↩ Parent](#grafanadashboardspecurlauthorization)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadashboardspecurlauthorizationbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          SecretKeySelector selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardspecurlauthorizationbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          SecretKeySelector selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.urlAuthorization.basicAuth.password
<sup><sup>[↩ Parent](#grafanadashboardspecurlauthorizationbasicauth)</sup></sup>



SecretKeySelector selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.spec.urlAuthorization.basicAuth.username
<sup><sup>[↩ Parent](#grafanadashboardspecurlauthorizationbasicauth)</sup></sup>



SecretKeySelector selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.status
<sup><sup>[↩ Parent](#grafanadashboard)</sup></sup>



GrafanaDashboardStatus defines the observed state of GrafanaDashboard

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>NoMatchingInstances</b></td>
        <td>boolean</td>
        <td>
          The dashboard instanceSelector can't find matching grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadashboardstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentCache</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentTimestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentUrl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hash</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDashboard.status.conditions[index]
<sup><sup>[↩ Parent](#grafanadashboardstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaDatasource
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaDatasource is the Schema for the grafanadatasources API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaDatasource</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaDatasourceSpec defines the desired state of GrafanaDatasource<br/>
          <br/>
            <i>Validations</i>:<li>((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid))): spec.uid is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcestatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaDatasourceStatus defines the observed state of GrafanaDatasource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec
<sup><sup>[↩ Parent](#grafanadatasource)</sup></sup>



GrafanaDatasourceSpec defines the desired state of GrafanaDatasource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadatasourcespecdatasource">datasource</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespecpluginsindex">plugins</a></b></td>
        <td>[]object</td>
        <td>
          plugins<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          The UID, for the datasource, fallback to the deprecated spec.datasource.uid
and metadata.uid. Can be any string consisting of alphanumeric characters,
- and _ with a maximum length of 40 +optional<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.uid is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespecvaluesfromindex">valuesFrom</a></b></td>
        <td>[]object</td>
        <td>
          environments variables from secrets or config maps<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.datasource
<sup><sup>[↩ Parent](#grafanadatasourcespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>access</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>basicAuth</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>basicAuthUser</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>database</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editable</b></td>
        <td>boolean</td>
        <td>
          Whether to enable/disable editing of the datasource in Grafana UI<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>isDefault</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jsonData</b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>orgId</b></td>
        <td>integer</td>
        <td>
          Deprecated field, it has no effect<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secureJsonData</b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          Deprecated field, use spec.uid instead<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanadatasourcespec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadatasourcespecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanadatasourcespecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.plugins[index]
<sup><sup>[↩ Parent](#grafanadatasourcespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.valuesFrom[index]
<sup><sup>[↩ Parent](#grafanadatasourcespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>targetPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespecvaluesfromindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          <br/>
          <br/>
            <i>Validations</i>:<li>(has(self.configMapKeyRef) && !has(self.secretKeyRef)) || (!has(self.configMapKeyRef) && has(self.secretKeyRef)): Either configMapKeyRef or secretKeyRef must be set</li>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.valuesFrom[index].valueFrom
<sup><sup>[↩ Parent](#grafanadatasourcespecvaluesfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanadatasourcespecvaluesfromindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcespecvaluesfromindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.valuesFrom[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanadatasourcespecvaluesfromindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.spec.valuesFrom[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanadatasourcespecvaluesfromindexvaluefrom)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.status
<sup><sup>[↩ Parent](#grafanadatasource)</sup></sup>



GrafanaDatasourceStatus defines the observed state of GrafanaDatasource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>NoMatchingInstances</b></td>
        <td>boolean</td>
        <td>
          The datasource instanceSelector can't find matching grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanadatasourcestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hash</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastMessage</b></td>
        <td>string</td>
        <td>
          Deprecated: Check status.conditions or operator logs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaDatasource.status.conditions[index]
<sup><sup>[↩ Parent](#grafanadatasourcestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaFolder
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaFolder is the Schema for the grafanafolders API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaFolder</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanafolderspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaFolderSpec defines the desired state of GrafanaFolder<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.parentFolderUID) && !(has(self.parentFolderRef))) || (has(self.parentFolderRef) && !(has(self.parentFolderUID))) || !(has(self.parentFolderRef) && (has(self.parentFolderUID))): Only one of parentFolderUID or parentFolderRef can be set</li><li>((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid))): spec.uid is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanafolderstatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaFolderStatus defines the observed state of GrafanaFolder<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaFolder.spec
<sup><sup>[↩ Parent](#grafanafolder)</sup></sup>



GrafanaFolderSpec defines the desired state of GrafanaFolder

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanafolderspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>parentFolderRef</b></td>
        <td>string</td>
        <td>
          Reference to an existing GrafanaFolder CR in the same namespace<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>parentFolderUID</b></td>
        <td>string</td>
        <td>
          UID of the folder in which the current folder should be created<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>permissions</b></td>
        <td>string</td>
        <td>
          Raw json with folder permissions, potentially exported from Grafana<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>title</b></td>
        <td>string</td>
        <td>
          Display name of the folder in Grafana<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          Manually specify the UID the Folder is created with. Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.uid is immutable</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaFolder.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanafolderspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanafolderspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaFolder.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanafolderspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaFolder.status
<sup><sup>[↩ Parent](#grafanafolder)</sup></sup>



GrafanaFolderStatus defines the observed state of GrafanaFolder

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>NoMatchingInstances</b></td>
        <td>boolean</td>
        <td>
          The folder instanceSelector can't find matching grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanafolderstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hash</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaFolder.status.conditions[index]
<sup><sup>[↩ Parent](#grafanafolderstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaLibraryPanel
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaLibraryPanel is the Schema for the grafanalibrarypanels API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaLibraryPanel</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaLibraryPanelSpec defines the desired state of GrafanaLibraryPanel<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.folderUID) && !(has(self.folderRef))) || (has(self.folderRef) && !(has(self.folderUID))) || !(has(self.folderRef) && (has(self.folderUID))): Only one of folderUID or folderRef can be declared at the same time</li><li>((!has(oldSelf.uid) && !has(self.uid)) || (has(oldSelf.uid) && has(self.uid))): spec.uid is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelstatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaLibraryPanelStatus defines the observed state of GrafanaLibraryPanel<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec
<sup><sup>[↩ Parent](#grafanalibrarypanel)</sup></sup>



GrafanaLibraryPanelSpec defines the desired state of GrafanaLibraryPanel

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          model from configmap<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentCacheDuration</b></td>
        <td>string</td>
        <td>
          Cache duration for models fetched from URLs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecdatasourcesindex">datasources</a></b></td>
        <td>[]object</td>
        <td>
          maps required data sources to existing ones<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          environments variables from secrets or config maps<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecenvsindex">envs</a></b></td>
        <td>[]object</td>
        <td>
          environments variables as a map<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderRef</b></td>
        <td>string</td>
        <td>
          Name of a `GrafanaFolder` resource in the same namespace<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folderUID</b></td>
        <td>string</td>
        <td>
          UID of the target folder for this dashboard<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecgrafanacom">grafanaCom</a></b></td>
        <td>object</td>
        <td>
          grafana.com/dashboards<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gzipJson</b></td>
        <td>string</td>
        <td>
          GzipJson the model's JSON compressed with Gzip. Base64-encoded when in YAML.<br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>json</b></td>
        <td>string</td>
        <td>
          model json<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>jsonnet</b></td>
        <td>string</td>
        <td>
          Jsonnet<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecjsonnetlib">jsonnetLib</a></b></td>
        <td>object</td>
        <td>
          Jsonnet project build<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecpluginsindex">plugins</a></b></td>
        <td>[]object</td>
        <td>
          plugins<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          Manually specify the uid, overwrites uids already present in the json model.
Can be any string consisting of alphanumeric characters, - and _ with a maximum length of 40.<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.uid is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          model url<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecurlauthorization">urlAuthorization</a></b></td>
        <td>object</td>
        <td>
          authorization options for model from url<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.configMapRef
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



model from configmap

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.datasources[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



GrafanaResourceDatasource is used to set the datasource name of any templated datasources in
content definitions (e.g., dashboard JSON).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasourceName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>inputName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envFrom[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecenvfromindexconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecenvfromindexsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envFrom[index].configMapKeyRef
<sup><sup>[↩ Parent](#grafanalibrarypanelspecenvfromindex)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envFrom[index].secretKeyRef
<sup><sup>[↩ Parent](#grafanalibrarypanelspecenvfromindex)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envs[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Inline env value<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecenvsindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          Reference on value source, might be the reference on a secret or config map<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envs[index].valueFrom
<sup><sup>[↩ Parent](#grafanalibrarypanelspecenvsindex)</sup></sup>



Reference on value source, might be the reference on a secret or config map

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecenvsindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a ConfigMap.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecenvsindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          Selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envs[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanalibrarypanelspecenvsindexvaluefrom)</sup></sup>



Selects a key of a ConfigMap.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key to select.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the ConfigMap or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.envs[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanalibrarypanelspecenvsindexvaluefrom)</sup></sup>



Selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.grafanaCom
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



grafana.com/dashboards

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>revision</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.jsonnetLib
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



Jsonnet project build

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fileName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>gzipJsonnetProject</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>jPath</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.plugins[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.urlAuthorization
<sup><sup>[↩ Parent](#grafanalibrarypanelspec)</sup></sup>



authorization options for model from url

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecurlauthorizationbasicauth">basicAuth</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.urlAuthorization.basicAuth
<sup><sup>[↩ Parent](#grafanalibrarypanelspecurlauthorization)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelspecurlauthorizationbasicauthpassword">password</a></b></td>
        <td>object</td>
        <td>
          SecretKeySelector selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanalibrarypanelspecurlauthorizationbasicauthusername">username</a></b></td>
        <td>object</td>
        <td>
          SecretKeySelector selects a key of a Secret.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.urlAuthorization.basicAuth.password
<sup><sup>[↩ Parent](#grafanalibrarypanelspecurlauthorizationbasicauth)</sup></sup>



SecretKeySelector selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.spec.urlAuthorization.basicAuth.username
<sup><sup>[↩ Parent](#grafanalibrarypanelspecurlauthorizationbasicauth)</sup></sup>



SecretKeySelector selects a key of a Secret.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.status
<sup><sup>[↩ Parent](#grafanalibrarypanel)</sup></sup>



GrafanaLibraryPanelStatus defines the observed state of GrafanaLibraryPanel

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanalibrarypanelstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentCache</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: byte<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentTimestamp</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contentUrl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hash</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaLibraryPanel.status.conditions[index]
<sup><sup>[↩ Parent](#grafanalibrarypanelstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaMuteTiming
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaMuteTiming is the Schema for the GrafanaMuteTiming API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaMuteTiming</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanamutetimingspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaMuteTimingSpec defines the desired state of GrafanaMuteTiming<br/>
          <br/>
            <i>Validations</i>:<li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanamutetimingstatus">status</a></b></td>
        <td>object</td>
        <td>
          The most recent observed state of a Grafana resource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.spec
<sup><sup>[↩ Parent](#grafanamutetiming)</sup></sup>



GrafanaMuteTimingSpec defines the desired state of GrafanaMuteTiming

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanamutetimingspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          A unique name for the mute timing<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanamutetimingspectime_intervalsindex">time_intervals</a></b></td>
        <td>[]object</td>
        <td>
          Time intervals for muting<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editable</b></td>
        <td>boolean</td>
        <td>
          Whether to enable or disable editing of the mute timing in Grafana UI<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.editable is immutable</li>
            <i>Default</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.spec.instanceSelector
<sup><sup>[↩ Parent](#grafanamutetimingspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanamutetimingspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanamutetimingspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.spec.time_intervals[index]
<sup><sup>[↩ Parent](#grafanamutetimingspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>days_of_month</b></td>
        <td>[]string</td>
        <td>
          The date 1-31 of a month. Negative values can also be used to represent days that begin at the end of the month.
For example: -1 for the last day of the month.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>location</b></td>
        <td>string</td>
        <td>
          Depending on the location, the time range is displayed in local time.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>months</b></td>
        <td>[]string</td>
        <td>
          The months of the year in either numerical or the full calendar month.
For example: 1, may.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanamutetimingspectime_intervalsindextimesindex">times</a></b></td>
        <td>[]object</td>
        <td>
          The time inclusive of the start and exclusive of the end time (in UTC if no location has been selected, otherwise local time).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>weekdays</b></td>
        <td>[]string</td>
        <td>
          The day or range of days of the week.
For example: monday, thursday<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>years</b></td>
        <td>[]string</td>
        <td>
          The year or years for the interval.
For example: 2021<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.spec.time_intervals[index].times[index]
<sup><sup>[↩ Parent](#grafanamutetimingspectime_intervalsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>end_time</b></td>
        <td>string</td>
        <td>
          end time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>start_time</b></td>
        <td>string</td>
        <td>
          start time<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.status
<sup><sup>[↩ Parent](#grafanamutetiming)</sup></sup>



The most recent observed state of a Grafana resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanamutetimingstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaMuteTiming.status.conditions[index]
<sup><sup>[↩ Parent](#grafanamutetimingstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaNotificationPolicy
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaNotificationPolicy is the Schema for the GrafanaNotificationPolicy API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaNotificationPolicy</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaNotificationPolicySpec defines the desired state of GrafanaNotificationPolicy<br/>
          <br/>
            <i>Validations</i>:<li>((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable))): spec.editable is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicystatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaNotificationPolicyStatus defines the observed state of GrafanaNotificationPolicy<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec
<sup><sup>[↩ Parent](#grafananotificationpolicy)</sup></sup>



GrafanaNotificationPolicySpec defines the desired state of GrafanaNotificationPolicy

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicyspecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyspecroute">route</a></b></td>
        <td>object</td>
        <td>
          Routes for alerts to match against<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editable</b></td>
        <td>boolean</td>
        <td>
          Whether to enable or disable editing of the notification policy in Grafana UI<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: Value is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.instanceSelector
<sup><sup>[↩ Parent](#grafananotificationpolicyspec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicyspecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyspecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.route
<sup><sup>[↩ Parent](#grafananotificationpolicyspec)</sup></sup>



Routes for alerts to match against

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>receiver</b></td>
        <td>string</td>
        <td>
          receiver<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>active_time_intervals</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>continue</b></td>
        <td>boolean</td>
        <td>
          continue<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_by</b></td>
        <td>[]string</td>
        <td>
          group by<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_interval</b></td>
        <td>string</td>
        <td>
          group interval<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_wait</b></td>
        <td>string</td>
        <td>
          group wait<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>match_re</b></td>
        <td>map[string]string</td>
        <td>
          match re<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyspecroutematchersindex">matchers</a></b></td>
        <td>[]object</td>
        <td>
          matchers<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mute_time_intervals</b></td>
        <td>[]string</td>
        <td>
          mute time intervals<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>object_matchers</b></td>
        <td>[][]string</td>
        <td>
          object matchers<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>provenance</b></td>
        <td>string</td>
        <td>
          provenance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>repeat_interval</b></td>
        <td>string</td>
        <td>
          repeat interval<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyspecrouterouteselector">routeSelector</a></b></td>
        <td>object</td>
        <td>
          selects GrafanaNotificationPolicyRoutes to merge in when specified
mutually exclusive with Routes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>routes</b></td>
        <td>JSON</td>
        <td>
          routes, mutually exclusive with RouteSelector<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.route.matchers[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyspecroute)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>isRegex</b></td>
        <td>boolean</td>
        <td>
          is regex<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          value<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>isEqual</b></td>
        <td>boolean</td>
        <td>
          is equal<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.route.routeSelector
<sup><sup>[↩ Parent](#grafananotificationpolicyspecroute)</sup></sup>



selects GrafanaNotificationPolicyRoutes to merge in when specified
mutually exclusive with Routes

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicyspecrouterouteselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.spec.route.routeSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyspecrouterouteselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.status
<sup><sup>[↩ Parent](#grafananotificationpolicy)</sup></sup>



GrafanaNotificationPolicyStatus defines the observed state of GrafanaNotificationPolicy

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicystatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>discoveredRoutes</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicy.status.conditions[index]
<sup><sup>[↩ Parent](#grafananotificationpolicystatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaNotificationPolicyRoute
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaNotificationPolicyRoute is the Schema for the grafananotificationpolicyroutes API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaNotificationPolicyRoute</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyroutespec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaNotificationPolicyRouteSpec defines the desired state of GrafanaNotificationPolicyRoute<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyroutestatus">status</a></b></td>
        <td>object</td>
        <td>
          The most recent observed state of a Grafana resource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.spec
<sup><sup>[↩ Parent](#grafananotificationpolicyroute)</sup></sup>



GrafanaNotificationPolicyRouteSpec defines the desired state of GrafanaNotificationPolicyRoute

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>receiver</b></td>
        <td>string</td>
        <td>
          receiver<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>active_time_intervals</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>continue</b></td>
        <td>boolean</td>
        <td>
          continue<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_by</b></td>
        <td>[]string</td>
        <td>
          group by<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_interval</b></td>
        <td>string</td>
        <td>
          group interval<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group_wait</b></td>
        <td>string</td>
        <td>
          group wait<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>match_re</b></td>
        <td>map[string]string</td>
        <td>
          match re<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyroutespecmatchersindex">matchers</a></b></td>
        <td>[]object</td>
        <td>
          matchers<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mute_time_intervals</b></td>
        <td>[]string</td>
        <td>
          mute time intervals<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>object_matchers</b></td>
        <td>[][]string</td>
        <td>
          object matchers<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>provenance</b></td>
        <td>string</td>
        <td>
          provenance<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>repeat_interval</b></td>
        <td>string</td>
        <td>
          repeat interval<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafananotificationpolicyroutespecrouteselector">routeSelector</a></b></td>
        <td>object</td>
        <td>
          selects GrafanaNotificationPolicyRoutes to merge in when specified
mutually exclusive with Routes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>routes</b></td>
        <td>JSON</td>
        <td>
          routes, mutually exclusive with RouteSelector<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.spec.matchers[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyroutespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>isRegex</b></td>
        <td>boolean</td>
        <td>
          is regex<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          value<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>isEqual</b></td>
        <td>boolean</td>
        <td>
          is equal<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.spec.routeSelector
<sup><sup>[↩ Parent](#grafananotificationpolicyroutespec)</sup></sup>



selects GrafanaNotificationPolicyRoutes to merge in when specified
mutually exclusive with Routes

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicyroutespecrouteselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.spec.routeSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyroutespecrouteselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.status
<sup><sup>[↩ Parent](#grafananotificationpolicyroute)</sup></sup>



The most recent observed state of a Grafana resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationpolicyroutestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationPolicyRoute.status.conditions[index]
<sup><sup>[↩ Parent](#grafananotificationpolicyroutestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaNotificationTemplate
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaNotificationTemplate is the Schema for the GrafanaNotificationTemplate API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaNotificationTemplate</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationtemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaNotificationTemplateSpec defines the desired state of GrafanaNotificationTemplate<br/>
          <br/>
            <i>Validations</i>:<li>((!has(oldSelf.editable) && !has(self.editable)) || (has(oldSelf.editable) && has(self.editable))): spec.editable is immutable</li><li>!oldSelf.allowCrossNamespaceImport || (oldSelf.allowCrossNamespaceImport && self.allowCrossNamespaceImport): disabling spec.allowCrossNamespaceImport requires a recreate to ensure desired state</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafananotificationtemplatestatus">status</a></b></td>
        <td>object</td>
        <td>
          The most recent observed state of a Grafana resource<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationTemplate.spec
<sup><sup>[↩ Parent](#grafananotificationtemplate)</sup></sup>



GrafanaNotificationTemplateSpec defines the desired state of GrafanaNotificationTemplate

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationtemplatespecinstanceselector">instanceSelector</a></b></td>
        <td>object</td>
        <td>
          Selects Grafana instances for import<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceSelector is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Template name<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>allowCrossNamespaceImport</b></td>
        <td>boolean</td>
        <td>
          Allow the Operator to match this resource with Grafanas outside the current namespace<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>editable</b></td>
        <td>boolean</td>
        <td>
          Whether to enable or disable editing of the notification template in Grafana UI<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.editable is immutable</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses synchronizing attempts and tells the operator to ignore changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>template</b></td>
        <td>string</td>
        <td>
          Template content<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationTemplate.spec.instanceSelector
<sup><sup>[↩ Parent](#grafananotificationtemplatespec)</sup></sup>



Selects Grafana instances for import

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationtemplatespecinstanceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationTemplate.spec.instanceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafananotificationtemplatespecinstanceselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationTemplate.status
<sup><sup>[↩ Parent](#grafananotificationtemplate)</sup></sup>



The most recent observed state of a Grafana resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafananotificationtemplatestatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaNotificationTemplate.status.conditions[index]
<sup><sup>[↩ Parent](#grafananotificationtemplatestatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Grafana
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






Grafana is the Schema for the grafanas API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Grafana</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaSpec defines the desired state of Grafana<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanastatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaStatus defines the observed state of Grafana<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec
<sup><sup>[↩ Parent](#grafana)</sup></sup>



GrafanaSpec defines the desired state of Grafana

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecclient">client</a></b></td>
        <td>object</td>
        <td>
          Client defines how the grafana-operator talks to the grafana instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>config</b></td>
        <td>map[string]map[string]string</td>
        <td>
          Config defines how your grafana ini file should looks like.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeployment">deployment</a></b></td>
        <td>object</td>
        <td>
          Deployment sets how the deployment object should look like with your grafana instance, contains a number of defaults.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disableDefaultAdminSecret</b></td>
        <td>boolean</td>
        <td>
          DisableDefaultAdminSecret prevents operator from creating default admin-credentials secret<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>disableDefaultSecurityContext</b></td>
        <td>enum</td>
        <td>
          DisableDefaultSecurityContext prevents the operator from populating securityContext on deployments<br/>
          <br/>
            <i>Enum</i>: Pod, Container, All<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecexternal">external</a></b></td>
        <td>object</td>
        <td>
          External enables you to configure external grafana instances that is not managed by the operator.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproute">httpRoute</a></b></td>
        <td>object</td>
        <td>
          HTTPRoute sets how the ingress object should look like with your grafana instance, this only works use gateway api.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingress">ingress</a></b></td>
        <td>object</td>
        <td>
          Ingress sets how the ingress object should look like with your grafana instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecjsonnet">jsonnet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaim">persistentVolumeClaim</a></b></td>
        <td>object</td>
        <td>
          PersistentVolumeClaim creates a PVC if you need to attach one to your grafana instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpreferences">preferences</a></b></td>
        <td>object</td>
        <td>
          Preferences holds the Grafana Preferences settings<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroute">route</a></b></td>
        <td>object</td>
        <td>
          Route sets how the ingress object should look like with your grafana instance, this only works in Openshift.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecservice">service</a></b></td>
        <td>object</td>
        <td>
          Service sets how the service object should look like with your grafana instance, contains a number of defaults.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecserviceaccount">serviceAccount</a></b></td>
        <td>object</td>
        <td>
          ServiceAccount sets how the ServiceAccount object should look like with your grafana instance, contains a number of defaults.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses reconciliation of owned resources like deployments, Services, Etc. upon changes<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          Version sets the tag of the default image: docker.io/grafana/grafana.
Allows full image refs with/without sha256checksum: "registry/repo/image:tag@sha"
default: 12.2.1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.client
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Client defines how the grafana-operator talks to the grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>headers</b></td>
        <td>map[string]string</td>
        <td>
          Custom HTTP headers to use when interacting with this Grafana.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>preferIngress</b></td>
        <td>boolean</td>
        <td>
          If the operator should send it's request through the grafana instances ingress object instead of through the service.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeout</b></td>
        <td>integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecclienttls">tls</a></b></td>
        <td>object</td>
        <td>
          TLS Configuration used to talk with the grafana instance.<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.insecureSkipVerify) && !(has(self.certSecretRef))) || (has(self.certSecretRef) && !(has(self.insecureSkipVerify))): insecureSkipVerify and certSecretRef cannot be set at the same time</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>useKubeAuth</b></td>
        <td>boolean</td>
        <td>
          Use Kubernetes Serviceaccount as authentication
Requires configuring [auth.jwt] in the instance<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.client.tls
<sup><sup>[↩ Parent](#grafanaspecclient)</sup></sup>



TLS Configuration used to talk with the grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecclienttlscertsecretref">certSecretRef</a></b></td>
        <td>object</td>
        <td>
          Use a secret as a reference to give TLS Certificate information<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable the CA check of the server<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.client.tls.certSecretRef
<sup><sup>[↩ Parent](#grafanaspecclienttls)</sup></sup>



Use a secret as a reference to give TLS Certificate information

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is unique within a namespace to reference a secret resource.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          namespace defines the space within which the secret name must be unique.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Deployment sets how the deployment object should look like with your grafana instance, contains a number of defaults.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.metadata
<sup><sup>[↩ Parent](#grafanaspecdeployment)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec
<sup><sup>[↩ Parent](#grafanaspecdeployment)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>minReadySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>paused</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>progressDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replicas</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>revisionHistoryLimit</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspecstrategy">strategy</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplate">template</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.selector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspecselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.strategy
<sup><sup>[↩ Parent](#grafanaspecdeploymentspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspecstrategyrollingupdate">rollingUpdate</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.strategy.rollingUpdate
<sup><sup>[↩ Parent](#grafanaspecdeploymentspecstrategy)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>maxSurge</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxUnavailable</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template
<sup><sup>[↩ Parent](#grafanaspecdeploymentspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.metadata
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplate)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplate)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>activeDeadlineSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinity">affinity</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>automountServiceAccountToken</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindex">containers</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecdnsconfig">dnsConfig</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dnsPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>enableServiceLinks</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindex">ephemeralContainers</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespechostaliasesindex">hostAliases</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostIPC</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostNetwork</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostPID</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostUsers</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecimagepullsecretsindex">imagePullSecrets</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindex">initContainers</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeSelector</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecos">os</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>overhead</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>preemptionPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priority</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>priorityClassName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecreadinessgatesindex">readinessGates</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runtimeClassName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>schedulerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontext">securityContext</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serviceAccount</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serviceAccountName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>setHostnameAsFQDN</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>shareProcessNamespace</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subdomain</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespectolerationsindex">tolerations</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespectopologyspreadconstraintsindex">topologySpreadConstraints</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindex">volumes</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinity">nodeAffinity</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinity">podAffinity</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinity">podAntiAffinity</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecution">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference">preference</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreferencematchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreferencematchfieldsindex">matchFields</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].preference.matchFields[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinitypreferredduringschedulingignoredduringexecutionindexpreference)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex">nodeSelectorTerms</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecution)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindexmatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindexmatchfieldsindex">matchFields</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index].matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[index].matchFields[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitynodeaffinityrequiredduringschedulingignoredduringexecutionnodeselectortermsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm">podAffinityTerm</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindex">preferredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex">requiredDuringSchedulingIgnoredDuringExecution</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm">podAffinityTerm</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermlabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinityterm)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.preferredDuringSchedulingIgnoredDuringExecution[index].podAffinityTerm.namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinitypreferredduringschedulingignoredduringexecutionindexpodaffinitytermnamespaceselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinity)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mismatchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector">namespaceSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespaces</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexlabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.affinity.podAntiAffinity.requiredDuringSchedulingIgnoredDuringExecution[index].namespaceSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecaffinitypodantiaffinityrequiredduringschedulingignoredduringexecutionindexnamespaceselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycle">lifecycle</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobe">livenessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobe">readinessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexresizepolicyindex">resizePolicy</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexresources">resources</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexrestartpolicyrulesindex">restartPolicyRules</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext">securityContext</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobe">startupProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdin</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdinOnce</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tty</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexvolumedevicesindex">volumeDevices</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexvolumemountsindex">volumeMounts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>workingDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom.fileKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].envFrom[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvfromindexconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>prefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexenvfromindexsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].envFrom[index].configMapRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].envFrom[index].secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststart">postStart</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestop">preStop</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stopSignal</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststartexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststarthttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststartsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststarttcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststarthttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststarthttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.postStart.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestopexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestophttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestopsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestoptcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestophttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestophttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].lifecycle.preStop.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].livenessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].ports[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>containerPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>hostIP</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: TCP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].readinessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].resizePolicy[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resourceName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].resources
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].resources.claims[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexresources)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>request</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].restartPolicyRules[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexrestartpolicyrulesindexexitcodes">exitCodes</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].restartPolicyRules[index].exitCodes
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexrestartpolicyrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowPrivilegeEscalation</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontextapparmorprofile">appArmorProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontextcapabilities">capabilities</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>procMount</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnlyRootFilesystem</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsGroup</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsNonRoot</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUser</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontextselinuxoptions">seLinuxOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontextseccompprofile">seccompProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontextwindowsoptions">windowsOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext.appArmorProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext.capabilities
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>add</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>drop</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext.seLinuxOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext.seccompProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].securityContext.windowsOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gmsaCredentialSpec</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gmsaCredentialSpecName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostProcess</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUserName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].startupProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].volumeDevices[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>devicePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.containers[index].volumeMounts[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespeccontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPropagation</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>recursiveReadOnly</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPathExpr</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.dnsConfig
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>nameservers</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecdnsconfigoptionsindex">options</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>searches</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.dnsConfig.options[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecdnsconfig)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycle">lifecycle</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobe">livenessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobe">readinessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexresizepolicyindex">resizePolicy</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexresources">resources</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexrestartpolicyrulesindex">restartPolicyRules</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext">securityContext</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobe">startupProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdin</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdinOnce</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>targetContainerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tty</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexvolumedevicesindex">volumeDevices</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexvolumemountsindex">volumeMounts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>workingDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom.fileKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].envFrom[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvfromindexconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>prefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvfromindexsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].envFrom[index].configMapRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].envFrom[index].secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststart">postStart</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestop">preStop</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stopSignal</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststartexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststarthttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststartsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststarttcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststarthttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststarthttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.postStart.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestopexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestophttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestopsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestoptcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestophttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestophttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].lifecycle.preStop.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].livenessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].ports[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>containerPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>hostIP</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: TCP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].readinessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].resizePolicy[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resourceName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].resources
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].resources.claims[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexresources)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>request</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].restartPolicyRules[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexrestartpolicyrulesindexexitcodes">exitCodes</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].restartPolicyRules[index].exitCodes
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexrestartpolicyrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowPrivilegeEscalation</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontextapparmorprofile">appArmorProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontextcapabilities">capabilities</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>procMount</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnlyRootFilesystem</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsGroup</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsNonRoot</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUser</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontextselinuxoptions">seLinuxOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontextseccompprofile">seccompProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontextwindowsoptions">windowsOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext.appArmorProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext.capabilities
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>add</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>drop</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext.seLinuxOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext.seccompProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].securityContext.windowsOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gmsaCredentialSpec</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gmsaCredentialSpecName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostProcess</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUserName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].startupProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].volumeDevices[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>devicePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.ephemeralContainers[index].volumeMounts[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecephemeralcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPropagation</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>recursiveReadOnly</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPathExpr</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.hostAliases[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>ip</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>hostnames</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.imagePullSecrets[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>args</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindex">env</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvfromindex">envFrom</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>imagePullPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycle">lifecycle</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobe">livenessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobe">readinessProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexresizepolicyindex">resizePolicy</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexresources">resources</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexrestartpolicyrulesindex">restartPolicyRules</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext">securityContext</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobe">startupProbe</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdin</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stdinOnce</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationMessagePolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tty</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexvolumedevicesindex">volumeDevices</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexvolumemountsindex">volumeMounts</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>workingDir</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom">valueFrom</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefromconfigmapkeyref">configMapKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefromfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefromfilekeyref">fileKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefromresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefromsecretkeyref">secretKeyRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom.configMapKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom.fieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom.fileKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom.resourceFieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].env[index].valueFrom.secretKeyRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvindexvaluefrom)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].envFrom[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvfromindexconfigmapref">configMapRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>prefix</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexenvfromindexsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].envFrom[index].configMapRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].envFrom[index].secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexenvfromindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststart">postStart</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestop">preStop</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stopSignal</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststartexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststarthttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststartsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststarttcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststarthttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststarthttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.postStart.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecyclepoststart)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestopexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestophttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestopsleep">sleep</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestoptcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestophttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestophttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop.sleep
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>seconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].lifecycle.preStop.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlifecycleprestop)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].livenessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexlivenessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].ports[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>containerPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>hostIP</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostPort</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: TCP<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].readinessProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexreadinessprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].resizePolicy[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resourceName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>restartPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].resources
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].resources.claims[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexresources)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>request</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].restartPolicyRules[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>action</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexrestartpolicyrulesindexexitcodes">exitCodes</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].restartPolicyRules[index].exitCodes
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexrestartpolicyrulesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowPrivilegeEscalation</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontextapparmorprofile">appArmorProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontextcapabilities">capabilities</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>privileged</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>procMount</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnlyRootFilesystem</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsGroup</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsNonRoot</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUser</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontextselinuxoptions">seLinuxOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontextseccompprofile">seccompProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontextwindowsoptions">windowsOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext.appArmorProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext.capabilities
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>add</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>drop</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext.seLinuxOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext.seccompProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].securityContext.windowsOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gmsaCredentialSpec</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gmsaCredentialSpecName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostProcess</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUserName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobeexec">exec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>failureThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobegrpc">grpc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobehttpget">httpGet</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initialDelaySeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>periodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>successThreshold</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobetcpsocket">tcpSocket</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>terminationGracePeriodSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe.exec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>command</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe.grpc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>service</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe.httpGet
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobehttpgethttpheadersindex">httpHeaders</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe.httpGet.httpHeaders[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobehttpget)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].startupProbe.tcpSocket
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindexstartupprobe)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].volumeDevices[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>devicePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.initContainers[index].volumeMounts[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecinitcontainersindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>mountPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mountPropagation</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>recursiveReadOnly</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subPathExpr</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.os
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.readinessGates[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>conditionType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontextapparmorprofile">appArmorProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsGroup</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsGroupChangePolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsGroup</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsNonRoot</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUser</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>seLinuxChangePolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontextselinuxoptions">seLinuxOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontextseccompprofile">seccompProfile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>supplementalGroups</b></td>
        <td>[]integer</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>supplementalGroupsPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontextsysctlsindex">sysctls</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecsecuritycontextwindowsoptions">windowsOptions</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext.appArmorProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext.seLinuxOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>level</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext.seccompProfile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>localhostProfile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext.sysctls[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.securityContext.windowsOptions
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecsecuritycontext)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gmsaCredentialSpec</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>gmsaCredentialSpecName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>hostProcess</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>runAsUserName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.tolerations[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>effect</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tolerationSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.topologySpreadConstraints[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>maxSkew</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>topologyKey</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>whenUnsatisfiable</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespectopologyspreadconstraintsindexlabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabelKeys</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>minDomains</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeAffinityPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodeTaintsPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.topologySpreadConstraints[index].labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespectopologyspreadconstraintsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespectopologyspreadconstraintsindexlabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.topologySpreadConstraints[index].labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespectopologyspreadconstraintsindexlabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexawselasticblockstore">awsElasticBlockStore</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexazuredisk">azureDisk</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexazurefile">azureFile</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcephfs">cephfs</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcinder">cinder</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcsi">csi</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexemptydir">emptyDir</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeral">ephemeral</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexfc">fc</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexflexvolume">flexVolume</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexflocker">flocker</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexgcepersistentdisk">gcePersistentDisk</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexgitrepo">gitRepo</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexglusterfs">glusterfs</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexhostpath">hostPath</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindeximage">image</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexiscsi">iscsi</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexnfs">nfs</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexpersistentvolumeclaim">persistentVolumeClaim</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexphotonpersistentdisk">photonPersistentDisk</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexportworxvolume">portworxVolume</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojected">projected</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexquobyte">quobyte</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexrbd">rbd</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexscaleio">scaleIO</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexstorageos">storageos</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexvspherevolume">vsphereVolume</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].awsElasticBlockStore
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].azureDisk
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>diskName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>diskURI</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>cachingMode</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: ext4<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].azureFile
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>shareName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].cephfs
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretFile</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcephfssecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].cephfs.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexcephfs)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].cinder
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcindersecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].cinder.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexcinder)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].configMap
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].configMap.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexconfigmap)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].csi
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexcsinodepublishsecretref">nodePublishSecretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributes</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].csi.nodePublishSecretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexcsi)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].downwardAPI
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapi)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapiitemsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexdownwardapiitemsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].emptyDir
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>medium</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sizeLimit</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplate">volumeClaimTemplate</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeral)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>metadata</b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplate)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessModes</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecdatasource">dataSource</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecdatasourceref">dataSourceRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecresources">resources</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecselector">selector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeAttributesClassName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeMode</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.dataSource
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.dataSourceRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.resources
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.selector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].ephemeral.volumeClaimTemplate.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexephemeralvolumeclaimtemplatespecselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].fc
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>targetWWNs</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>wwids</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].flexVolume
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>driver</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>options</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexflexvolumesecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].flexVolume.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexflexvolume)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].flocker
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>datasetName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>datasetUUID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].gcePersistentDisk
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>partition</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].gitRepo
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>repository</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>directory</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>revision</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].glusterfs
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>endpoints</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].hostPath
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].image
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pullPolicy</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>reference</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].iscsi
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>iqn</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>lun</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>targetPortal</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>chapAuthDiscovery</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>chapAuthSession</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>initiatorName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>iscsiInterface</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: default<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>portals</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexiscsisecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].iscsi.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexiscsi)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].nfs
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>server</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].persistentVolumeClaim
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>claimName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].photonPersistentDisk
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>pdID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].portworxVolume
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumeID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex">sources</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojected)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexclustertrustbundle">clusterTrustBundle</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexconfigmap">configMap</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapi">downwardAPI</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexpodcertificate">podCertificate</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexserviceaccounttoken">serviceAccountToken</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].clusterTrustBundle
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexclustertrustbundlelabelselector">labelSelector</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>signerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].clusterTrustBundle.labelSelector
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexclustertrustbundle)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexclustertrustbundlelabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].clusterTrustBundle.labelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexclustertrustbundlelabelselector)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].configMap
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexconfigmapitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].configMap.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexconfigmap)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].downwardAPI
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapiitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].downwardAPI.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapi)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapiitemsindexfieldref">fieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapiitemsindexresourcefieldref">resourceFieldRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].downwardAPI.items[index].fieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].downwardAPI.items[index].resourceFieldRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexdownwardapiitemsindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>resource</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>containerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>divisor</b></td>
        <td>int or string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].podCertificate
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>keyType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>signerName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>certificateChainPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>credentialBundlePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyPath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxExpirationSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].secret
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].secret.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindexsecret)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].projected.sources[index].serviceAccountToken
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexprojectedsourcesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>audience</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>expirationSeconds</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].quobyte
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>registry</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>volume</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>tenant</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].rbd
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>monitors</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>keyring</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: /etc/ceph/keyring<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>pool</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: rbd<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexrbdsecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>user</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: admin<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].rbd.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexrbd)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].scaleIO
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>gateway</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexscaleiosecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>system</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: xfs<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protectionDomain</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sslEnabled</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageMode</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: ThinProvisioned<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePool</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].scaleIO.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexscaleio)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].secret
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>defaultMode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexsecretitemsindex">items</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].secret.items[index]
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexsecret)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>mode</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].storageos
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>readOnly</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecdeploymentspectemplatespecvolumesindexstorageossecretref">secretRef</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeNamespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].storageos.secretRef
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindexstorageos)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.deployment.spec.template.spec.volumes[index].vsphereVolume
<sup><sup>[↩ Parent](#grafanaspecdeploymentspectemplatespecvolumesindex)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>volumePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fsType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyID</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storagePolicyName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



External enables you to configure external grafana instances that is not managed by the operator.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          URL of the external grafana instance you want to manage.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecexternaladminpassword">adminPassword</a></b></td>
        <td>object</td>
        <td>
          AdminPassword key to talk to the external grafana instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecexternaladminuser">adminUser</a></b></td>
        <td>object</td>
        <td>
          AdminUser key to talk to the external grafana instance.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecexternalapikey">apiKey</a></b></td>
        <td>object</td>
        <td>
          The API key to talk to the external grafana instance, you need to define ether apiKey or adminUser/adminPassword.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecexternaltls">tls</a></b></td>
        <td>object</td>
        <td>
          DEPRECATED, use top level `tls` instead.<br/>
          <br/>
            <i>Validations</i>:<li>(has(self.insecureSkipVerify) && !(has(self.certSecretRef))) || (has(self.certSecretRef) && !(has(self.insecureSkipVerify))): insecureSkipVerify and certSecretRef cannot be set at the same time</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external.adminPassword
<sup><sup>[↩ Parent](#grafanaspecexternal)</sup></sup>



AdminPassword key to talk to the external grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external.adminUser
<sup><sup>[↩ Parent](#grafanaspecexternal)</sup></sup>



AdminUser key to talk to the external grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external.apiKey
<sup><sup>[↩ Parent](#grafanaspecexternal)</sup></sup>



The API key to talk to the external grafana instance, you need to define ether apiKey or adminUser/adminPassword.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          The key of the secret to select from.  Must be a valid secret key.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>optional</b></td>
        <td>boolean</td>
        <td>
          Specify whether the Secret or its key must be defined<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external.tls
<sup><sup>[↩ Parent](#grafanaspecexternal)</sup></sup>



DEPRECATED, use top level `tls` instead.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecexternaltlscertsecretref">certSecretRef</a></b></td>
        <td>object</td>
        <td>
          Use a secret as a reference to give TLS Certificate information<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureSkipVerify</b></td>
        <td>boolean</td>
        <td>
          Disable the CA check of the server<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.external.tls.certSecretRef
<sup><sup>[↩ Parent](#grafanaspecexternaltls)</sup></sup>



Use a secret as a reference to give TLS Certificate information

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is unique within a namespace to reference a secret resource.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          namespace defines the space within which the secret name must be unique.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



HTTPRoute sets how the ingress object should look like with your grafana instance, this only works use gateway api.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespec">spec</a></b></td>
        <td>object</td>
        <td>
          HTTPRouteSpec defines the desired state of HTTPRoute<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.metadata
<sup><sup>[↩ Parent](#grafanaspechttproute)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec
<sup><sup>[↩ Parent](#grafanaspechttproute)</sup></sup>



HTTPRouteSpec defines the desired state of HTTPRoute

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostnames</b></td>
        <td>[]string</td>
        <td>
          Hostnames defines a set of hostnames that should match against the HTTP Host
header to select a HTTPRoute used to process the request. Implementations
MUST ignore any port value specified in the HTTP Host header while
performing a match and (absent of any applicable header modification
configuration) MUST forward this header unmodified to the backend.

Valid values for Hostnames are determined by RFC 1123 definition of a
hostname with 2 notable exceptions:

1. IPs are not allowed.
2. A hostname may be prefixed with a wildcard label (`*.`). The wildcard
   label must appear by itself as the first label.

If a hostname is specified by both the Listener and HTTPRoute, there
must be at least one intersecting hostname for the HTTPRoute to be
attached to the Listener. For example:

* A Listener with `test.example.com` as the hostname matches HTTPRoutes
  that have either not specified any hostnames, or have specified at
  least one of `test.example.com` or `*.example.com`.
* A Listener with `*.example.com` as the hostname matches HTTPRoutes
  that have either not specified any hostnames or have specified at least
  one hostname that matches the Listener hostname. For example,
  `*.example.com`, `test.example.com`, and `foo.test.example.com` would
  all match. On the other hand, `example.com` and `test.example.net` would
  not match.

Hostnames that are prefixed with a wildcard label (`*.`) are interpreted
as a suffix match. That means that a match for `*.example.com` would match
both `test.example.com`, and `foo.test.example.com`, but not `example.com`.

If both the Listener and HTTPRoute have specified hostnames, any
HTTPRoute hostnames that do not match the Listener hostname MUST be
ignored. For example, if a Listener specified `*.example.com`, and the
HTTPRoute specified `test.example.com` and `test.example.net`,
`test.example.net` must not be considered for a match.

If both the Listener and HTTPRoute have specified hostnames, and none
match with the criteria above, then the HTTPRoute is not accepted. The
implementation must raise an 'Accepted' Condition with a status of
`False` in the corresponding RouteParentStatus.

In the event that multiple HTTPRoutes specify intersecting hostnames (e.g.
overlapping wildcard matching and exact matching hostnames), precedence must
be given to rules from the HTTPRoute with the largest number of:

* Characters in a matching non-wildcard hostname.
* Characters in a matching hostname.

If ties exist across multiple Routes, the matching precedence rules for
HTTPRouteMatches takes over.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecparentrefsindex">parentRefs</a></b></td>
        <td>[]object</td>
        <td>
          ParentRefs references the resources (usually Gateways) that a Route wants
to be attached to. Note that the referenced parent resource needs to
allow this for the attachment to be complete. For Gateways, that means
the Gateway needs to allow attachment from Routes of this kind and
namespace. For Services, that means the Service must either be in the same
namespace for a "producer" route, or the mesh implementation must support
and allow "consumer" routes for the referenced Service. ReferenceGrant is
not applicable for governing ParentRefs to Services - it is not possible to
create a "producer" route for a Service in a different namespace from the
Route.

There are two kinds of parent resources with "Core" support:

* Gateway (Gateway conformance profile)
* Service (Mesh conformance profile, ClusterIP Services only)

This API may be extended in the future to support additional kinds of parent
resources.

ParentRefs must be _distinct_. This means either that:

* They select different objects.  If this is the case, then parentRef
  entries are distinct. In terms of fields, this means that the
  multi-part key defined by `group`, `kind`, `namespace`, and `name` must
  be unique across all parentRef entries in the Route.
* They do not select different objects, but for each optional field used,
  each ParentRef that selects the same object must set the same set of
  optional fields to different values. If one ParentRef sets a
  combination of optional fields, all must set the same combination.

Some examples:

* If one ParentRef sets `sectionName`, all ParentRefs referencing the
  same object must also set `sectionName`.
* If one ParentRef sets `port`, all ParentRefs referencing the same
  object must also set `port`.
* If one ParentRef sets `sectionName` and `port`, all ParentRefs
  referencing the same object must also set `sectionName` and `port`.

It is possible to separately reference multiple distinct objects that may
be collapsed by an implementation. For example, some implementations may
choose to merge compatible Gateway Listeners together. If that is the
case, the list of routes attached to those resources should also be
merged.

Note that for ParentRefs that cross namespace boundaries, there are specific
rules. Cross-namespace references are only valid if they are explicitly
allowed by something in the namespace they are referring to. For example,
Gateway has the AllowedRoutes field, and ReferenceGrant provides a
generic way to enable other kinds of cross-namespace reference.

<gateway:experimental:description>
ParentRefs from a Route to a Service in the same namespace are "producer"
routes, which apply default routing rules to inbound connections from
any namespace to the Service.

ParentRefs from a Route to a Service in a different namespace are
"consumer" routes, and these routing rules are only applied to outbound
connections originating from the same namespace as the Route, for which
the intended destination of the connections are a Service targeted as a
ParentRef of the Route.
</gateway:experimental:description>

<gateway:standard:validation:XValidation:message="sectionName must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) || p1.__namespace__ == '') && (!has(p2.__namespace__) || p2.__namespace__ == '')) || (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) ? ((!has(p1.sectionName) || p1.sectionName == '') == (!has(p2.sectionName) || p2.sectionName == '')) : true))">
<gateway:standard:validation:XValidation:message="sectionName must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) || p1.__namespace__ == '') && (!has(p2.__namespace__) || p2.__namespace__ == '')) || (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) || p1.sectionName == '') && (!has(p2.sectionName) || p2.sectionName == '')) || (has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName))))">
<gateway:experimental:validation:XValidation:message="sectionName or port must be specified when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.all(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) || p1.__namespace__ == '') && (!has(p2.__namespace__) || p2.__namespace__ == '')) || (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__)) ? ((!has(p1.sectionName) || p1.sectionName == '') == (!has(p2.sectionName) || p2.sectionName == '') && (!has(p1.port) || p1.port == 0) == (!has(p2.port) || p2.port == 0)): true))">
<gateway:experimental:validation:XValidation:message="sectionName or port must be unique when parentRefs includes 2 or more references to the same parent",rule="self.all(p1, self.exists_one(p2, p1.group == p2.group && p1.kind == p2.kind && p1.name == p2.name && (((!has(p1.__namespace__) || p1.__namespace__ == '') && (!has(p2.__namespace__) || p2.__namespace__ == '')) || (has(p1.__namespace__) && has(p2.__namespace__) && p1.__namespace__ == p2.__namespace__ )) && (((!has(p1.sectionName) || p1.sectionName == '') && (!has(p2.sectionName) || p2.sectionName == '')) || ( has(p1.sectionName) && has(p2.sectionName) && p1.sectionName == p2.sectionName)) && (((!has(p1.port) || p1.port == 0) && (!has(p2.port) || p2.port == 0)) || (has(p1.port) && has(p2.port) && p1.port == p2.port))))"><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          Rules are a list of HTTP matchers, filters and actions.

<gateway:experimental:validation:XValidation:message="Rule name must be unique within the route",rule="self.all(l1, !has(l1.name) || self.exists_one(l2, has(l2.name) && l1.name == l2.name))"><br/>
          <br/>
            <i>Validations</i>:<li>(self.size() > 0 ? self[0].matches.size() : 0) + (self.size() > 1 ? self[1].matches.size() : 0) + (self.size() > 2 ? self[2].matches.size() : 0) + (self.size() > 3 ? self[3].matches.size() : 0) + (self.size() > 4 ? self[4].matches.size() : 0) + (self.size() > 5 ? self[5].matches.size() : 0) + (self.size() > 6 ? self[6].matches.size() : 0) + (self.size() > 7 ? self[7].matches.size() : 0) + (self.size() > 8 ? self[8].matches.size() : 0) + (self.size() > 9 ? self[9].matches.size() : 0) + (self.size() > 10 ? self[10].matches.size() : 0) + (self.size() > 11 ? self[11].matches.size() : 0) + (self.size() > 12 ? self[12].matches.size() : 0) + (self.size() > 13 ? self[13].matches.size() : 0) + (self.size() > 14 ? self[14].matches.size() : 0) + (self.size() > 15 ? self[15].matches.size() : 0) <= 128: While 16 rules and 64 matches per rule are allowed, the total number of matches across all rules in a route must be less than 128</li>
            <i>Default</i>: [map[matches:[map[path:map[type:PathPrefix value:/]]]]]<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.parentRefs[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespec)</sup></sup>



ParentReference identifies an API object (usually a Gateway) that can be considered
a parent of this resource (usually a route). There are two kinds of parent resources
with "Core" support:

* Gateway (Gateway conformance profile)
* Service (Mesh conformance profile, ClusterIP Services only)

This API may be extended in the future to support additional kinds of parent
resources.

The API object must be valid in the cluster; the Group and Kind must
be registered in the cluster for this reference to be valid.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.

Support: Core<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent.
When unspecified, "gateway.networking.k8s.io" is inferred.
To set the core API group (such as for a "Service" kind referent),
Group must be explicitly set to "" (empty string).

Support: Core<br/>
          <br/>
            <i>Default</i>: gateway.networking.k8s.io<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is kind of the referent.

There are two kinds of parent resources with "Core" support:

* Gateway (Gateway conformance profile)
* Service (Mesh conformance profile, ClusterIP Services only)

Support for other resources is Implementation-Specific.<br/>
          <br/>
            <i>Default</i>: Gateway<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of the referent. When unspecified, this refers
to the local namespace of the Route.

Note that there are specific rules for ParentRefs which cross namespace
boundaries. Cross-namespace references are only valid if they are explicitly
allowed by something in the namespace they are referring to. For example:
Gateway has the AllowedRoutes field, and ReferenceGrant provides a
generic way to enable any other kind of cross-namespace reference.

<gateway:experimental:description>
ParentRefs from a Route to a Service in the same namespace are "producer"
routes, which apply default routing rules to inbound connections from
any namespace to the Service.

ParentRefs from a Route to a Service in a different namespace are
"consumer" routes, and these routing rules are only applied to outbound
connections originating from the same namespace as the Route, for which
the intended destination of the connections are a Service targeted as a
ParentRef of the Route.
</gateway:experimental:description>

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the network port this Route targets. It can be interpreted
differently based on the type of parent resource.

When the parent resource is a Gateway, this targets all listeners
listening on the specified port that also support this kind of Route(and
select this Route). It's not recommended to set `Port` unless the
networking behaviors specified in a Route must apply to a specific port
as opposed to a listener(s) whose port(s) may be changed. When both Port
and SectionName are specified, the name and port of the selected listener
must match both specified values.

<gateway:experimental:description>
When the parent resource is a Service, this targets a specific port in the
Service spec. When both Port (experimental) and SectionName are specified,
the name and port of the selected port must match both specified values.
</gateway:experimental:description>

Implementations MAY choose to support other parent resources.
Implementations supporting other types of parent resources MUST clearly
document how/if Port is interpreted.

For the purpose of status, an attachment is considered successful as
long as the parent resource accepts it partially. For example, Gateway
listeners can restrict which Routes can attach to them by Route kind,
namespace, or hostname. If 1 of 2 Gateway listeners accept attachment
from the referencing Route, the Route MUST be considered successfully
attached. If no Gateway listeners accept attachment from this Route,
the Route MUST be considered detached from the Gateway.

Support: Extended<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sectionName</b></td>
        <td>string</td>
        <td>
          SectionName is the name of a section within the target resource. In the
following resources, SectionName is interpreted as the following:

* Gateway: Listener name. When both Port (experimental) and SectionName
are specified, the name and port of the selected listener must match
both specified values.
* Service: Port name. When both Port (experimental) and SectionName
are specified, the name and port of the selected listener must match
both specified values.

Implementations MAY choose to support attaching Routes to other resources.
If that is the case, they MUST clearly document how SectionName is
interpreted.

When unspecified (empty string), this will reference the entire resource.
For the purpose of status, an attachment is considered successful if at
least one section in the parent resource accepts it. For example, Gateway
listeners can restrict which Routes can attach to them by Route kind,
namespace, or hostname. If 1 of 2 Gateway listeners accept attachment from
the referencing Route, the Route MUST be considered successfully
attached. If no Gateway listeners accept attachment from this Route, the
Route MUST be considered detached from the Gateway.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespec)</sup></sup>



HTTPRouteRule defines semantics for matching an HTTP request based on
conditions (matches), processing it (filters), and forwarding the request to
an API object (backendRefs).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindex">backendRefs</a></b></td>
        <td>[]object</td>
        <td>
          BackendRefs defines the backend(s) where matching requests should be
sent.

Failure behavior here depends on how many BackendRefs are specified and
how many are invalid.

If *all* entries in BackendRefs are invalid, and there are also no filters
specified in this route rule, *all* traffic which matches this rule MUST
receive a 500 status code.

See the HTTPBackendRef definition for the rules about what makes a single
HTTPBackendRef invalid.

When a HTTPBackendRef is invalid, 500 status codes MUST be returned for
requests that would have otherwise been routed to an invalid backend. If
multiple backends are specified, and some are invalid, the proportion of
requests that would otherwise have been routed to an invalid backend
MUST receive a 500 status code.

For example, if two backends are specified with equal weights, and one is
invalid, 50 percent of traffic must receive a 500. Implementations may
choose how that 50 percent is determined.

When a HTTPBackendRef refers to a Service that has no ready endpoints,
implementations SHOULD return a 503 for requests to that backend instead.
If an implementation chooses to do this, all of the above rules for 500 responses
MUST also apply for responses that return a 503.

Support: Core for Kubernetes Service

Support: Extended for Kubernetes ServiceImport

Support: Implementation-specific for any other resource

Support for weight: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindex">filters</a></b></td>
        <td>[]object</td>
        <td>
          Filters define the filters that are applied to requests that match
this rule.

Wherever possible, implementations SHOULD implement filters in the order
they are specified.

Implementations MAY choose to implement this ordering strictly, rejecting
any combination or order of filters that cannot be supported. If implementations
choose a strict interpretation of filter ordering, they MUST clearly document
that behavior.

To reject an invalid combination or order of filters, implementations SHOULD
consider the Route Rules with this configuration invalid. If all Route Rules
in a Route are invalid, the entire Route would be considered invalid. If only
a portion of Route Rules are invalid, implementations MUST set the
"PartiallyInvalid" condition for the Route.

Conformance-levels at this level are defined based on the type of filter:

- ALL core filters MUST be supported by all implementations.
- Implementers are encouraged to support extended filters.
- Implementation-specific custom filters have no API guarantees across
  implementations.

Specifying the same filter multiple times is not supported unless explicitly
indicated in the filter.

All filters are expected to be compatible with each other except for the
URLRewrite and RequestRedirect filters, which may not be combined. If an
implementation cannot support other combinations of filters, they must clearly
document that limitation. In cases where incompatible or unsupported
filters are specified and cause the `Accepted` condition to be set to status
`False`, implementations may use the `IncompatibleFilters` reason to specify
this configuration error.

Support: Core<br/>
          <br/>
            <i>Validations</i>:<li>!(self.exists(f, f.type == 'RequestRedirect') && self.exists(f, f.type == 'URLRewrite')): May specify either httpRouteFilterRequestRedirect or httpRouteFilterRequestRewrite, but not both</li><li>self.filter(f, f.type == 'RequestHeaderModifier').size() <= 1: RequestHeaderModifier filter cannot be repeated</li><li>self.filter(f, f.type == 'ResponseHeaderModifier').size() <= 1: ResponseHeaderModifier filter cannot be repeated</li><li>self.filter(f, f.type == 'RequestRedirect').size() <= 1: RequestRedirect filter cannot be repeated</li><li>self.filter(f, f.type == 'URLRewrite').size() <= 1: URLRewrite filter cannot be repeated</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexmatchesindex">matches</a></b></td>
        <td>[]object</td>
        <td>
          Matches define conditions used for matching the rule against incoming
HTTP requests. Each match is independent, i.e. this rule will be matched
if **any** one of the matches is satisfied.

For example, take the following matches configuration:

```
matches:
- path:
    value: "/foo"
  headers:
  - name: "version"
    value: "v2"
- path:
    value: "/v2/foo"
```

For a request to match against this rule, a request must satisfy
EITHER of the two conditions:

- path prefixed with `/foo` AND contains the header `version: v2`
- path prefix of `/v2/foo`

See the documentation for HTTPRouteMatch on how to specify multiple
match conditions that should be ANDed together.

If no matches are specified, the default is a prefix
path match on "/", which has the effect of matching every
HTTP request.

Proxy or Load Balancer routing configuration generated from HTTPRoutes
MUST prioritize matches based on the following criteria, continuing on
ties. Across all rules specified on applicable Routes, precedence must be
given to the match having:

* "Exact" path match.
* "Prefix" path match with largest number of characters.
* Method match.
* Largest number of header matches.
* Largest number of query param matches.

Note: The precedence of RegularExpression path matches are implementation-specific.

If ties still exist across multiple Routes, matching precedence MUST be
determined in order of the following criteria, continuing on ties:

* The oldest Route based on creation timestamp.
* The Route appearing first in alphabetical order by
  "{namespace}/{name}".

If ties still exist within an HTTPRoute, matching precedence MUST be granted
to the FIRST matching rule (in list order) with a match meeting the above
criteria.

When no rules matching a request have been successfully attached to the
parent a request is coming from, a HTTP 404 status code MUST be returned.<br/>
          <br/>
            <i>Default</i>: [map[path:map[type:PathPrefix value:/]]]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the route rule. This name MUST be unique within a Route if it is set.

Support: Extended
<gateway:experimental><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexretry">retry</a></b></td>
        <td>object</td>
        <td>
          Retry defines the configuration for when to retry an HTTP request.

Support: Extended

<gateway:experimental><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexsessionpersistence">sessionPersistence</a></b></td>
        <td>object</td>
        <td>
          SessionPersistence defines and configures session persistence
for the route rule.

Support: Extended

<gateway:experimental><br/>
          <br/>
            <i>Validations</i>:<li>!has(self.cookieConfig) || !has(self.cookieConfig.lifetimeType) || self.cookieConfig.lifetimeType != 'Permanent' || has(self.absoluteTimeout): AbsoluteTimeout must be specified when cookie lifetimeType is Permanent</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindextimeouts">timeouts</a></b></td>
        <td>object</td>
        <td>
          Timeouts defines the timeouts that can be configured for an HTTP request.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>!(has(self.request) && has(self.backendRequest) && duration(self.request) != duration('0s') && duration(self.backendRequest) > duration(self.request)): backendRequest timeout cannot be longer than request timeout</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



HTTPBackendRef defines how a HTTPRoute forwards a HTTP request.

Note that when a namespace different than the local namespace is specified, a
ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

<gateway:experimental:description>

When the BackendRef points to a Kubernetes Service, implementations SHOULD
honor the appProtocol field if it is set for the target Service Port.

Implementations supporting appProtocol SHOULD recognize the Kubernetes
Standard Application Protocols defined in KEP-3726.

If a Service appProtocol isn't specified, an implementation MAY infer the
backend protocol through its own means. Implementations MAY infer the
protocol from the Route type referring to the backend Service.

If a Route is not able to send traffic to the backend using the specified
protocol then the backend is considered invalid. Implementations MUST set the
"ResolvedRefs" condition to "False" with the "UnsupportedProtocol" reason.

</gateway:experimental:description>

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex">filters</a></b></td>
        <td>[]object</td>
        <td>
          Filters defined at this level should be executed if and only if the
request is being forwarded to the backend defined here.

Support: Implementation-specific (For broader support of filters, use the
Filters field in HTTPRouteRule.)<br/>
          <br/>
            <i>Validations</i>:<li>!(self.exists(f, f.type == 'RequestRedirect') && self.exists(f, f.type == 'URLRewrite')): May specify either httpRouteFilterRequestRedirect or httpRouteFilterRequestRewrite, but not both</li><li>!(self.exists(f, f.type == 'RequestRedirect') && self.exists(f, f.type == 'URLRewrite')): May specify either httpRouteFilterRequestRedirect or httpRouteFilterRequestRewrite, but not both</li><li>self.filter(f, f.type == 'RequestHeaderModifier').size() <= 1: RequestHeaderModifier filter cannot be repeated</li><li>self.filter(f, f.type == 'ResponseHeaderModifier').size() <= 1: ResponseHeaderModifier filter cannot be repeated</li><li>self.filter(f, f.type == 'RequestRedirect').size() <= 1: RequestRedirect filter cannot be repeated</li><li>self.filter(f, f.type == 'URLRewrite').size() <= 1: URLRewrite filter cannot be repeated</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent. For example, "gateway.networking.k8s.io".
When unspecified or empty string, core API group is inferred.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the Kubernetes resource kind of the referent. For example
"Service".

Defaults to "Service" when not specified.

ExternalName services can refer to CNAME DNS records that may live
outside of the cluster and as such are difficult to reason about in
terms of conformance. They also may not be safe to forward to (see
CVE-2021-25740 for more information). Implementations SHOULD NOT
support ExternalName Services.

Support: Core (Services with a type other than ExternalName)

Support: Implementation-specific (Services with type ExternalName)<br/>
          <br/>
            <i>Default</i>: Service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of the backend. When unspecified, the local
namespace is inferred.

Note that when a namespace different than the local namespace is specified,
a ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port specifies the destination port number to use for this resource.
Port is required when the referent is a Kubernetes Service. In this
case, the port number is the service port number, not the target port.
For other resources, destination port might be derived from the referent
resource or this field.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          Weight specifies the proportion of requests forwarded to the referenced
backend. This is computed as weight/(sum of all weights in this
BackendRefs list). For non-zero values, there may be some epsilon from
the exact proportion defined here depending on the precision an
implementation supports. Weight is not a percentage and the sum of
weights does not need to equal 100.

If only one backend is specified and it has a weight greater than 0, 100%
of the traffic is forwarded to that backend. If weight is set to 0, no
traffic should be forwarded for this entry. If unspecified, weight
defaults to 1.

Support for this field varies based on the context where used.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 1<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 1e+06<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindex)</sup></sup>



HTTPRouteFilter defines processing steps that must be completed during the
request or response lifecycle. HTTPRouteFilters are meant as an extension
point to express processing that may be done in Gateway implementations. Some
examples include request or response modification, implementing
authentication strategies, rate-limiting, and traffic shaping. API
guarantee/conformance is defined based on the type of the filter.

<gateway:experimental:validation:XValidation:message="filter.cors must be nil if the filter.type is not CORS",rule="!(has(self.cors) && self.type != 'CORS')">
<gateway:experimental:validation:XValidation:message="filter.cors must be specified for CORS filter.type",rule="!(!has(self.cors) && self.type == 'CORS')">

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type identifies the type of filter to apply. As with other API fields,
types are classified into three conformance levels:

- Core: Filter types and their corresponding configuration defined by
  "Support: Core" in this package, e.g. "RequestHeaderModifier". All
  implementations must support core filters.

- Extended: Filter types and their corresponding configuration defined by
  "Support: Extended" in this package, e.g. "RequestMirror". Implementers
  are encouraged to support extended filters.

- Implementation-specific: Filters that are defined and supported by
  specific vendors.
  In the future, filters showing convergence in behavior across multiple
  implementations will be considered for inclusion in extended or core
  conformance levels. Filter-specific configuration for such filters
  is specified using the ExtensionRef field. `Type` should be set to
  "ExtensionRef" for custom filters.

Implementers are encouraged to define custom implementation types to
extend the core API with implementation-specific behavior.

If a reference to a custom filter type cannot be resolved, the filter
MUST NOT be skipped. Instead, requests that would have been processed by
that filter MUST receive a HTTP error response.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

<gateway:experimental:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;RequestMirror;RequestRedirect;URLRewrite;ExtensionRef;CORS><br/>
          <br/>
            <i>Enum</i>: RequestHeaderModifier, ResponseHeaderModifier, RequestMirror, RequestRedirect, URLRewrite, ExtensionRef<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexcors">cors</a></b></td>
        <td>object</td>
        <td>
          CORS defines a schema for a filter that responds to the
cross-origin request based on HTTP response header.

Support: Extended

<gateway:experimental><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexextensionref">extensionRef</a></b></td>
        <td>object</td>
        <td>
          ExtensionRef is an optional, implementation-specific extension to the
"filter" behavior.  For example, resource "myroutefilter" in group
"networking.example.net"). ExtensionRef MUST NOT be used for core and
extended filters.

This filter can be used multiple times within the same rule.

Support: Implementation-specific<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestheadermodifier">requestHeaderModifier</a></b></td>
        <td>object</td>
        <td>
          RequestHeaderModifier defines a schema for a filter that modifies request
headers.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestmirror">requestMirror</a></b></td>
        <td>object</td>
        <td>
          RequestMirror defines a schema for a filter that mirrors requests.
Requests are sent to the specified destination, but responses from
that destination are ignored.

This filter can be used multiple times within the same rule. Note that
not all implementations will be able to support mirroring to multiple
backends.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>!(has(self.percent) && has(self.fraction)): Only one of percent or fraction may be specified in HTTPRequestMirrorFilter</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestredirect">requestRedirect</a></b></td>
        <td>object</td>
        <td>
          RequestRedirect defines a schema for a filter that responds to the
request with an HTTP redirection.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexresponseheadermodifier">responseHeaderModifier</a></b></td>
        <td>object</td>
        <td>
          ResponseHeaderModifier defines a schema for a filter that modifies response
headers.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexurlrewrite">urlRewrite</a></b></td>
        <td>object</td>
        <td>
          URLRewrite defines a schema for a filter that modifies a request during forwarding.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].cors
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



CORS defines a schema for a filter that responds to the
cross-origin request based on HTTP response header.

Support: Extended

<gateway:experimental>

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowCredentials</b></td>
        <td>boolean</td>
        <td>
          AllowCredentials indicates whether the actual cross-origin request allows
to include credentials.

The only valid value for the `Access-Control-Allow-Credentials` response
header is true (case-sensitive).

If the credentials are not allowed in cross-origin requests, the gateway
will omit the header `Access-Control-Allow-Credentials` entirely rather
than setting its value to false.

Support: Extended<br/>
          <br/>
            <i>Enum</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowHeaders</b></td>
        <td>[]string</td>
        <td>
          AllowHeaders indicates which HTTP request headers are supported for
accessing the requested resource.

Header names are not case sensitive.

Multiple header names in the value of the `Access-Control-Allow-Headers`
response header are separated by a comma (",").

When the `AllowHeaders` field is configured with one or more headers, the
gateway must return the `Access-Control-Allow-Headers` response header
which value is present in the `AllowHeaders` field.

If any header name in the `Access-Control-Request-Headers` request header
is not included in the list of header names specified by the response
header `Access-Control-Allow-Headers`, it will present an error on the
client side.

If any header name in the `Access-Control-Allow-Headers` response header
does not recognize by the client, it will also occur an error on the
client side.

A wildcard indicates that the requests with all HTTP headers are allowed.
The `Access-Control-Allow-Headers` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowHeaders` field
specified with the `*` wildcard, the gateway must specify one or more
HTTP headers in the value of the `Access-Control-Allow-Headers` response
header. The value of the header `Access-Control-Allow-Headers` is same as
the `Access-Control-Request-Headers` header provided by the client. If
the header `Access-Control-Request-Headers` is not included in the
request, the gateway will omit the `Access-Control-Allow-Headers`
response header, instead of specifying the `*` wildcard. A Gateway
implementation may choose to add implementation-specific default headers.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowMethods</b></td>
        <td>[]enum</td>
        <td>
          AllowMethods indicates which HTTP methods are supported for accessing the
requested resource.

Valid values are any method defined by RFC9110, along with the special
value `*`, which represents all HTTP methods are allowed.

Method names are case sensitive, so these values are also case-sensitive.
(See https://www.rfc-editor.org/rfc/rfc2616#section-5.1.1)

Multiple method names in the value of the `Access-Control-Allow-Methods`
response header are separated by a comma (",").

A CORS-safelisted method is a method that is `GET`, `HEAD`, or `POST`.
(See https://fetch.spec.whatwg.org/#cors-safelisted-method) The
CORS-safelisted methods are always allowed, regardless of whether they
are specified in the `AllowMethods` field.

When the `AllowMethods` field is configured with one or more methods, the
gateway must return the `Access-Control-Allow-Methods` response header
which value is present in the `AllowMethods` field.

If the HTTP method of the `Access-Control-Request-Method` request header
is not included in the list of methods specified by the response header
`Access-Control-Allow-Methods`, it will present an error on the client
side.

The `Access-Control-Allow-Methods` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowMethods` field
specified with the `*` wildcard, the gateway must specify one HTTP method
in the value of the Access-Control-Allow-Methods response header. The
value of the header `Access-Control-Allow-Methods` is same as the
`Access-Control-Request-Method` header provided by the client. If the
header `Access-Control-Request-Method` is not included in the request,
the gateway will omit the `Access-Control-Allow-Methods` response header,
instead of specifying the `*` wildcard. A Gateway implementation may
choose to add implementation-specific default methods.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>!('*' in self && self.size() > 1): AllowMethods cannot contain '*' alongside other methods</li>
            <i>Enum</i>: GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH, *<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowOrigins</b></td>
        <td>[]string</td>
        <td>
          AllowOrigins indicates whether the response can be shared with requested
resource from the given `Origin`.

The `Origin` consists of a scheme and a host, with an optional port, and
takes the form `<scheme>://<host>(:<port>)`.

Valid values for scheme are: `http` and `https`.

Valid values for port are any integer between 1 and 65535 (the list of
available TCP/UDP ports). Note that, if not included, port `80` is
assumed for `http` scheme origins, and port `443` is assumed for `https`
origins. This may affect origin matching.

The host part of the origin may contain the wildcard character `*`. These
wildcard characters behave as follows:

* `*` is a greedy match to the _left_, including any number of
  DNS labels to the left of its position. This also means that
  `*` will include any number of period `.` characters to the
  left of its position.
* A wildcard by itself matches all hosts.

An origin value that includes _only_ the `*` character indicates requests
from all `Origin`s are allowed.

When the `AllowOrigins` field is configured with multiple origins, it
means the server supports clients from multiple origins. If the request
`Origin` matches the configured allowed origins, the gateway must return
the given `Origin` and sets value of the header
`Access-Control-Allow-Origin` same as the `Origin` header provided by the
client.

The status code of a successful response to a "preflight" request is
always an OK status (i.e., 204 or 200).

If the request `Origin` does not match the configured allowed origins,
the gateway returns 204/200 response but doesn't set the relevant
cross-origin response headers. Alternatively, the gateway responds with
403 status to the "preflight" request is denied, coupled with omitting
the CORS headers. The cross-origin request fails on the client side.
Therefore, the client doesn't attempt the actual cross-origin request.

The `Access-Control-Allow-Origin` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowOrigins` field
specified with the `*` wildcard, the gateway must return a single origin
in the value of the `Access-Control-Allow-Origin` response header,
instead of specifying the `*` wildcard. The value of the header
`Access-Control-Allow-Origin` is same as the `Origin` header provided by
the client.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>exposeHeaders</b></td>
        <td>[]string</td>
        <td>
          ExposeHeaders indicates which HTTP response headers can be exposed
to client-side scripts in response to a cross-origin request.

A CORS-safelisted response header is an HTTP header in a CORS response
that it is considered safe to expose to the client scripts.
The CORS-safelisted response headers include the following headers:
`Cache-Control`
`Content-Language`
`Content-Length`
`Content-Type`
`Expires`
`Last-Modified`
`Pragma`
(See https://fetch.spec.whatwg.org/#cors-safelisted-response-header-name)
The CORS-safelisted response headers are exposed to client by default.

When an HTTP header name is specified using the `ExposeHeaders` field,
this additional header will be exposed as part of the response to the
client.

Header names are not case sensitive.

Multiple header names in the value of the `Access-Control-Expose-Headers`
response header are separated by a comma (",").

A wildcard indicates that the responses with all HTTP headers are exposed
to clients. The `Access-Control-Expose-Headers` response header can only
use `*` wildcard as value when the `AllowCredentials` field is
unspecified.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxAge</b></td>
        <td>integer</td>
        <td>
          MaxAge indicates the duration (in seconds) for the client to cache the
results of a "preflight" request.

The information provided by the `Access-Control-Allow-Methods` and
`Access-Control-Allow-Headers` response headers can be cached by the
client until the time specified by `Access-Control-Max-Age` elapses.

The default value of `Access-Control-Max-Age` response header is 5
(seconds).<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 5<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].extensionRef
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



ExtensionRef is an optional, implementation-specific extension to the
"filter" behavior.  For example, resource "myroutefilter" in group
"networking.example.net"). ExtensionRef MUST NOT be used for core and
extended filters.

This filter can be used multiple times within the same rule.

Support: Implementation-specific

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent. For example, "gateway.networking.k8s.io".
When unspecified or empty string, core API group is inferred.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is kind of the referent. For example "HTTPRoute" or "Service".<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestHeaderModifier
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



RequestHeaderModifier defines a schema for a filter that modifies request
headers.

Support: Core

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestheadermodifieraddindex">add</a></b></td>
        <td>[]object</td>
        <td>
          Add adds the given header(s) (name, value) to the request
before the action. It appends to any existing values associated
with the header name.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  add:
  - name: "my-header"
    value: "bar,baz"

Output:
  GET /foo HTTP/1.1
  my-header: foo,bar,baz<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>remove</b></td>
        <td>[]string</td>
        <td>
          Remove the given header(s) from the HTTP request before the action. The
value of Remove is a list of HTTP header names. Note that the header
names are case-insensitive (see
https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).

Input:
  GET /foo HTTP/1.1
  my-header1: foo
  my-header2: bar
  my-header3: baz

Config:
  remove: ["my-header1", "my-header3"]

Output:
  GET /foo HTTP/1.1
  my-header2: bar<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestheadermodifiersetindex">set</a></b></td>
        <td>[]object</td>
        <td>
          Set overwrites the request with the given header (name, value)
before the action.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  set:
  - name: "my-header"
    value: "bar"

Output:
  GET /foo HTTP/1.1
  my-header: bar<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestHeaderModifier.add[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestHeaderModifier.set[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestMirror
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



RequestMirror defines a schema for a filter that mirrors requests.
Requests are sent to the specified destination, but responses from
that destination are ignored.

This filter can be used multiple times within the same rule. Note that
not all implementations will be able to support mirroring to multiple
backends.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestmirrorbackendref">backendRef</a></b></td>
        <td>object</td>
        <td>
          BackendRef references a resource where mirrored requests are sent.

Mirrored requests must be sent only to a single destination endpoint
within this BackendRef, irrespective of how many endpoints are present
within this BackendRef.

If the referent cannot be found, this BackendRef is invalid and must be
dropped from the Gateway. The controller must ensure the "ResolvedRefs"
condition on the Route status is set to `status: False` and not configure
this backend in the underlying implementation.

If there is a cross-namespace reference to an *existing* object
that is not allowed by a ReferenceGrant, the controller must ensure the
"ResolvedRefs"  condition on the Route is set to `status: False`,
with the "RefNotPermitted" reason and not configure this backend in the
underlying implementation.

In either error case, the Message of the `ResolvedRefs` Condition
should be used to provide more detail about the problem.

Support: Extended for Kubernetes Service

Support: Implementation-specific for any other resource<br/>
          <br/>
            <i>Validations</i>:<li>(size(self.group) == 0 && self.kind == 'Service') ? has(self.port) : true: Must have port for Service reference</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestmirrorfraction">fraction</a></b></td>
        <td>object</td>
        <td>
          Fraction represents the fraction of requests that should be
mirrored to BackendRef.

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.<br/>
          <br/>
            <i>Validations</i>:<li>self.numerator <= self.denominator: numerator must be less than or equal to denominator</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>percent</b></td>
        <td>integer</td>
        <td>
          Percent represents the percentage of requests that should be
mirrored to BackendRef. Its minimum value is 0 (indicating 0% of
requests) and its maximum value is 100 (indicating 100% of requests).

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 100<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestMirror.backendRef
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestmirror)</sup></sup>



BackendRef references a resource where mirrored requests are sent.

Mirrored requests must be sent only to a single destination endpoint
within this BackendRef, irrespective of how many endpoints are present
within this BackendRef.

If the referent cannot be found, this BackendRef is invalid and must be
dropped from the Gateway. The controller must ensure the "ResolvedRefs"
condition on the Route status is set to `status: False` and not configure
this backend in the underlying implementation.

If there is a cross-namespace reference to an *existing* object
that is not allowed by a ReferenceGrant, the controller must ensure the
"ResolvedRefs"  condition on the Route is set to `status: False`,
with the "RefNotPermitted" reason and not configure this backend in the
underlying implementation.

In either error case, the Message of the `ResolvedRefs` Condition
should be used to provide more detail about the problem.

Support: Extended for Kubernetes Service

Support: Implementation-specific for any other resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent. For example, "gateway.networking.k8s.io".
When unspecified or empty string, core API group is inferred.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the Kubernetes resource kind of the referent. For example
"Service".

Defaults to "Service" when not specified.

ExternalName services can refer to CNAME DNS records that may live
outside of the cluster and as such are difficult to reason about in
terms of conformance. They also may not be safe to forward to (see
CVE-2021-25740 for more information). Implementations SHOULD NOT
support ExternalName Services.

Support: Core (Services with a type other than ExternalName)

Support: Implementation-specific (Services with type ExternalName)<br/>
          <br/>
            <i>Default</i>: Service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of the backend. When unspecified, the local
namespace is inferred.

Note that when a namespace different than the local namespace is specified,
a ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port specifies the destination port number to use for this resource.
Port is required when the referent is a Kubernetes Service. In this
case, the port number is the service port number, not the target port.
For other resources, destination port might be derived from the referent
resource or this field.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestMirror.fraction
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestmirror)</sup></sup>



Fraction represents the fraction of requests that should be
mirrored to BackendRef.

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>numerator</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>denominator</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestRedirect
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



RequestRedirect defines a schema for a filter that responds to the
request with an HTTP redirection.

Support: Core

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Hostname is the hostname to be used in the value of the `Location`
header in the response.
When empty, the hostname in the `Host` header of the request is used.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestredirectpath">path</a></b></td>
        <td>object</td>
        <td>
          Path defines parameters used to modify the path of the incoming request.
The modified path is then used to construct the `Location` header. When
empty, the request path is used as-is.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>self.type == 'ReplaceFullPath' ? has(self.replaceFullPath) : true: replaceFullPath must be specified when type is set to 'ReplaceFullPath'</li><li>has(self.replaceFullPath) ? self.type == 'ReplaceFullPath' : true: type must be 'ReplaceFullPath' when replaceFullPath is set</li><li>self.type == 'ReplacePrefixMatch' ? has(self.replacePrefixMatch) : true: replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'</li><li>has(self.replacePrefixMatch) ? self.type == 'ReplacePrefixMatch' : true: type must be 'ReplacePrefixMatch' when replacePrefixMatch is set</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the port to be used in the value of the `Location`
header in the response.

If no port is specified, the redirect port MUST be derived using the
following rules:

* If redirect scheme is not-empty, the redirect port MUST be the well-known
  port associated with the redirect scheme. Specifically "http" to port 80
  and "https" to port 443. If the redirect scheme does not have a
  well-known port, the listener port of the Gateway SHOULD be used.
* If redirect scheme is empty, the redirect port MUST be the Gateway
  Listener port.

Implementations SHOULD NOT add the port number in the 'Location'
header in the following cases:

* A Location header that will use HTTP (whether that is determined via
  the Listener protocol or the Scheme field) _and_ use port 80.
* A Location header that will use HTTPS (whether that is determined via
  the Listener protocol or the Scheme field) _and_ use port 443.

Support: Extended<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>enum</td>
        <td>
          Scheme is the scheme to be used in the value of the `Location` header in
the response. When empty, the scheme of the request is used.

Scheme redirects can affect the port of the redirect, for more information,
refer to the documentation for the port field of this filter.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

Support: Extended<br/>
          <br/>
            <i>Enum</i>: http, https<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>statusCode</b></td>
        <td>integer</td>
        <td>
          StatusCode is the HTTP status code to be used in response.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

Support: Core<br/>
          <br/>
            <i>Enum</i>: 301, 302<br/>
            <i>Default</i>: 302<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].requestRedirect.path
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexrequestredirect)</sup></sup>



Path defines parameters used to modify the path of the incoming request.
The modified path is then used to construct the `Location` header. When
empty, the request path is used as-is.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type defines the type of path modifier. Additional types may be
added in a future release of the API.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.<br/>
          <br/>
            <i>Enum</i>: ReplaceFullPath, ReplacePrefixMatch<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>replaceFullPath</b></td>
        <td>string</td>
        <td>
          ReplaceFullPath specifies the value with which to replace the full path
of a request during a rewrite or redirect.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replacePrefixMatch</b></td>
        <td>string</td>
        <td>
          ReplacePrefixMatch specifies the value with which to replace the prefix
match of a request during a rewrite or redirect. For example, a request
to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch
of "/xyz" would be modified to "/xyz/bar".

Note that this matches the behavior of the PathPrefix match type. This
matches full path elements. A path element refers to the list of labels
in the path split by the `/` separator. When specified, a trailing `/` is
ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
match the prefix `/abc`, but the path `/abcd` would not.

ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.
Using any other HTTPRouteMatch type on the same HTTPRouteRule will result in
the implementation setting the Accepted Condition for the Route to `status: False`.

Request Path | Prefix Match | Replace Prefix | Modified Path<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].responseHeaderModifier
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



ResponseHeaderModifier defines a schema for a filter that modifies response
headers.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexresponseheadermodifieraddindex">add</a></b></td>
        <td>[]object</td>
        <td>
          Add adds the given header(s) (name, value) to the request
before the action. It appends to any existing values associated
with the header name.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  add:
  - name: "my-header"
    value: "bar,baz"

Output:
  GET /foo HTTP/1.1
  my-header: foo,bar,baz<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>remove</b></td>
        <td>[]string</td>
        <td>
          Remove the given header(s) from the HTTP request before the action. The
value of Remove is a list of HTTP header names. Note that the header
names are case-insensitive (see
https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).

Input:
  GET /foo HTTP/1.1
  my-header1: foo
  my-header2: bar
  my-header3: baz

Config:
  remove: ["my-header1", "my-header3"]

Output:
  GET /foo HTTP/1.1
  my-header2: bar<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexresponseheadermodifiersetindex">set</a></b></td>
        <td>[]object</td>
        <td>
          Set overwrites the request with the given header (name, value)
before the action.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  set:
  - name: "my-header"
    value: "bar"

Output:
  GET /foo HTTP/1.1
  my-header: bar<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].responseHeaderModifier.add[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexresponseheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].responseHeaderModifier.set[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexresponseheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].urlRewrite
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindex)</sup></sup>



URLRewrite defines a schema for a filter that modifies a request during forwarding.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Hostname is the value to be used to replace the Host header value during
forwarding.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexurlrewritepath">path</a></b></td>
        <td>object</td>
        <td>
          Path defines a path rewrite.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>self.type == 'ReplaceFullPath' ? has(self.replaceFullPath) : true: replaceFullPath must be specified when type is set to 'ReplaceFullPath'</li><li>has(self.replaceFullPath) ? self.type == 'ReplaceFullPath' : true: type must be 'ReplaceFullPath' when replaceFullPath is set</li><li>self.type == 'ReplacePrefixMatch' ? has(self.replacePrefixMatch) : true: replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'</li><li>has(self.replacePrefixMatch) ? self.type == 'ReplacePrefixMatch' : true: type must be 'ReplacePrefixMatch' when replacePrefixMatch is set</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].backendRefs[index].filters[index].urlRewrite.path
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexbackendrefsindexfiltersindexurlrewrite)</sup></sup>



Path defines a path rewrite.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type defines the type of path modifier. Additional types may be
added in a future release of the API.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.<br/>
          <br/>
            <i>Enum</i>: ReplaceFullPath, ReplacePrefixMatch<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>replaceFullPath</b></td>
        <td>string</td>
        <td>
          ReplaceFullPath specifies the value with which to replace the full path
of a request during a rewrite or redirect.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replacePrefixMatch</b></td>
        <td>string</td>
        <td>
          ReplacePrefixMatch specifies the value with which to replace the prefix
match of a request during a rewrite or redirect. For example, a request
to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch
of "/xyz" would be modified to "/xyz/bar".

Note that this matches the behavior of the PathPrefix match type. This
matches full path elements. A path element refers to the list of labels
in the path split by the `/` separator. When specified, a trailing `/` is
ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
match the prefix `/abc`, but the path `/abcd` would not.

ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.
Using any other HTTPRouteMatch type on the same HTTPRouteRule will result in
the implementation setting the Accepted Condition for the Route to `status: False`.

Request Path | Prefix Match | Replace Prefix | Modified Path<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



HTTPRouteFilter defines processing steps that must be completed during the
request or response lifecycle. HTTPRouteFilters are meant as an extension
point to express processing that may be done in Gateway implementations. Some
examples include request or response modification, implementing
authentication strategies, rate-limiting, and traffic shaping. API
guarantee/conformance is defined based on the type of the filter.

<gateway:experimental:validation:XValidation:message="filter.cors must be nil if the filter.type is not CORS",rule="!(has(self.cors) && self.type != 'CORS')">
<gateway:experimental:validation:XValidation:message="filter.cors must be specified for CORS filter.type",rule="!(!has(self.cors) && self.type == 'CORS')">

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type identifies the type of filter to apply. As with other API fields,
types are classified into three conformance levels:

- Core: Filter types and their corresponding configuration defined by
  "Support: Core" in this package, e.g. "RequestHeaderModifier". All
  implementations must support core filters.

- Extended: Filter types and their corresponding configuration defined by
  "Support: Extended" in this package, e.g. "RequestMirror". Implementers
  are encouraged to support extended filters.

- Implementation-specific: Filters that are defined and supported by
  specific vendors.
  In the future, filters showing convergence in behavior across multiple
  implementations will be considered for inclusion in extended or core
  conformance levels. Filter-specific configuration for such filters
  is specified using the ExtensionRef field. `Type` should be set to
  "ExtensionRef" for custom filters.

Implementers are encouraged to define custom implementation types to
extend the core API with implementation-specific behavior.

If a reference to a custom filter type cannot be resolved, the filter
MUST NOT be skipped. Instead, requests that would have been processed by
that filter MUST receive a HTTP error response.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

<gateway:experimental:validation:Enum=RequestHeaderModifier;ResponseHeaderModifier;RequestMirror;RequestRedirect;URLRewrite;ExtensionRef;CORS><br/>
          <br/>
            <i>Enum</i>: RequestHeaderModifier, ResponseHeaderModifier, RequestMirror, RequestRedirect, URLRewrite, ExtensionRef<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexcors">cors</a></b></td>
        <td>object</td>
        <td>
          CORS defines a schema for a filter that responds to the
cross-origin request based on HTTP response header.

Support: Extended

<gateway:experimental><br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexextensionref">extensionRef</a></b></td>
        <td>object</td>
        <td>
          ExtensionRef is an optional, implementation-specific extension to the
"filter" behavior.  For example, resource "myroutefilter" in group
"networking.example.net"). ExtensionRef MUST NOT be used for core and
extended filters.

This filter can be used multiple times within the same rule.

Support: Implementation-specific<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestheadermodifier">requestHeaderModifier</a></b></td>
        <td>object</td>
        <td>
          RequestHeaderModifier defines a schema for a filter that modifies request
headers.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestmirror">requestMirror</a></b></td>
        <td>object</td>
        <td>
          RequestMirror defines a schema for a filter that mirrors requests.
Requests are sent to the specified destination, but responses from
that destination are ignored.

This filter can be used multiple times within the same rule. Note that
not all implementations will be able to support mirroring to multiple
backends.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>!(has(self.percent) && has(self.fraction)): Only one of percent or fraction may be specified in HTTPRequestMirrorFilter</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestredirect">requestRedirect</a></b></td>
        <td>object</td>
        <td>
          RequestRedirect defines a schema for a filter that responds to the
request with an HTTP redirection.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexresponseheadermodifier">responseHeaderModifier</a></b></td>
        <td>object</td>
        <td>
          ResponseHeaderModifier defines a schema for a filter that modifies response
headers.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexurlrewrite">urlRewrite</a></b></td>
        <td>object</td>
        <td>
          URLRewrite defines a schema for a filter that modifies a request during forwarding.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].cors
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



CORS defines a schema for a filter that responds to the
cross-origin request based on HTTP response header.

Support: Extended

<gateway:experimental>

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allowCredentials</b></td>
        <td>boolean</td>
        <td>
          AllowCredentials indicates whether the actual cross-origin request allows
to include credentials.

The only valid value for the `Access-Control-Allow-Credentials` response
header is true (case-sensitive).

If the credentials are not allowed in cross-origin requests, the gateway
will omit the header `Access-Control-Allow-Credentials` entirely rather
than setting its value to false.

Support: Extended<br/>
          <br/>
            <i>Enum</i>: true<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowHeaders</b></td>
        <td>[]string</td>
        <td>
          AllowHeaders indicates which HTTP request headers are supported for
accessing the requested resource.

Header names are not case sensitive.

Multiple header names in the value of the `Access-Control-Allow-Headers`
response header are separated by a comma (",").

When the `AllowHeaders` field is configured with one or more headers, the
gateway must return the `Access-Control-Allow-Headers` response header
which value is present in the `AllowHeaders` field.

If any header name in the `Access-Control-Request-Headers` request header
is not included in the list of header names specified by the response
header `Access-Control-Allow-Headers`, it will present an error on the
client side.

If any header name in the `Access-Control-Allow-Headers` response header
does not recognize by the client, it will also occur an error on the
client side.

A wildcard indicates that the requests with all HTTP headers are allowed.
The `Access-Control-Allow-Headers` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowHeaders` field
specified with the `*` wildcard, the gateway must specify one or more
HTTP headers in the value of the `Access-Control-Allow-Headers` response
header. The value of the header `Access-Control-Allow-Headers` is same as
the `Access-Control-Request-Headers` header provided by the client. If
the header `Access-Control-Request-Headers` is not included in the
request, the gateway will omit the `Access-Control-Allow-Headers`
response header, instead of specifying the `*` wildcard. A Gateway
implementation may choose to add implementation-specific default headers.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowMethods</b></td>
        <td>[]enum</td>
        <td>
          AllowMethods indicates which HTTP methods are supported for accessing the
requested resource.

Valid values are any method defined by RFC9110, along with the special
value `*`, which represents all HTTP methods are allowed.

Method names are case sensitive, so these values are also case-sensitive.
(See https://www.rfc-editor.org/rfc/rfc2616#section-5.1.1)

Multiple method names in the value of the `Access-Control-Allow-Methods`
response header are separated by a comma (",").

A CORS-safelisted method is a method that is `GET`, `HEAD`, or `POST`.
(See https://fetch.spec.whatwg.org/#cors-safelisted-method) The
CORS-safelisted methods are always allowed, regardless of whether they
are specified in the `AllowMethods` field.

When the `AllowMethods` field is configured with one or more methods, the
gateway must return the `Access-Control-Allow-Methods` response header
which value is present in the `AllowMethods` field.

If the HTTP method of the `Access-Control-Request-Method` request header
is not included in the list of methods specified by the response header
`Access-Control-Allow-Methods`, it will present an error on the client
side.

The `Access-Control-Allow-Methods` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowMethods` field
specified with the `*` wildcard, the gateway must specify one HTTP method
in the value of the Access-Control-Allow-Methods response header. The
value of the header `Access-Control-Allow-Methods` is same as the
`Access-Control-Request-Method` header provided by the client. If the
header `Access-Control-Request-Method` is not included in the request,
the gateway will omit the `Access-Control-Allow-Methods` response header,
instead of specifying the `*` wildcard. A Gateway implementation may
choose to add implementation-specific default methods.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>!('*' in self && self.size() > 1): AllowMethods cannot contain '*' alongside other methods</li>
            <i>Enum</i>: GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH, *<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>allowOrigins</b></td>
        <td>[]string</td>
        <td>
          AllowOrigins indicates whether the response can be shared with requested
resource from the given `Origin`.

The `Origin` consists of a scheme and a host, with an optional port, and
takes the form `<scheme>://<host>(:<port>)`.

Valid values for scheme are: `http` and `https`.

Valid values for port are any integer between 1 and 65535 (the list of
available TCP/UDP ports). Note that, if not included, port `80` is
assumed for `http` scheme origins, and port `443` is assumed for `https`
origins. This may affect origin matching.

The host part of the origin may contain the wildcard character `*`. These
wildcard characters behave as follows:

* `*` is a greedy match to the _left_, including any number of
  DNS labels to the left of its position. This also means that
  `*` will include any number of period `.` characters to the
  left of its position.
* A wildcard by itself matches all hosts.

An origin value that includes _only_ the `*` character indicates requests
from all `Origin`s are allowed.

When the `AllowOrigins` field is configured with multiple origins, it
means the server supports clients from multiple origins. If the request
`Origin` matches the configured allowed origins, the gateway must return
the given `Origin` and sets value of the header
`Access-Control-Allow-Origin` same as the `Origin` header provided by the
client.

The status code of a successful response to a "preflight" request is
always an OK status (i.e., 204 or 200).

If the request `Origin` does not match the configured allowed origins,
the gateway returns 204/200 response but doesn't set the relevant
cross-origin response headers. Alternatively, the gateway responds with
403 status to the "preflight" request is denied, coupled with omitting
the CORS headers. The cross-origin request fails on the client side.
Therefore, the client doesn't attempt the actual cross-origin request.

The `Access-Control-Allow-Origin` response header can only use `*`
wildcard as value when the `AllowCredentials` field is unspecified.

When the `AllowCredentials` field is specified and `AllowOrigins` field
specified with the `*` wildcard, the gateway must return a single origin
in the value of the `Access-Control-Allow-Origin` response header,
instead of specifying the `*` wildcard. The value of the header
`Access-Control-Allow-Origin` is same as the `Origin` header provided by
the client.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>exposeHeaders</b></td>
        <td>[]string</td>
        <td>
          ExposeHeaders indicates which HTTP response headers can be exposed
to client-side scripts in response to a cross-origin request.

A CORS-safelisted response header is an HTTP header in a CORS response
that it is considered safe to expose to the client scripts.
The CORS-safelisted response headers include the following headers:
`Cache-Control`
`Content-Language`
`Content-Length`
`Content-Type`
`Expires`
`Last-Modified`
`Pragma`
(See https://fetch.spec.whatwg.org/#cors-safelisted-response-header-name)
The CORS-safelisted response headers are exposed to client by default.

When an HTTP header name is specified using the `ExposeHeaders` field,
this additional header will be exposed as part of the response to the
client.

Header names are not case sensitive.

Multiple header names in the value of the `Access-Control-Expose-Headers`
response header are separated by a comma (",").

A wildcard indicates that the responses with all HTTP headers are exposed
to clients. The `Access-Control-Expose-Headers` response header can only
use `*` wildcard as value when the `AllowCredentials` field is
unspecified.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>maxAge</b></td>
        <td>integer</td>
        <td>
          MaxAge indicates the duration (in seconds) for the client to cache the
results of a "preflight" request.

The information provided by the `Access-Control-Allow-Methods` and
`Access-Control-Allow-Headers` response headers can be cached by the
client until the time specified by `Access-Control-Max-Age` elapses.

The default value of `Access-Control-Max-Age` response header is 5
(seconds).<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 5<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].extensionRef
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



ExtensionRef is an optional, implementation-specific extension to the
"filter" behavior.  For example, resource "myroutefilter" in group
"networking.example.net"). ExtensionRef MUST NOT be used for core and
extended filters.

This filter can be used multiple times within the same rule.

Support: Implementation-specific

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent. For example, "gateway.networking.k8s.io".
When unspecified or empty string, core API group is inferred.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is kind of the referent. For example "HTTPRoute" or "Service".<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestHeaderModifier
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



RequestHeaderModifier defines a schema for a filter that modifies request
headers.

Support: Core

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestheadermodifieraddindex">add</a></b></td>
        <td>[]object</td>
        <td>
          Add adds the given header(s) (name, value) to the request
before the action. It appends to any existing values associated
with the header name.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  add:
  - name: "my-header"
    value: "bar,baz"

Output:
  GET /foo HTTP/1.1
  my-header: foo,bar,baz<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>remove</b></td>
        <td>[]string</td>
        <td>
          Remove the given header(s) from the HTTP request before the action. The
value of Remove is a list of HTTP header names. Note that the header
names are case-insensitive (see
https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).

Input:
  GET /foo HTTP/1.1
  my-header1: foo
  my-header2: bar
  my-header3: baz

Config:
  remove: ["my-header1", "my-header3"]

Output:
  GET /foo HTTP/1.1
  my-header2: bar<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestheadermodifiersetindex">set</a></b></td>
        <td>[]object</td>
        <td>
          Set overwrites the request with the given header (name, value)
before the action.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  set:
  - name: "my-header"
    value: "bar"

Output:
  GET /foo HTTP/1.1
  my-header: bar<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestHeaderModifier.add[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexrequestheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestHeaderModifier.set[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexrequestheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestMirror
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



RequestMirror defines a schema for a filter that mirrors requests.
Requests are sent to the specified destination, but responses from
that destination are ignored.

This filter can be used multiple times within the same rule. Note that
not all implementations will be able to support mirroring to multiple
backends.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestmirrorbackendref">backendRef</a></b></td>
        <td>object</td>
        <td>
          BackendRef references a resource where mirrored requests are sent.

Mirrored requests must be sent only to a single destination endpoint
within this BackendRef, irrespective of how many endpoints are present
within this BackendRef.

If the referent cannot be found, this BackendRef is invalid and must be
dropped from the Gateway. The controller must ensure the "ResolvedRefs"
condition on the Route status is set to `status: False` and not configure
this backend in the underlying implementation.

If there is a cross-namespace reference to an *existing* object
that is not allowed by a ReferenceGrant, the controller must ensure the
"ResolvedRefs"  condition on the Route is set to `status: False`,
with the "RefNotPermitted" reason and not configure this backend in the
underlying implementation.

In either error case, the Message of the `ResolvedRefs` Condition
should be used to provide more detail about the problem.

Support: Extended for Kubernetes Service

Support: Implementation-specific for any other resource<br/>
          <br/>
            <i>Validations</i>:<li>(size(self.group) == 0 && self.kind == 'Service') ? has(self.port) : true: Must have port for Service reference</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestmirrorfraction">fraction</a></b></td>
        <td>object</td>
        <td>
          Fraction represents the fraction of requests that should be
mirrored to BackendRef.

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.<br/>
          <br/>
            <i>Validations</i>:<li>self.numerator <= self.denominator: numerator must be less than or equal to denominator</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>percent</b></td>
        <td>integer</td>
        <td>
          Percent represents the percentage of requests that should be
mirrored to BackendRef. Its minimum value is 0 (indicating 0% of
requests) and its maximum value is 100 (indicating 100% of requests).

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 100<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestMirror.backendRef
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexrequestmirror)</sup></sup>



BackendRef references a resource where mirrored requests are sent.

Mirrored requests must be sent only to a single destination endpoint
within this BackendRef, irrespective of how many endpoints are present
within this BackendRef.

If the referent cannot be found, this BackendRef is invalid and must be
dropped from the Gateway. The controller must ensure the "ResolvedRefs"
condition on the Route status is set to `status: False` and not configure
this backend in the underlying implementation.

If there is a cross-namespace reference to an *existing* object
that is not allowed by a ReferenceGrant, the controller must ensure the
"ResolvedRefs"  condition on the Route is set to `status: False`,
with the "RefNotPermitted" reason and not configure this backend in the
underlying implementation.

In either error case, the Message of the `ResolvedRefs` Condition
should be used to provide more detail about the problem.

Support: Extended for Kubernetes Service

Support: Implementation-specific for any other resource

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the referent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          Group is the group of the referent. For example, "gateway.networking.k8s.io".
When unspecified or empty string, core API group is inferred.<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the Kubernetes resource kind of the referent. For example
"Service".

Defaults to "Service" when not specified.

ExternalName services can refer to CNAME DNS records that may live
outside of the cluster and as such are difficult to reason about in
terms of conformance. They also may not be safe to forward to (see
CVE-2021-25740 for more information). Implementations SHOULD NOT
support ExternalName Services.

Support: Core (Services with a type other than ExternalName)

Support: Implementation-specific (Services with type ExternalName)<br/>
          <br/>
            <i>Default</i>: Service<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace is the namespace of the backend. When unspecified, the local
namespace is inferred.

Note that when a namespace different than the local namespace is specified,
a ReferenceGrant object is required in the referent namespace to allow that
namespace's owner to accept the reference. See the ReferenceGrant
documentation for details.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port specifies the destination port number to use for this resource.
Port is required when the referent is a Kubernetes Service. In this
case, the port number is the service port number, not the target port.
For other resources, destination port might be derived from the referent
resource or this field.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestMirror.fraction
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexrequestmirror)</sup></sup>



Fraction represents the fraction of requests that should be
mirrored to BackendRef.

Only one of Fraction or Percent may be specified. If neither field
is specified, 100% of requests will be mirrored.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>numerator</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>denominator</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100<br/>
            <i>Minimum</i>: 1<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestRedirect
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



RequestRedirect defines a schema for a filter that responds to the
request with an HTTP redirection.

Support: Core

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Hostname is the hostname to be used in the value of the `Location`
header in the response.
When empty, the hostname in the `Host` header of the request is used.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexrequestredirectpath">path</a></b></td>
        <td>object</td>
        <td>
          Path defines parameters used to modify the path of the incoming request.
The modified path is then used to construct the `Location` header. When
empty, the request path is used as-is.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>self.type == 'ReplaceFullPath' ? has(self.replaceFullPath) : true: replaceFullPath must be specified when type is set to 'ReplaceFullPath'</li><li>has(self.replaceFullPath) ? self.type == 'ReplaceFullPath' : true: type must be 'ReplaceFullPath' when replaceFullPath is set</li><li>self.type == 'ReplacePrefixMatch' ? has(self.replacePrefixMatch) : true: replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'</li><li>has(self.replacePrefixMatch) ? self.type == 'ReplacePrefixMatch' : true: type must be 'ReplacePrefixMatch' when replacePrefixMatch is set</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          Port is the port to be used in the value of the `Location`
header in the response.

If no port is specified, the redirect port MUST be derived using the
following rules:

* If redirect scheme is not-empty, the redirect port MUST be the well-known
  port associated with the redirect scheme. Specifically "http" to port 80
  and "https" to port 443. If the redirect scheme does not have a
  well-known port, the listener port of the Gateway SHOULD be used.
* If redirect scheme is empty, the redirect port MUST be the Gateway
  Listener port.

Implementations SHOULD NOT add the port number in the 'Location'
header in the following cases:

* A Location header that will use HTTP (whether that is determined via
  the Listener protocol or the Scheme field) _and_ use port 80.
* A Location header that will use HTTPS (whether that is determined via
  the Listener protocol or the Scheme field) _and_ use port 443.

Support: Extended<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Minimum</i>: 1<br/>
            <i>Maximum</i>: 65535<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>scheme</b></td>
        <td>enum</td>
        <td>
          Scheme is the scheme to be used in the value of the `Location` header in
the response. When empty, the scheme of the request is used.

Scheme redirects can affect the port of the redirect, for more information,
refer to the documentation for the port field of this filter.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

Support: Extended<br/>
          <br/>
            <i>Enum</i>: http, https<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>statusCode</b></td>
        <td>integer</td>
        <td>
          StatusCode is the HTTP status code to be used in response.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.

Support: Core<br/>
          <br/>
            <i>Enum</i>: 301, 302<br/>
            <i>Default</i>: 302<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].requestRedirect.path
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexrequestredirect)</sup></sup>



Path defines parameters used to modify the path of the incoming request.
The modified path is then used to construct the `Location` header. When
empty, the request path is used as-is.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type defines the type of path modifier. Additional types may be
added in a future release of the API.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.<br/>
          <br/>
            <i>Enum</i>: ReplaceFullPath, ReplacePrefixMatch<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>replaceFullPath</b></td>
        <td>string</td>
        <td>
          ReplaceFullPath specifies the value with which to replace the full path
of a request during a rewrite or redirect.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replacePrefixMatch</b></td>
        <td>string</td>
        <td>
          ReplacePrefixMatch specifies the value with which to replace the prefix
match of a request during a rewrite or redirect. For example, a request
to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch
of "/xyz" would be modified to "/xyz/bar".

Note that this matches the behavior of the PathPrefix match type. This
matches full path elements. A path element refers to the list of labels
in the path split by the `/` separator. When specified, a trailing `/` is
ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
match the prefix `/abc`, but the path `/abcd` would not.

ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.
Using any other HTTPRouteMatch type on the same HTTPRouteRule will result in
the implementation setting the Accepted Condition for the Route to `status: False`.

Request Path | Prefix Match | Replace Prefix | Modified Path<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].responseHeaderModifier
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



ResponseHeaderModifier defines a schema for a filter that modifies response
headers.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexresponseheadermodifieraddindex">add</a></b></td>
        <td>[]object</td>
        <td>
          Add adds the given header(s) (name, value) to the request
before the action. It appends to any existing values associated
with the header name.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  add:
  - name: "my-header"
    value: "bar,baz"

Output:
  GET /foo HTTP/1.1
  my-header: foo,bar,baz<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>remove</b></td>
        <td>[]string</td>
        <td>
          Remove the given header(s) from the HTTP request before the action. The
value of Remove is a list of HTTP header names. Note that the header
names are case-insensitive (see
https://datatracker.ietf.org/doc/html/rfc2616#section-4.2).

Input:
  GET /foo HTTP/1.1
  my-header1: foo
  my-header2: bar
  my-header3: baz

Config:
  remove: ["my-header1", "my-header3"]

Output:
  GET /foo HTTP/1.1
  my-header2: bar<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexresponseheadermodifiersetindex">set</a></b></td>
        <td>[]object</td>
        <td>
          Set overwrites the request with the given header (name, value)
before the action.

Input:
  GET /foo HTTP/1.1
  my-header: foo

Config:
  set:
  - name: "my-header"
    value: "bar"

Output:
  GET /foo HTTP/1.1
  my-header: bar<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].responseHeaderModifier.add[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexresponseheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].responseHeaderModifier.set[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexresponseheadermodifier)</sup></sup>



HTTPHeader represents an HTTP Header name and value as defined by RFC 7230.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, the first entry with
an equivalent name MUST be considered for a match. Subsequent entries
with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].urlRewrite
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindex)</sup></sup>



URLRewrite defines a schema for a filter that modifies a request during forwarding.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hostname</b></td>
        <td>string</td>
        <td>
          Hostname is the value to be used to replace the Host header value during
forwarding.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexfiltersindexurlrewritepath">path</a></b></td>
        <td>object</td>
        <td>
          Path defines a path rewrite.

Support: Extended<br/>
          <br/>
            <i>Validations</i>:<li>self.type == 'ReplaceFullPath' ? has(self.replaceFullPath) : true: replaceFullPath must be specified when type is set to 'ReplaceFullPath'</li><li>has(self.replaceFullPath) ? self.type == 'ReplaceFullPath' : true: type must be 'ReplaceFullPath' when replaceFullPath is set</li><li>self.type == 'ReplacePrefixMatch' ? has(self.replacePrefixMatch) : true: replacePrefixMatch must be specified when type is set to 'ReplacePrefixMatch'</li><li>has(self.replacePrefixMatch) ? self.type == 'ReplacePrefixMatch' : true: type must be 'ReplacePrefixMatch' when replacePrefixMatch is set</li>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].filters[index].urlRewrite.path
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexfiltersindexurlrewrite)</sup></sup>



Path defines a path rewrite.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type defines the type of path modifier. Additional types may be
added in a future release of the API.

Note that values may be added to this enum, implementations
must ensure that unknown values will not cause a crash.

Unknown values here must result in the implementation setting the
Accepted Condition for the Route to `status: False`, with a
Reason of `UnsupportedValue`.<br/>
          <br/>
            <i>Enum</i>: ReplaceFullPath, ReplacePrefixMatch<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>replaceFullPath</b></td>
        <td>string</td>
        <td>
          ReplaceFullPath specifies the value with which to replace the full path
of a request during a rewrite or redirect.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>replacePrefixMatch</b></td>
        <td>string</td>
        <td>
          ReplacePrefixMatch specifies the value with which to replace the prefix
match of a request during a rewrite or redirect. For example, a request
to "/foo/bar" with a prefix match of "/foo" and a ReplacePrefixMatch
of "/xyz" would be modified to "/xyz/bar".

Note that this matches the behavior of the PathPrefix match type. This
matches full path elements. A path element refers to the list of labels
in the path split by the `/` separator. When specified, a trailing `/` is
ignored. For example, the paths `/abc`, `/abc/`, and `/abc/def` would all
match the prefix `/abc`, but the path `/abcd` would not.

ReplacePrefixMatch is only compatible with a `PathPrefix` HTTPRouteMatch.
Using any other HTTPRouteMatch type on the same HTTPRouteRule will result in
the implementation setting the Accepted Condition for the Route to `status: False`.

Request Path | Prefix Match | Replace Prefix | Modified Path<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].matches[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



HTTPRouteMatch defines the predicate used to match requests to a given
action. Multiple match types are ANDed together, i.e. the match will
evaluate to true only if all conditions are satisfied.

For example, the match below will match a HTTP request only if its path
starts with `/foo` AND it contains the `version: v1` header:

```
match:

	path:
	  value: "/foo"
	headers:
	- name: "version"
	  value "v1"

```

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexmatchesindexheadersindex">headers</a></b></td>
        <td>[]object</td>
        <td>
          Headers specifies HTTP request header matchers. Multiple match values are
ANDed together, meaning, a request must match all the specified headers
to select the route.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>method</b></td>
        <td>enum</td>
        <td>
          Method specifies HTTP method matcher.
When specified, this route will be matched only if the request has the
specified method.

Support: Extended<br/>
          <br/>
            <i>Enum</i>: GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexmatchesindexpath">path</a></b></td>
        <td>object</td>
        <td>
          Path specifies a HTTP request path matcher. If this field is not
specified, a default prefix match on the "/" path is provided.<br/>
          <br/>
            <i>Validations</i>:<li>(self.type in ['Exact','PathPrefix']) ? self.value.startsWith('/') : true: value must be an absolute path and start with '/' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('//') : true: must not contain '//' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('/./') : true: must not contain '/./' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('/../') : true: must not contain '/../' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('%2f') : true: must not contain '%2f' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('%2F') : true: must not contain '%2F' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.contains('#') : true: must not contain '#' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.endsWith('/..') : true: must not end with '/..' when type one of ['Exact', 'PathPrefix']</li><li>(self.type in ['Exact','PathPrefix']) ? !self.value.endsWith('/.') : true: must not end with '/.' when type one of ['Exact', 'PathPrefix']</li><li>self.type in ['Exact','PathPrefix'] || self.type == 'RegularExpression': type must be one of ['Exact', 'PathPrefix', 'RegularExpression']</li><li>(self.type in ['Exact','PathPrefix']) ? self.value.matches(r"""^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$""") : true: must only contain valid characters (matching ^(?:[-A-Za-z0-9/._~!$&'()*+,;=:@]|[%][0-9a-fA-F]{2})+$) for types ['Exact', 'PathPrefix']</li>
            <i>Default</i>: map[type:PathPrefix value:/]<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexmatchesindexqueryparamsindex">queryParams</a></b></td>
        <td>[]object</td>
        <td>
          QueryParams specifies HTTP query parameter matchers. Multiple match
values are ANDed together, meaning, a request must match all the
specified query parameters to select the route.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].matches[index].headers[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexmatchesindex)</sup></sup>



HTTPHeaderMatch describes how to select a HTTP route by matching HTTP request
headers.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP Header to be matched. Name matching MUST be
case-insensitive. (See https://tools.ietf.org/html/rfc7230#section-3.2).

If multiple entries specify equivalent header names, only the first
entry with an equivalent name MUST be considered for a match. Subsequent
entries with an equivalent header name MUST be ignored. Due to the
case-insensitivity of header names, "foo" and "Foo" are considered
equivalent.

When a header is repeated in an HTTP request, it is
implementation-specific behavior as to how this is represented.
Generally, proxies should follow the guidance from the RFC:
https://www.rfc-editor.org/rfc/rfc7230.html#section-3.2.2 regarding
processing a repeated header, with special handling for "Set-Cookie".<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP Header to be matched.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type specifies how to match against the value of the header.

Support: Core (Exact)

Support: Implementation-specific (RegularExpression)

Since RegularExpression HeaderMatchType has implementation-specific
conformance, implementations can support POSIX, PCRE or any other dialects
of regular expressions. Please read the implementation's documentation to
determine the supported dialect.<br/>
          <br/>
            <i>Enum</i>: Exact, RegularExpression<br/>
            <i>Default</i>: Exact<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].matches[index].path
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexmatchesindex)</sup></sup>



Path specifies a HTTP request path matcher. If this field is not
specified, a default prefix match on the "/" path is provided.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type specifies how to match against the path Value.

Support: Core (Exact, PathPrefix)

Support: Implementation-specific (RegularExpression)<br/>
          <br/>
            <i>Enum</i>: Exact, PathPrefix, RegularExpression<br/>
            <i>Default</i>: PathPrefix<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value of the HTTP path to match against.<br/>
          <br/>
            <i>Default</i>: /<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].matches[index].queryParams[index]
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexmatchesindex)</sup></sup>



HTTPQueryParamMatch describes how to select a HTTP route by matching HTTP
query parameters.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of the HTTP query param to be matched. This must be an
exact string match. (See
https://tools.ietf.org/html/rfc7230#section-2.7.3).

If multiple entries specify equivalent query param names, only the first
entry with an equivalent name MUST be considered for a match. Subsequent
entries with an equivalent query param name MUST be ignored.

If a query param is repeated in an HTTP request, the behavior is
purposely left undefined, since different data planes have different
capabilities. However, it is *recommended* that implementations should
match against the first value of the param if the data plane supports it,
as this behavior is expected in other load balancing contexts outside of
the Gateway API.

Users SHOULD NOT route traffic based on repeated query params to guard
themselves against potential differences in the implementations.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of HTTP query param to be matched.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type specifies how to match against the value of the query parameter.

Support: Extended (Exact)

Support: Implementation-specific (RegularExpression)

Since RegularExpression QueryParamMatchType has Implementation-specific
conformance, implementations can support POSIX, PCRE or any other
dialects of regular expressions. Please read the implementation's
documentation to determine the supported dialect.<br/>
          <br/>
            <i>Enum</i>: Exact, RegularExpression<br/>
            <i>Default</i>: Exact<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].retry
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



Retry defines the configuration for when to retry an HTTP request.

Support: Extended

<gateway:experimental>

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>attempts</b></td>
        <td>integer</td>
        <td>
          Attempts specifies the maximum number of times an individual request
from the gateway to a backend should be retried.

If the maximum number of retries has been attempted without a successful
response from the backend, the Gateway MUST return an error.

When this field is unspecified, the number of times to attempt to retry
a backend request is implementation-specific.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>backoff</b></td>
        <td>string</td>
        <td>
          Backoff specifies the minimum duration a Gateway should wait between
retry attempts and is represented in Gateway API Duration formatting.

For example, setting the `rules[].retry.backoff` field to the value
`100ms` will cause a backend request to first be retried approximately
100 milliseconds after timing out or receiving a response code configured
to be retryable.

An implementation MAY use an exponential or alternative backoff strategy
for subsequent retry attempts, MAY cap the maximum backoff duration to
some amount greater than the specified minimum, and MAY add arbitrary
jitter to stagger requests, as long as unsuccessful backend requests are
not retried before the configured minimum duration.

If a Request timeout (`rules[].timeouts.request`) is configured on the
route, the entire duration of the initial request and any retry attempts
MUST not exceed the Request timeout duration. If any retry attempts are
still in progress when the Request timeout duration has been reached,
these SHOULD be canceled if possible and the Gateway MUST immediately
return a timeout error.

If a BackendRequest timeout (`rules[].timeouts.backendRequest`) is
configured on the route, any retry attempts which reach the configured
BackendRequest timeout duration without a response SHOULD be canceled if
possible and the Gateway should wait for at least the specified backoff
duration before attempting to retry the backend request again.

If a BackendRequest timeout is _not_ configured on the route, retry
attempts MAY time out after an implementation default duration, or MAY
remain pending until a configured Request timeout or implementation
default duration for total request time is reached.

When this field is unspecified, the time to wait between retry attempts
is implementation-specific.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>codes</b></td>
        <td>[]integer</td>
        <td>
          Codes defines the HTTP response status codes for which a backend request
should be retried.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].sessionPersistence
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



SessionPersistence defines and configures session persistence
for the route rule.

Support: Extended

<gateway:experimental>

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>absoluteTimeout</b></td>
        <td>string</td>
        <td>
          AbsoluteTimeout defines the absolute timeout of the persistent
session. Once the AbsoluteTimeout duration has elapsed, the
session becomes invalid.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspechttproutespecrulesindexsessionpersistencecookieconfig">cookieConfig</a></b></td>
        <td>object</td>
        <td>
          CookieConfig provides configuration settings that are specific
to cookie-based session persistence.

Support: Core<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>idleTimeout</b></td>
        <td>string</td>
        <td>
          IdleTimeout defines the idle timeout of the persistent session.
Once the session has been idle for more than the specified
IdleTimeout duration, the session becomes invalid.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sessionName</b></td>
        <td>string</td>
        <td>
          SessionName defines the name of the persistent session token
which may be reflected in the cookie or the header. Users
should avoid reusing session names to prevent unintended
consequences, such as rejection or unpredictable behavior.

Support: Implementation-specific<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>enum</td>
        <td>
          Type defines the type of session persistence such as through
the use a header or cookie. Defaults to cookie based session
persistence.

Support: Core for "Cookie" type

Support: Extended for "Header" type<br/>
          <br/>
            <i>Enum</i>: Cookie, Header<br/>
            <i>Default</i>: Cookie<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].sessionPersistence.cookieConfig
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindexsessionpersistence)</sup></sup>



CookieConfig provides configuration settings that are specific
to cookie-based session persistence.

Support: Core

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lifetimeType</b></td>
        <td>enum</td>
        <td>
          LifetimeType specifies whether the cookie has a permanent or
session-based lifetime. A permanent cookie persists until its
specified expiry time, defined by the Expires or Max-Age cookie
attributes, while a session cookie is deleted when the current
session ends.

When set to "Permanent", AbsoluteTimeout indicates the
cookie's lifetime via the Expires or Max-Age cookie attributes
and is required.

When set to "Session", AbsoluteTimeout indicates the
absolute lifetime of the cookie tracked by the gateway and
is optional.

Defaults to "Session".

Support: Core for "Session" type

Support: Extended for "Permanent" type<br/>
          <br/>
            <i>Enum</i>: Permanent, Session<br/>
            <i>Default</i>: Session<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.httpRoute.spec.rules[index].timeouts
<sup><sup>[↩ Parent](#grafanaspechttproutespecrulesindex)</sup></sup>



Timeouts defines the timeouts that can be configured for an HTTP request.

Support: Extended

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>backendRequest</b></td>
        <td>string</td>
        <td>
          BackendRequest specifies a timeout for an individual request from the gateway
to a backend. This covers the time from when the request first starts being
sent from the gateway to when the full response has been received from the backend.

Setting a timeout to the zero duration (e.g. "0s") SHOULD disable the timeout
completely. Implementations that cannot completely disable the timeout MUST
instead interpret the zero duration as the longest possible value to which
the timeout can be set.

An entire client HTTP transaction with a gateway, covered by the Request timeout,
may result in more than one call from the gateway to the destination backend,
for example, if automatic retries are supported.

The value of BackendRequest must be a Gateway API Duration string as defined by
GEP-2257.  When this field is unspecified, its behavior is implementation-specific;
when specified, the value of BackendRequest must be no more than the value of the
Request timeout (since the Request timeout encompasses the BackendRequest timeout).

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>request</b></td>
        <td>string</td>
        <td>
          Request specifies the maximum duration for a gateway to respond to an HTTP request.
If the gateway has not been able to respond before this deadline is met, the gateway
MUST return a timeout error.

For example, setting the `rules.timeouts.request` field to the value `10s` in an
`HTTPRoute` will cause a timeout if a client request is taking longer than 10 seconds
to complete.

Setting a timeout to the zero duration (e.g. "0s") SHOULD disable the timeout
completely. Implementations that cannot completely disable the timeout MUST
instead interpret the zero duration as the longest possible value to which
the timeout can be set.

This timeout is intended to cover as close to the whole request-response transaction
as possible although an implementation MAY choose to start the timeout after the entire
request stream has been received instead of immediately after the transaction is
initiated by the client.

The value of Request is a Gateway API Duration string as defined by GEP-2257. When this
field is unspecified, request timeout behavior is implementation-specific.

Support: Extended<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Ingress sets how the ingress object should look like with your grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspec">spec</a></b></td>
        <td>object</td>
        <td>
          IngressSpec describes the Ingress the user wishes to exist.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.metadata
<sup><sup>[↩ Parent](#grafanaspecingress)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec
<sup><sup>[↩ Parent](#grafanaspecingress)</sup></sup>



IngressSpec describes the Ingress the user wishes to exist.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressspecdefaultbackend">defaultBackend</a></b></td>
        <td>object</td>
        <td>
          defaultBackend is the backend that should handle requests that don't
match any rule. If Rules are not specified, DefaultBackend must be specified.
If DefaultBackend is not set, the handling of requests that do not match any
of the rules will be up to the Ingress controller.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ingressClassName</b></td>
        <td>string</td>
        <td>
          ingressClassName is the name of an IngressClass cluster resource. Ingress
controller implementations use this field to know whether they should be
serving this Ingress resource, by a transitive connection
(controller -> IngressClass -> Ingress resource). Although the
`kubernetes.io/ingress.class` annotation (simple constant name) was never
formally defined, it was widely supported by Ingress controllers to create
a direct binding between Ingress controller and Ingress resources. Newly
created Ingress resources should prefer using the field. However, even
though the annotation is officially deprecated, for backwards compatibility
reasons, ingress controllers should still honor that annotation if present.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecrulesindex">rules</a></b></td>
        <td>[]object</td>
        <td>
          rules is a list of host rules used to configure the Ingress. If unspecified,
or no rule matches, all traffic is sent to the default backend.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspectlsindex">tls</a></b></td>
        <td>[]object</td>
        <td>
          tls represents the TLS configuration. Currently the Ingress only supports a
single TLS port, 443. If multiple members of this list specify different hosts,
they will be multiplexed on the same port according to the hostname specified
through the SNI TLS extension, if the ingress controller fulfilling the
ingress supports SNI.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.defaultBackend
<sup><sup>[↩ Parent](#grafanaspecingressspec)</sup></sup>



defaultBackend is the backend that should handle requests that don't
match any rule. If Rules are not specified, DefaultBackend must be specified.
If DefaultBackend is not set, the handling of requests that do not match any
of the rules will be up to the Ingress controller.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressspecdefaultbackendresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource is an ObjectRef to another Kubernetes resource in the namespace
of the Ingress object. If resource is specified, a service.Name and
service.Port must not be specified.
This is a mutually exclusive setting with "Service".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecdefaultbackendservice">service</a></b></td>
        <td>object</td>
        <td>
          service references a service as a backend.
This is a mutually exclusive setting with "Resource".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.defaultBackend.resource
<sup><sup>[↩ Parent](#grafanaspecingressspecdefaultbackend)</sup></sup>



resource is an ObjectRef to another Kubernetes resource in the namespace
of the Ingress object. If resource is specified, a service.Name and
service.Port must not be specified.
This is a mutually exclusive setting with "Service".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.defaultBackend.service
<sup><sup>[↩ Parent](#grafanaspecingressspecdefaultbackend)</sup></sup>



service references a service as a backend.
This is a mutually exclusive setting with "Resource".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the referenced service. The service must exist in
the same namespace as the Ingress object.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecdefaultbackendserviceport">port</a></b></td>
        <td>object</td>
        <td>
          port of the referenced service. A port name or port number
is required for a IngressServiceBackend.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.defaultBackend.service.port
<sup><sup>[↩ Parent](#grafanaspecingressspecdefaultbackendservice)</sup></sup>



port of the referenced service. A port name or port number
is required for a IngressServiceBackend.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the port on the Service.
This is a mutually exclusive setting with "Number".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>number</b></td>
        <td>integer</td>
        <td>
          number is the numerical port number (e.g. 80) on the Service.
This is a mutually exclusive setting with "Name".<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index]
<sup><sup>[↩ Parent](#grafanaspecingressspec)</sup></sup>



IngressRule represents the rules mapping the paths under a specified host to
the related backend services. Incoming requests are first evaluated for a host
match, then routed to the backend associated with the matching IngressRuleValue.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          host is the fully qualified domain name of a network host, as defined by RFC 3986.
Note the following deviations from the "host" part of the
URI as defined in RFC 3986:
1. IPs are not allowed. Currently an IngressRuleValue can only apply to
   the IP in the Spec of the parent Ingress.
2. The `:` delimiter is not respected because ports are not allowed.
	  Currently the port of an Ingress is implicitly :80 for http and
	  :443 for https.
Both these may change in the future.
Incoming requests are matched against the host before the
IngressRuleValue. If the host is unspecified, the Ingress routes all
traffic based on the specified IngressRuleValue.

host can be "precise" which is a domain name without the terminating dot of
a network host (e.g. "foo.bar.com") or "wildcard", which is a domain name
prefixed with a single wildcard label (e.g. "*.foo.com").
The wildcard character '*' must appear by itself as the first DNS label and
matches only a single label. You cannot have a wildcard label by itself (e.g. Host == "*").
Requests will be matched against the Host field in the following way:
1. If host is precise, the request matches this rule if the http host header is equal to Host.
2. If host is a wildcard, then the request matches this rule if the http host header
is to equal to the suffix (removing the first label) of the wildcard rule.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttp">http</a></b></td>
        <td>object</td>
        <td>
          HTTPIngressRuleValue is a list of http selectors pointing to backends.
In the example: http://<host>/<path>?<searchpart> -> backend where
where parts of the url correspond to RFC 3986, this resource will be used
to match against everything after the last '/' and before the first '?'
or '#'.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindex)</sup></sup>



HTTPIngressRuleValue is a list of http selectors pointing to backends.
In the example: http://<host>/<path>?<searchpart> -> backend where
where parts of the url correspond to RFC 3986, this resource will be used
to match against everything after the last '/' and before the first '?'
or '#'.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttppathsindex">paths</a></b></td>
        <td>[]object</td>
        <td>
          paths is a collection of paths that map requests to backends.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http.paths[index]
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindexhttp)</sup></sup>



HTTPIngressPath associates a path with a backend. Incoming urls matching the
path are forwarded to the backend.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttppathsindexbackend">backend</a></b></td>
        <td>object</td>
        <td>
          backend defines the referenced service endpoint to which the traffic
will be forwarded to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>pathType</b></td>
        <td>string</td>
        <td>
          pathType determines the interpretation of the path matching. PathType can
be one of the following values:
* Exact: Matches the URL path exactly.
* Prefix: Matches based on a URL path prefix split by '/'. Matching is
  done on a path element by element basis. A path element refers is the
  list of labels in the path split by the '/' separator. A request is a
  match for path p if every p is an element-wise prefix of p of the
  request path. Note that if the last element of the path is a substring
  of the last element in request path, it is not a match (e.g. /foo/bar
  matches /foo/bar/baz, but does not match /foo/barbaz).
* ImplementationSpecific: Interpretation of the Path matching is up to
  the IngressClass. Implementations can treat this as a separate PathType
  or treat it identically to Prefix or Exact path types.
Implementations are required to support all path types.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          path is matched against the path of an incoming request. Currently it can
contain characters disallowed from the conventional "path" part of a URL
as defined by RFC 3986. Paths must begin with a '/' and must be present
when using PathType with value "Exact" or "Prefix".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http.paths[index].backend
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindexhttppathsindex)</sup></sup>



backend defines the referenced service endpoint to which the traffic
will be forwarded to.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttppathsindexbackendresource">resource</a></b></td>
        <td>object</td>
        <td>
          resource is an ObjectRef to another Kubernetes resource in the namespace
of the Ingress object. If resource is specified, a service.Name and
service.Port must not be specified.
This is a mutually exclusive setting with "Service".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttppathsindexbackendservice">service</a></b></td>
        <td>object</td>
        <td>
          service references a service as a backend.
This is a mutually exclusive setting with "Resource".<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http.paths[index].backend.resource
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindexhttppathsindexbackend)</sup></sup>



resource is an ObjectRef to another Kubernetes resource in the namespace
of the Ingress object. If resource is specified, a service.Name and
service.Port must not be specified.
This is a mutually exclusive setting with "Service".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http.paths[index].backend.service
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindexhttppathsindexbackend)</sup></sup>



service references a service as a backend.
This is a mutually exclusive setting with "Resource".

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the referenced service. The service must exist in
the same namespace as the Ingress object.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaspecingressspecrulesindexhttppathsindexbackendserviceport">port</a></b></td>
        <td>object</td>
        <td>
          port of the referenced service. A port name or port number
is required for a IngressServiceBackend.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.rules[index].http.paths[index].backend.service.port
<sup><sup>[↩ Parent](#grafanaspecingressspecrulesindexhttppathsindexbackendservice)</sup></sup>



port of the referenced service. A port name or port number
is required for a IngressServiceBackend.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name is the name of the port on the Service.
This is a mutually exclusive setting with "Number".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>number</b></td>
        <td>integer</td>
        <td>
          number is the numerical port number (e.g. 80) on the Service.
This is a mutually exclusive setting with "Name".<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.ingress.spec.tls[index]
<sup><sup>[↩ Parent](#grafanaspecingressspec)</sup></sup>



IngressTLS describes the transport layer security associated with an ingress.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>hosts</b></td>
        <td>[]string</td>
        <td>
          hosts is a list of hosts included in the TLS certificate. The values in
this list must match the name/s used in the tlsSecret. Defaults to the
wildcard host setting for the loadbalancer controller fulfilling this
Ingress, if left unspecified.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          secretName is the name of the secret used to terminate TLS traffic on
port 443. Field is left optional to allow TLS routing based on SNI
hostname alone. If the SNI host in a listener conflicts with the "Host"
header field used by an IngressRule, the SNI host is used for termination
and value of the "Host" header is used for routing.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.jsonnet
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecjsonnetlibrarylabelselector">libraryLabelSelector</a></b></td>
        <td>object</td>
        <td>
          A label selector is a label query over a set of resources. The result of matchLabels and
matchExpressions are ANDed. An empty label selector matches all objects. A null
label selector matches no objects.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.jsonnet.libraryLabelSelector
<sup><sup>[↩ Parent](#grafanaspecjsonnet)</sup></sup>



A label selector is a label query over a set of resources. The result of matchLabels and
matchExpressions are ANDed. An empty label selector matches all objects. A null
label selector matches no objects.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecjsonnetlibrarylabelselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.jsonnet.libraryLabelSelector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecjsonnetlibrarylabelselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



PersistentVolumeClaim creates a PVC if you need to attach one to your grafana instance.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.metadata
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaim)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaim)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>accessModes</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecdatasource">dataSource</a></b></td>
        <td>object</td>
        <td>
          TypedLocalObjectReference contains enough information to let you locate the
typed referenced object inside the same namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecdatasourceref">dataSourceRef</a></b></td>
        <td>object</td>
        <td>
          TypedLocalObjectReference contains enough information to let you locate the
typed referenced object inside the same namespace.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecresources">resources</a></b></td>
        <td>object</td>
        <td>
          ResourceRequirements describes the compute resource requirements.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecselector">selector</a></b></td>
        <td>object</td>
        <td>
          A label selector is a label query over a set of resources. The result of matchLabels and
matchExpressions are ANDed. An empty label selector matches all objects. A null
label selector matches no objects.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>storageClassName</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeMode</b></td>
        <td>string</td>
        <td>
          PersistentVolumeMode describes how a volume is intended to be consumed, either Block or Filesystem.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>volumeName</b></td>
        <td>string</td>
        <td>
          VolumeName is the binding reference to the PersistentVolume backing this claim.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.dataSource
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspec)</sup></sup>



TypedLocalObjectReference contains enough information to let you locate the
typed referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.dataSourceRef
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspec)</sup></sup>



TypedLocalObjectReference contains enough information to let you locate the
typed referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind is the type of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is the name of resource being referenced<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>apiGroup</b></td>
        <td>string</td>
        <td>
          APIGroup is the group for the resource being referenced.
If APIGroup is not specified, the specified Kind must be in the core API group.
For any other third-party types, APIGroup is required.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.resources
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspec)</sup></sup>



ResourceRequirements describes the compute resource requirements.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecresourcesclaimsindex">claims</a></b></td>
        <td>[]object</td>
        <td>
          Claims lists the names of resources, defined in spec.resourceClaims,
that are used by this container.

This field depends on the
DynamicResourceAllocation feature gate.

This field is immutable. It can only be set for containers.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>limits</b></td>
        <td>map[string]int or string</td>
        <td>
          Limits describes the maximum amount of compute resources allowed.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>requests</b></td>
        <td>map[string]int or string</td>
        <td>
          Requests describes the minimum amount of compute resources required.
If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
otherwise to an implementation-defined value. Requests cannot exceed Limits.
More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.resources.claims[index]
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspecresources)</sup></sup>



ResourceClaim references one entry in PodSpec.ResourceClaims.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name must match the name of one entry in pod.spec.resourceClaims of
the Pod where this field is used. It makes that resource available
inside a container.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>request</b></td>
        <td>string</td>
        <td>
          Request is the name chosen for a request in the referenced claim.
If empty, everything from the claim is made available, otherwise
only the result of this request.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.selector
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspec)</sup></sup>



A label selector is a label query over a set of resources. The result of matchLabels and
matchExpressions are ANDed. An empty label selector matches all objects. A null
label selector matches no objects.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecpersistentvolumeclaimspecselectormatchexpressionsindex">matchExpressions</a></b></td>
        <td>[]object</td>
        <td>
          matchExpressions is a list of label selector requirements. The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>matchLabels</b></td>
        <td>map[string]string</td>
        <td>
          matchLabels is a map of {key,value} pairs. A single {key,value} in the matchLabels
map is equivalent to an element of matchExpressions, whose key field is "key", the
operator is "In", and the values array contains only "value". The requirements are ANDed.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.persistentVolumeClaim.spec.selector.matchExpressions[index]
<sup><sup>[↩ Parent](#grafanaspecpersistentvolumeclaimspecselector)</sup></sup>



A label selector requirement is a selector that contains values, a key, and an operator that
relates the key and values.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key is the label key that the selector applies to.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>operator</b></td>
        <td>string</td>
        <td>
          operator represents a key's relationship to a set of values.
Valid operators are In, NotIn, Exists and DoesNotExist.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          values is an array of string values. If the operator is In or NotIn,
the values array must be non-empty. If the operator is Exists or DoesNotExist,
the values array must be empty. This array is replaced during a strategic
merge patch.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.preferences
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Preferences holds the Grafana Preferences settings

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>homeDashboardUid</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Route sets how the ingress object should look like with your grafana instance, this only works in Openshift.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecroutemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroutespec">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.metadata
<sup><sup>[↩ Parent](#grafanaspecroute)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec
<sup><sup>[↩ Parent](#grafanaspecroute)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecroutespecalternatebackendsindex">alternateBackends</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>host</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>path</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroutespecport">port</a></b></td>
        <td>object</td>
        <td>
          RoutePort defines a port mapping from a router to an endpoint in the service endpoints.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>subdomain</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroutespectls">tls</a></b></td>
        <td>object</td>
        <td>
          TLSConfig defines config used to secure a route and provide termination<br/>
          <br/>
            <i>Validations</i>:<li>has(self.termination) && has(self.insecureEdgeTerminationPolicy) ? !((self.termination=='passthrough') && (self.insecureEdgeTerminationPolicy=='Allow')) : true: cannot have both spec.tls.termination: passthrough and spec.tls.insecureEdgeTerminationPolicy: Allow</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroutespecto">to</a></b></td>
        <td>object</td>
        <td>
          RouteTargetReference specifies the target that resolve into endpoints. Only the 'Service'
kind is allowed. Use 'weight' field to emphasize one over others.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>wildcardPolicy</b></td>
        <td>string</td>
        <td>
          WildcardPolicyType indicates the type of wildcard support needed by routes.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec.alternateBackends[index]
<sup><sup>[↩ Parent](#grafanaspecroutespec)</sup></sup>



RouteTargetReference specifies the target that resolve into endpoints. Only the 'Service'
kind is allowed. Use 'weight' field to emphasize one over others.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          The kind of target that the route is referring to. Currently, only 'Service' is allowed<br/>
          <br/>
            <i>Enum</i>: Service, <br/>
            <i>Default</i>: Service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name of the service/target that is being referred to. e.g. name of the service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          weight as an integer between 0 and 256, default 100, that specifies the target's relative weight
against other target reference objects. 0 suppresses requests to this backend.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 256<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec.port
<sup><sup>[↩ Parent](#grafanaspecroutespec)</sup></sup>



RoutePort defines a port mapping from a router to an endpoint in the service endpoints.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>targetPort</b></td>
        <td>int or string</td>
        <td>
          The target port on pods selected by the service this route points to.
If this is a string, it will be looked up as a named port in the target
endpoints port list. Required<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec.tls
<sup><sup>[↩ Parent](#grafanaspecroutespec)</sup></sup>



TLSConfig defines config used to secure a route and provide termination

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>termination</b></td>
        <td>enum</td>
        <td>
          termination indicates termination type.

* edge - TLS termination is done by the router and http is used to communicate with the backend (default)
* passthrough - Traffic is sent straight to the destination without the router providing TLS termination
* reencrypt - TLS termination is done by the router and https is used to communicate with the backend

Note: passthrough termination is incompatible with httpHeader actions<br/>
          <br/>
            <i>Enum</i>: edge, reencrypt, passthrough<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>caCertificate</b></td>
        <td>string</td>
        <td>
          caCertificate provides the cert authority certificate contents<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>certificate</b></td>
        <td>string</td>
        <td>
          certificate provides certificate contents. This should be a single serving certificate, not a certificate
chain. Do not include a CA certificate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>destinationCACertificate</b></td>
        <td>string</td>
        <td>
          destinationCACertificate provides the contents of the ca certificate of the final destination.  When using reencrypt
termination this file should be provided in order to have routers use it for health checks on the secure connection.
If this field is not specified, the router may provide its own destination CA and perform hostname validation using
the short service name (service.namespace.svc), which allows infrastructure generated certificates to automatically
verify.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecroutespectlsexternalcertificate">externalCertificate</a></b></td>
        <td>object</td>
        <td>
          externalCertificate provides certificate contents as a secret reference.
This should be a single serving certificate, not a certificate
chain. Do not include a CA certificate. The secret referenced should
be present in the same namespace as that of the Route.
Forbidden when `certificate` is set.
The router service account needs to be granted with read-only access to this secret,
please refer to openshift docs for additional details.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>insecureEdgeTerminationPolicy</b></td>
        <td>enum</td>
        <td>
          insecureEdgeTerminationPolicy indicates the desired behavior for insecure connections to a route. While
each router may make its own decisions on which ports to expose, this is normally port 80.

If a route does not specify insecureEdgeTerminationPolicy, then the default behavior is "None".

* Allow - traffic is sent to the server on the insecure port (edge/reencrypt terminations only).

* None - no traffic is allowed on the insecure port (default).

* Redirect - clients are redirected to the secure port.<br/>
          <br/>
            <i>Enum</i>: Allow, None, Redirect, <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>key</b></td>
        <td>string</td>
        <td>
          key provides key file contents<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec.tls.externalCertificate
<sup><sup>[↩ Parent](#grafanaspecroutespectls)</sup></sup>



externalCertificate provides certificate contents as a secret reference.
This should be a single serving certificate, not a certificate
chain. Do not include a CA certificate. The secret referenced should
be present in the same namespace as that of the Route.
Forbidden when `certificate` is set.
The router service account needs to be granted with read-only access to this secret,
please refer to openshift docs for additional details.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.route.spec.to
<sup><sup>[↩ Parent](#grafanaspecroutespec)</sup></sup>



RouteTargetReference specifies the target that resolve into endpoints. Only the 'Service'
kind is allowed. Use 'weight' field to emphasize one over others.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>kind</b></td>
        <td>enum</td>
        <td>
          The kind of target that the route is referring to. Currently, only 'Service' is allowed<br/>
          <br/>
            <i>Enum</i>: Service, <br/>
            <i>Default</i>: Service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          name of the service/target that is being referred to. e.g. name of the service<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>weight</b></td>
        <td>integer</td>
        <td>
          weight as an integer between 0 and 256, default 100, that specifies the target's relative weight
against other target reference objects. 0 suppresses requests to this backend.<br/>
          <br/>
            <i>Format</i>: int32<br/>
            <i>Default</i>: 100<br/>
            <i>Minimum</i>: 0<br/>
            <i>Maximum</i>: 256<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



Service sets how the service object should look like with your grafana instance, contains a number of defaults.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecservicemetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecservicespec">spec</a></b></td>
        <td>object</td>
        <td>
          ServiceSpec describes the attributes that a user creates on a service.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service.metadata
<sup><sup>[↩ Parent](#grafanaspecservice)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service.spec
<sup><sup>[↩ Parent](#grafanaspecservice)</sup></sup>



ServiceSpec describes the attributes that a user creates on a service.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>allocateLoadBalancerNodePorts</b></td>
        <td>boolean</td>
        <td>
          allocateLoadBalancerNodePorts defines if NodePorts will be automatically
allocated for services with type LoadBalancer.  Default is "true". It
may be set to "false" if the cluster load-balancer does not rely on
NodePorts.  If the caller requests specific NodePorts (by specifying a
value), those requests will be respected, regardless of this field.
This field may only be set for services with type LoadBalancer and will
be cleared if the type is changed to any other type.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusterIP</b></td>
        <td>string</td>
        <td>
          clusterIP is the IP address of the service and is usually assigned
randomly. If an address is specified manually, is in-range (as per
system configuration), and is not in use, it will be allocated to the
service; otherwise creation of the service will fail. This field may not
be changed through updates unless the type field is also being changed
to ExternalName (which requires this field to be blank) or the type
field is being changed from ExternalName (in which case this field may
optionally be specified, as describe above).  Valid values are "None",
empty string (""), or a valid IP address. Setting this to "None" makes a
"headless service" (no virtual IP), which is useful when direct endpoint
connections are preferred and proxying is not required.  Only applies to
types ClusterIP, NodePort, and LoadBalancer. If this field is specified
when creating a Service of type ExternalName, creation will fail. This
field will be wiped when updating a Service to type ExternalName.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>clusterIPs</b></td>
        <td>[]string</td>
        <td>
          ClusterIPs is a list of IP addresses assigned to this service, and are
usually assigned randomly.  If an address is specified manually, is
in-range (as per system configuration), and is not in use, it will be
allocated to the service; otherwise creation of the service will fail.
This field may not be changed through updates unless the type field is
also being changed to ExternalName (which requires this field to be
empty) or the type field is being changed from ExternalName (in which
case this field may optionally be specified, as describe above).  Valid
values are "None", empty string (""), or a valid IP address.  Setting
this to "None" makes a "headless service" (no virtual IP), which is
useful when direct endpoint connections are preferred and proxying is
not required.  Only applies to types ClusterIP, NodePort, and
LoadBalancer. If this field is specified when creating a Service of type
ExternalName, creation will fail. This field will be wiped when updating
a Service to type ExternalName.  If this field is not specified, it will
be initialized from the clusterIP field.  If this field is specified,
clients must ensure that clusterIPs[0] and clusterIP have the same
value.

This field may hold a maximum of two entries (dual-stack IPs, in either order).
These IPs must correspond to the values of the ipFamilies field. Both
clusterIPs and ipFamilies are governed by the ipFamilyPolicy field.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalIPs</b></td>
        <td>[]string</td>
        <td>
          externalIPs is a list of IP addresses for which nodes in the cluster
will also accept traffic for this service.  These IPs are not managed by
Kubernetes.  The user is responsible for ensuring that traffic arrives
at a node with this IP.  A common example is external load-balancers
that are not part of the Kubernetes system.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalName</b></td>
        <td>string</td>
        <td>
          externalName is the external reference that discovery mechanisms will
return as an alias for this service (e.g. a DNS CNAME record). No
proxying will be involved.  Must be a lowercase RFC-1123 hostname
(https://tools.ietf.org/html/rfc1123) and requires `type` to be "ExternalName".<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalTrafficPolicy</b></td>
        <td>string</td>
        <td>
          externalTrafficPolicy describes how nodes distribute service traffic they
receive on one of the Service's "externally-facing" addresses (NodePorts,
ExternalIPs, and LoadBalancer IPs). If set to "Local", the proxy will configure
the service in a way that assumes that external load balancers will take care
of balancing the service traffic between nodes, and so each node will deliver
traffic only to the node-local endpoints of the service, without masquerading
the client source IP. (Traffic mistakenly sent to a node with no endpoints will
be dropped.) The default value, "Cluster", uses the standard behavior of
routing to all endpoints evenly (possibly modified by topology and other
features). Note that traffic sent to an External IP or LoadBalancer IP from
within the cluster will always get "Cluster" semantics, but clients sending to
a NodePort from within the cluster may need to take traffic policy into account
when picking a node.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>healthCheckNodePort</b></td>
        <td>integer</td>
        <td>
          healthCheckNodePort specifies the healthcheck nodePort for the service.
This only applies when type is set to LoadBalancer and
externalTrafficPolicy is set to Local. If a value is specified, is
in-range, and is not in use, it will be used.  If not specified, a value
will be automatically allocated.  External systems (e.g. load-balancers)
can use this port to determine if a given node holds endpoints for this
service or not.  If this field is specified when creating a Service
which does not need it, creation will fail. This field will be wiped
when updating a Service to no longer need it (e.g. changing type).
This field cannot be updated once set.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>internalTrafficPolicy</b></td>
        <td>string</td>
        <td>
          InternalTrafficPolicy describes how nodes distribute service traffic they
receive on the ClusterIP. If set to "Local", the proxy will assume that pods
only want to talk to endpoints of the service on the same node as the pod,
dropping the traffic if there are no local endpoints. The default value,
"Cluster", uses the standard behavior of routing to all endpoints evenly
(possibly modified by topology and other features).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ipFamilies</b></td>
        <td>[]string</td>
        <td>
          IPFamilies is a list of IP families (e.g. IPv4, IPv6) assigned to this
service. This field is usually assigned automatically based on cluster
configuration and the ipFamilyPolicy field. If this field is specified
manually, the requested family is available in the cluster,
and ipFamilyPolicy allows it, it will be used; otherwise creation of
the service will fail. This field is conditionally mutable: it allows
for adding or removing a secondary IP family, but it does not allow
changing the primary IP family of the Service. Valid values are "IPv4"
and "IPv6".  This field only applies to Services of types ClusterIP,
NodePort, and LoadBalancer, and does apply to "headless" services.
This field will be wiped when updating a Service to type ExternalName.

This field may hold a maximum of two entries (dual-stack families, in
either order).  These families must correspond to the values of the
clusterIPs field, if specified. Both clusterIPs and ipFamilies are
governed by the ipFamilyPolicy field.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>ipFamilyPolicy</b></td>
        <td>string</td>
        <td>
          IPFamilyPolicy represents the dual-stack-ness requested or required by
this Service. If there is no value provided, then this field will be set
to SingleStack. Services can be "SingleStack" (a single IP family),
"PreferDualStack" (two IP families on dual-stack configured clusters or
a single IP family on single-stack clusters), or "RequireDualStack"
(two IP families on dual-stack configured clusters, otherwise fail). The
ipFamilies and clusterIPs fields depend on the value of this field. This
field will be wiped when updating a service to type ExternalName.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerClass</b></td>
        <td>string</td>
        <td>
          loadBalancerClass is the class of the load balancer implementation this Service belongs to.
If specified, the value of this field must be a label-style identifier, with an optional prefix,
e.g. "internal-vip" or "example.com/internal-vip". Unprefixed names are reserved for end-users.
This field can only be set when the Service type is 'LoadBalancer'. If not set, the default load
balancer implementation is used, today this is typically done through the cloud provider integration,
but should apply for any default implementation. If set, it is assumed that a load balancer
implementation is watching for Services with a matching class. Any default load balancer
implementation (e.g. cloud providers) should ignore Services that set this field.
This field can only be set when creating or updating a Service to type 'LoadBalancer'.
Once set, it can not be changed. This field will be wiped when a service is updated to a non 'LoadBalancer' type.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerIP</b></td>
        <td>string</td>
        <td>
          Only applies to Service Type: LoadBalancer.
This feature depends on whether the underlying cloud-provider supports specifying
the loadBalancerIP when a load balancer is created.
This field will be ignored if the cloud-provider does not support the feature.
Deprecated: This field was under-specified and its meaning varies across implementations.
Using it is non-portable and it may not support dual-stack.
Users are encouraged to use implementation-specific annotations when available.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>loadBalancerSourceRanges</b></td>
        <td>[]string</td>
        <td>
          If specified and supported by the platform, this will restrict traffic through the cloud-provider
load-balancer will be restricted to the specified client IPs. This field will be ignored if the
cloud-provider does not support the feature."
More info: https://kubernetes.io/docs/tasks/access-application-cluster/create-external-load-balancer/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecservicespecportsindex">ports</a></b></td>
        <td>[]object</td>
        <td>
          The list of ports that are exposed by this service.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>publishNotReadyAddresses</b></td>
        <td>boolean</td>
        <td>
          publishNotReadyAddresses indicates that any agent which deals with endpoints for this
Service should disregard any indications of ready/not-ready.
The primary use case for setting this field is for a StatefulSet's Headless Service to
propagate SRV DNS records for its Pods for the purpose of peer discovery.
The Kubernetes controllers that generate Endpoints and EndpointSlice resources for
Services interpret this to mean that all endpoints are considered "ready" even if the
Pods themselves are not. Agents which consume only Kubernetes generated endpoints
through the Endpoints or EndpointSlice resources can safely assume this behavior.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>selector</b></td>
        <td>map[string]string</td>
        <td>
          Route service traffic to pods with label keys and values matching this
selector. If empty or not present, the service is assumed to have an
external process managing its endpoints, which Kubernetes will not
modify. Only applies to types ClusterIP, NodePort, and LoadBalancer.
Ignored if type is ExternalName.
More info: https://kubernetes.io/docs/concepts/services-networking/service/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>sessionAffinity</b></td>
        <td>string</td>
        <td>
          Supports "ClientIP" and "None". Used to maintain session affinity.
Enable client IP based session affinity.
Must be ClientIP or None.
Defaults to None.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecservicespecsessionaffinityconfig">sessionAffinityConfig</a></b></td>
        <td>object</td>
        <td>
          sessionAffinityConfig contains the configurations of session affinity.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>trafficDistribution</b></td>
        <td>string</td>
        <td>
          TrafficDistribution offers a way to express preferences for how traffic
is distributed to Service endpoints. Implementations can use this field
as a hint, but are not required to guarantee strict adherence. If the
field is not set, the implementation will apply its default routing
strategy. If set to "PreferClose", implementations should prioritize
endpoints that are in the same zone.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type determines how the Service is exposed. Defaults to ClusterIP. Valid
options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
"ClusterIP" allocates a cluster-internal IP address for load-balancing
to endpoints. Endpoints are determined by the selector or if that is not
specified, by manual construction of an Endpoints object or
EndpointSlice objects. If clusterIP is "None", no virtual IP is
allocated and the endpoints are published as a set of endpoints rather
than a virtual IP.
"NodePort" builds on ClusterIP and allocates a port on every node which
routes to the same endpoints as the clusterIP.
"LoadBalancer" builds on NodePort and creates an external load-balancer
(if supported in the current cloud) which routes to the same endpoints
as the clusterIP.
"ExternalName" aliases this service to the specified externalName.
Several other fields do not apply to ExternalName services.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service.spec.ports[index]
<sup><sup>[↩ Parent](#grafanaspecservicespec)</sup></sup>



ServicePort contains information on service's port.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>port</b></td>
        <td>integer</td>
        <td>
          The port that will be exposed by this service.<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>appProtocol</b></td>
        <td>string</td>
        <td>
          The application protocol for this port.
This is used as a hint for implementations to offer richer behavior for protocols that they understand.
This field follows standard Kubernetes label syntax.
Valid values are either:

* Un-prefixed protocol names - reserved for IANA standard service names (as per
RFC-6335 and https://www.iana.org/assignments/service-names).

* Kubernetes-defined prefixed names:
  * 'kubernetes.io/h2c' - HTTP/2 prior knowledge over cleartext as described in https://www.rfc-editor.org/rfc/rfc9113.html#name-starting-http-2-with-prior-
  * 'kubernetes.io/ws'  - WebSocket over cleartext as described in https://www.rfc-editor.org/rfc/rfc6455
  * 'kubernetes.io/wss' - WebSocket over TLS as described in https://www.rfc-editor.org/rfc/rfc6455

* Other protocols should use implementation-defined prefixed names such as
mycompany.com/my-custom-protocol.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          The name of this port within the service. This must be a DNS_LABEL.
All ports within a ServiceSpec must have unique names. When considering
the endpoints for a Service, this must match the 'name' field in the
EndpointPort.
Optional if only one ServicePort is defined on this service.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>nodePort</b></td>
        <td>integer</td>
        <td>
          The port on each node on which this service is exposed when type is
NodePort or LoadBalancer.  Usually assigned by the system. If a value is
specified, in-range, and not in use it will be used, otherwise the
operation will fail.  If not specified, a port will be allocated if this
Service requires one.  If this field is specified when creating a
Service which does not need it, creation will fail. This field will be
wiped when updating a Service to no longer need it (e.g. changing type
from NodePort to ClusterIP).
More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>protocol</b></td>
        <td>string</td>
        <td>
          The IP protocol for this port. Supports "TCP", "UDP", and "SCTP".
Default is TCP.<br/>
          <br/>
            <i>Default</i>: TCP<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>targetPort</b></td>
        <td>int or string</td>
        <td>
          Number or name of the port to access on the pods targeted by the service.
Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
If this is a string, it will be looked up as a named port in the
target Pod's container ports. If this is not specified, the value
of the 'port' field is used (an identity map).
This field is ignored for services with clusterIP=None, and should be
omitted or set equal to the 'port' field.
More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service.spec.sessionAffinityConfig
<sup><sup>[↩ Parent](#grafanaspecservicespec)</sup></sup>



sessionAffinityConfig contains the configurations of session affinity.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaspecservicespecsessionaffinityconfigclientip">clientIP</a></b></td>
        <td>object</td>
        <td>
          clientIP contains the configurations of Client IP based session affinity.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.service.spec.sessionAffinityConfig.clientIP
<sup><sup>[↩ Parent](#grafanaspecservicespecsessionaffinityconfig)</sup></sup>



clientIP contains the configurations of Client IP based session affinity.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>timeoutSeconds</b></td>
        <td>integer</td>
        <td>
          timeoutSeconds specifies the seconds of ClientIP type session sticky time.
The value must be >0 && <=86400(for 1 day) if ServiceAffinity == "ClientIP".
Default value is 10800(for 3 hours).<br/>
          <br/>
            <i>Format</i>: int32<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.serviceAccount
<sup><sup>[↩ Parent](#grafanaspec)</sup></sup>



ServiceAccount sets how the ServiceAccount object should look like with your grafana instance, contains a number of defaults.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>automountServiceAccountToken</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecserviceaccountimagepullsecretsindex">imagePullSecrets</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecserviceaccountmetadata">metadata</a></b></td>
        <td>object</td>
        <td>
          ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaspecserviceaccountsecretsindex">secrets</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.serviceAccount.imagePullSecrets[index]
<sup><sup>[↩ Parent](#grafanaspecserviceaccount)</sup></sup>



LocalObjectReference contains enough information to let you locate the
referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
This field is effectively required, but due to backwards compatibility is
allowed to be empty. Instances of this type with an empty value here are
almost certainly wrong.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
          <br/>
            <i>Default</i>: <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.serviceAccount.metadata
<sup><sup>[↩ Parent](#grafanaspecserviceaccount)</sup></sup>



ObjectMeta contains only a [subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta).

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>annotations</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>labels</b></td>
        <td>map[string]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.spec.serviceAccount.secrets[index]
<sup><sup>[↩ Parent](#grafanaspecserviceaccount)</sup></sup>



ObjectReference contains enough information to let you inspect or modify the referred object.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>apiVersion</b></td>
        <td>string</td>
        <td>
          API version of the referent.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>fieldPath</b></td>
        <td>string</td>
        <td>
          If referring to a piece of an object instead of an entire object, this string
should contain a valid JSON/Go field access statement, such as desiredState.manifest.containers[2].
For example, if the object reference is to a container within a pod, this would take on a value like:
"spec.containers{name}" (where "name" refers to the name of the container that triggered
the event) or if no container name is specified "spec.containers[2]" (container with
index 2 in this pod). This syntax is chosen only to have some well-defined way of
referencing a part of an object.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>kind</b></td>
        <td>string</td>
        <td>
          Kind of the referent.
More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          Namespace of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resourceVersion</b></td>
        <td>string</td>
        <td>
          Specific resourceVersion to which this reference is made, if any.
More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>uid</b></td>
        <td>string</td>
        <td>
          UID of the referent.
More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#uids<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.status
<sup><sup>[↩ Parent](#grafana)</sup></sup>



GrafanaStatus defines the observed state of Grafana

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>adminUrl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>alertRuleGroups</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanastatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>contactPoints</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>dashboards</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>datasources</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>folders</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastMessage</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>libraryPanels</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>muteTimings</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>notificationTemplates</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>serviceaccounts</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stage</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>stageStatus</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Grafana.status.conditions[index]
<sup><sup>[↩ Parent](#grafanastatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## GrafanaServiceAccount
<sup><sup>[↩ Parent](#grafanaintegreatlyorgv1beta1 )</sup></sup>






GrafanaServiceAccount is the Schema for the grafanaserviceaccounts API

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>grafana.integreatly.org/v1beta1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>GrafanaServiceAccount</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountspec">spec</a></b></td>
        <td>object</td>
        <td>
          GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountstatus">status</a></b></td>
        <td>object</td>
        <td>
          GrafanaServiceAccountStatus defines the observed state of a GrafanaServiceAccount<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.spec
<sup><sup>[↩ Parent](#grafanaserviceaccount)</sup></sup>



GrafanaServiceAccountSpec defines the desired state of a GrafanaServiceAccount.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>instanceName</b></td>
        <td>string</td>
        <td>
          Name of the Grafana instance to create the service account for<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.instanceName is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the service account in Grafana<br/>
          <br/>
            <i>Validations</i>:<li>self == oldSelf: spec.name is immutable</li>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>enum</td>
        <td>
          Role of the service account (Viewer, Editor, Admin)<br/>
          <br/>
            <i>Enum</i>: Viewer, Editor, Admin<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>isDisabled</b></td>
        <td>boolean</td>
        <td>
          Whether the service account is disabled<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>resyncPeriod</b></td>
        <td>string</td>
        <td>
          How often the resource is synced, defaults to 10m0s if not set<br/>
          <br/>
            <i>Validations</i>:<li>duration(self) > duration('0s'): spec.resyncPeriod must be greater than 0</li>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>suspend</b></td>
        <td>boolean</td>
        <td>
          Suspend pauses reconciliation of the service account<br/>
          <br/>
            <i>Default</i>: false<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountspectokensindex">tokens</a></b></td>
        <td>[]object</td>
        <td>
          Tokens to create for the service account<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.spec.tokens[index]
<sup><sup>[↩ Parent](#grafanaserviceaccountspec)</sup></sup>



GrafanaServiceAccountTokenSpec defines a token for a service account

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the token<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>expires</b></td>
        <td>string</td>
        <td>
          Expiration date of the token. If not set, the token never expires<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>secretName</b></td>
        <td>string</td>
        <td>
          Name of the secret to store the token. If not set, a name will be generated<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.status
<sup><sup>[↩ Parent](#grafanaserviceaccount)</sup></sup>



GrafanaServiceAccountStatus defines the observed state of a GrafanaServiceAccount

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#grafanaserviceaccountstatusaccount">account</a></b></td>
        <td>object</td>
        <td>
          Info contains the Grafana service account information<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountstatusconditionsindex">conditions</a></b></td>
        <td>[]object</td>
        <td>
          Results when synchonizing resource with Grafana instances<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastResync</b></td>
        <td>string</td>
        <td>
          Last time the resource was synchronized with Grafana instances<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.status.account
<sup><sup>[↩ Parent](#grafanaserviceaccountstatus)</sup></sup>



Info contains the Grafana service account information

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          ID of the service account in Grafana<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>isDisabled</b></td>
        <td>boolean</td>
        <td>
          IsDisabled indicates if the service account is disabled<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>login</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>role</b></td>
        <td>string</td>
        <td>
          Role is the Grafana role for the service account (Viewer, Editor, Admin)<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountstatusaccounttokensindex">tokens</a></b></td>
        <td>[]object</td>
        <td>
          Information about tokens<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.status.account.tokens[index]
<sup><sup>[↩ Parent](#grafanaserviceaccountstatusaccount)</sup></sup>



GrafanaServiceAccountTokenStatus describes a token created in Grafana.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>id</b></td>
        <td>integer</td>
        <td>
          ID of the token in Grafana<br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>expires</b></td>
        <td>string</td>
        <td>
          Expiration time of the token
N.B. There's possible discrepancy with the expiration time in spec
It happens because Grafana API accepts TTL in seconds then calculates the expiration time against the current time<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#grafanaserviceaccountstatusaccounttokensindexsecret">secret</a></b></td>
        <td>object</td>
        <td>
          Name of the secret containing the token<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.status.account.tokens[index].secret
<sup><sup>[↩ Parent](#grafanaserviceaccountstatusaccounttokensindex)</sup></sup>



Name of the secret containing the token

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>namespace</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### GrafanaServiceAccount.status.conditions[index]
<sup><sup>[↩ Parent](#grafanaserviceaccountstatus)</sup></sup>



Condition contains details for one aspect of the current state of this API Resource.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>lastTransitionTime</b></td>
        <td>string</td>
        <td>
          lastTransitionTime is the last time the condition transitioned from one status to another.
This should be when the underlying condition changed.  If that is not known, then using the time when the API field changed is acceptable.<br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>message</b></td>
        <td>string</td>
        <td>
          message is a human readable message indicating details about the transition.
This may be an empty string.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>reason</b></td>
        <td>string</td>
        <td>
          reason contains a programmatic identifier indicating the reason for the condition's last transition.
Producers of specific condition types may define expected values and meanings for this field,
and whether the values are considered a guaranteed API.
The value should be a CamelCase string.
This field may not be empty.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>enum</td>
        <td>
          status of the condition, one of True, False, Unknown.<br/>
          <br/>
            <i>Enum</i>: True, False, Unknown<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>type</b></td>
        <td>string</td>
        <td>
          type of condition in CamelCase or in foo.example.com/CamelCase.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>observedGeneration</b></td>
        <td>integer</td>
        <td>
          observedGeneration represents the .metadata.generation that the condition was set based upon.
For instance, if .metadata.generation is currently 12, but the .status.conditions[x].observedGeneration is 9, the condition is out of date
with respect to the current state of the instance.<br/>
          <br/>
            <i>Format</i>: int64<br/>
            <i>Minimum</i>: 0<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>
