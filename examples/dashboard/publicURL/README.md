---
title: "Public URL"
---

to allow access the dashboard on the current network.

As with normal publicDashboards, the accessToken, annotations, and time selection can be controlled.

The below shows the default values, but an empty object `.spec.publicSharing={}` is enough to enable the sharing.

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

{{% alert title="Note" color="primary" %}}
Grafana places strict requirements on the format of the accessToken.
It must be 33 hexadecimal characters and a valid uuid.
In effect, it can be a normal uuid without dashes `-`
`9da54597-3f48-435a-93eb-89aab761958d -> 9da545973f48435a93eb89aab761958d`
{{% /alert %}}
