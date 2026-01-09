---
title: "JaaS"
linkTitle: "jaas"
---

A basic deployment of Grafana with a dashboard using [Jsonnet-as-a-Service](https://github.com/metio/jaas) (JaaS).

{{< readfile file="resources.yaml" code="true" lang="yaml" >}}

JaaS expects that your dashboards are packaged as OCI objects and mounts them using the Kubernetes OCI volume feature. In general, you need to perform the following steps to expose your Grafana dashboards with JaaS:

1. Write your dashboards
2. Publish OCI objects
3. Modify your JaaS instance
4. Define your `GrafanaDashboard` resource

## Write your dashboards

The JaaS Helm chart includes [Grafonnet](https://github.com/grafana/grafonnet) and [other required libraries](https://github.com/metio/jsonnet-oci-images) as OCI volumes to write Grafana dashboards using Jsonnet. However, since Jsonnet is a superset of JSON, you can use JaaS to deliver your JSON based dashboards as well. This allows you to export an existing dashboard using the Grafana UI, package it as an OCI object, and expose it through JaaS.

When using Jsonnet to define your dashboards, you will most likely want to use Grafonnet which is exposed like this in JaaS:

{{< readfile file="imports.jsonnet" code="true" lang="jsonnet" >}}

In general, it is highly recommend to use the `latest` version and control the actual version of Grafonnet you want to use by modifying your JaaS deployment. If you follow this recommendation, you only have to change the version of Grafonnet in one single place instead of touching all of your dashboards every time Grafonnet releases a new version. Likewise, you can add additional libraries you have written yourself to JaaS and import them just like the built-in Grafonnet library.

An entire dashboard in Jsonnet might look like this:

{{< readfile file="dashboard.jsonnet" code="true" lang="jsonnet" >}}

In case your dashboard definition gets longer and longer, you can extract parts of it into their own files and import them, e.g., the following example moves the panels from the above example in their own file:

{{< readfile file="dashboard-short.jsonnet" code="true" lang="jsonnet" >}}

If you need to parameterize parts of your dashboard, you can use Jsonnet top-level arguments (TLAs) like this:

{{< readfile file="dashboard-tla.jsonnet" code="true" lang="jsonnet" >}}

Specify parameters to your dashboard in the `GrafanaDashboard` resource like this:

{{< readfile file="dashboard-tla.yaml" code="true" lang="yaml" >}}

The above example shows that top-level arguments are exposed as URL query parameters in JaaS. Likewise, it is possible to use Jsonnet external variables (`std.extVar`) like this:

{{< readfile file="dashboard-extvar.jsonnet" code="true" lang="jsonnet" >}}

In order to define external variables, modify your JaaS deployment so that it contains `JAAS_EXT_VAR_*` environment variables, e.g., `JAAS_EXT_VAR_description` and `JAAS_EXT_VAR_editable` in the above example. Take a look at the [JaaS documentation](https://github.com/metio/jaas/tree/main/helm#defining-external-variables) on how to do this.

### Publish OCI objects

JaaS evaluates Jsonnet snippets and expects that each snippet contains a `main.jsonnet` file as its entrypoint. When writing dashboards, you can use the following directory structure:

```console
<your-repository>
├── main.jsonnet
├── datasources.libsonnet   # local library
└── panels.libsonnet        # local library
```

An example `Dockerfile` which packages your dashboard looks like this:

{{< readfile file="dashboard.dockerfile" code="true" lang="dockerfile" >}}

In case you have split your dashboard definition into multiple files, adjust as necessary so that the `/main.jsonnet` file can find them in the `Dockerfile` as well, e.g., by using relative imports and copying everything into the root folder.

When packaging Jsonnet libraries, you just need to make sure that you recreate the same folder structure that is used by [jsonnet-bundler](https://github.com/jsonnet-bundler/jsonnet-bundler/) in non-legacy mode, e.g., if your library is in a repository at https://git.example.com/my/jsonnet/library and all your Jsonnet files are in a `src/something` subdirectory of that repository, your `Dockerfile` for your Jsonnet library should look like this:

{{< readfile file="library.dockerfile" code="true" lang="dockerfile" >}}

If you follow this structure, you can use `jsonnet-bundler` locally to develop your dashboards and use them as-is from JaaS as well. In case you do not care about `jsonnet-bundler`, you are free to choose any structure you want.

There is no restriction on file names for libraries since it's up to you to import them in your dashboards, e.g., the above example library could be imported like this, assuming that there is a file called `main.libsonnet` in `src/something`:

{{< readfile file="library.jsonnet" code="true" lang="jsonnet" >}}

## Modify your JaaS instance

Once your dashboards and libraries are packages as OCI objects, add them to your JaaS deployment as explained in their documentation for [snippets](https://github.com/metio/jaas/tree/main/helm#adding-jsonnet-snippets) and [libraries](https://github.com/metio/jaas/tree/main/helm#adding-jsonnet-libraries).

Note that this does require [OCI volume](https://kubernetes.io/docs/tasks/configure-pod-container/image-volumes/) support in your Kubernetes cluster.

## Define your `GrafanaDashboard` resource

Finally, you can define your `GrafanaDashboard` resource like this:

{{< readfile file="dashboard.yaml" code="true" lang="yaml" >}}

The name `your-dashboard` must match of the snippets you added to JaaS, or otherwise you will get a 404 response.
