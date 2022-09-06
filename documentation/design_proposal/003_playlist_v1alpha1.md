# Design proposal 003 playlist v1alpha1

## Summary

Create a separate controller to manage the grafana playlist API.

## Info

status: Draft

## Motivation

Playlist is a core feature of grafana and it would be nice to support it's resources through a CRD.

## Out of scope

This design focuses on grafana-operator version 5.
This won't be implemented in version 4.

## Verification

- Create integration tests for the new CRD:s
- Add new e2e tests to cover the new CR.

## Proposal

Make the playlist struct look like below.

Follow the same grafana selector as the rest of the controllers.

```.go
// GrafanaPlayListSpec defines the desired state of GrafanaPlayList
type GrafanaPlayListSpec struct {
	// Name of the playlist
	Name string `json:"name"`
	// Interval how often the playlist should change dashboard
	Interval string `json:"interval"`
	// PlaylistDashboards is a list of dashboards that should be in the playlist
	PlaylistDashboards DashboardsList `json:"dashboards,omitempty"`

	// selects Grafanas for import
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
}

type PlaylistDashboards struct {
	// +kubebuilder:validation:Enum=dashboard_by_id;dashboard_by_tag
	Type  string `json:"type"`
	Value string `json:"value"`
	// The order of the dashboard in the playlist
	Order int64 `json:"order"`
	// The title of the dashboard in the playlist
	Title string `json:"title"`
}

type DashboardsList []PlaylistDashboards

```

### Example yaml

```.yaml
apiVersion: grafana.integreatly.org/v1beta1
kind: GrafanaPlayList
metadata:
  name: tv
spec:
  name: tv1
  interval: "5m"
  dashboards:
    - title: "dashbord-tag"
      order: 1
      type: dashboard_by_tag
      value: "tv-tag"
    - title: "dashboard-uid"
      order: 2
      type: dashboard_by_id
      value: "random-uid"
```

## Related issues

- [817 Grafana playlist controller](https://github.com/grafana-operator/grafana-operator/issues/817)

## References

- [Grafana playlist API docs](https://grafana.com/docs/grafana/latest/developers/http_api/playlist/)
