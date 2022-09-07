# Dashboard-Folder permissions

In Grafana, Dashboards inherit the permissions of the folder they're in - and they can't be more restrictive than those!
And in some cases, you'd want to restrict the visibility of Dashboards, e.g. to only developers.

## Example use-case

Say you have some Dashboards that shall be accessible for all departments in your company.
And some other Dashboards with infrastructure- and applications-stats, that only developers should be allowed to access.

You'd need to give all developers the role "editor" and everyone else the role "viewer".
(e.g. via corresponding groups, if you're using authentication via
[AAD-SSO](https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/azuread/))

Then configure at least two folders - one for the public Dashboards and one for the developers Dashboards.
See [deploy/examples/folders/](../deploy/examples/folders/) for `GrafanaFolder` example configurations - they cover exactly this use-case.

## Folder properties

As mentioned in the description for [Dashboards](./dashboards.md), you can create a folder implicitly by specifying a `customFolderName`.
But in order to set permissions, you have to create a `GrafanaFolder` custom resource as well.
(Of course you can also create a folder by just deploying a `GrafanaFolder` without referencing it in a `GrafanaDashboard` right away)

To get a quick overview of the GrafanaFolder you can also look at the [API docs](api.md).
The following properties are accepted in the `spec`:

* *title*: The displayname of the folder. It must match the `customFolderName` of the GrafanaDashboard in order to "assign" it.
* *permissions*: the __complete__ permissions for the folder. Any permission not listed here, will be removed from the folder.
  * *permissionLevel*: 1 == View; 2 == Edit; 4 == Admin
  * *permissionTarget*: The target-role (can be "Viewer" or "Editor" - no need for admins as they always have access)
  * *permissionTargetType*: currently only "role" is supported
