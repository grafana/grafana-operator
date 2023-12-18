---
author: "Edvin 'NissesSenap' Norling"
date: 2023-03-29
title: "Install Grafana-operator using Flux and Kustomize"
linkTitle: "Install Grafana-operator using Flux and Kustomize"
description: "How to install grafana-operator using Flux and Kustomize"
---

As a part of grafana-operator v5.0.0-rc1 we introduce Kustomize as a way of installing the operator.

To showcase this new feature, I thought why not use GitOps?

GitOps is a rather well-used term nowadays, but you can summarize it in 4 steps.

- Declarative
- Versioned and immutable
- Pulled automatically
- Continuously reconciled

To find out more about GitOps look at the CNCF GitOps working groups [documentation](https://opengitops.dev/).

In this case I have decided to use [Flux](https://fluxcd.io/) to deploy grafana-operator through GitOps, but there are other options.
For example [ArgoCD](https://argo-cd.readthedocs.io/) which is also a graduated CNCF project just like Flux.

This blog's focus is to showcase how you can use Kustomize to install the grafana-operator, I will take many shortcuts to keep this blog **simple**, read the official Flux documentation for best practices.

## What we will do?

We will install the grafana-operator using Flux and to manage our grafana instance and dashboard.

Assuming that you follow the instructions of this blog, your Flux fleet-infra repository will look something like this.

```.txt
├── clusters
│   └── my-cluster
│       ├── flux-system
│       │   ├── gotk-components.yaml
│       │   ├── gotk-sync.yaml
│       │   └── kustomization.yaml
│       ├── grafana
│       │   ├── dashboard.yaml
│       │   ├── grafana.yaml
│       │   └── kustomization.yaml
│       ├── grafana-operator.yaml
│       └── grafana.yaml
```

You can find all the files available to copy in the grafana-operator [repository](https://github.com/grafana/grafana-operator/tree/master/docs/blog/flux-gitops),
or you can just copy paste them from the blog.

## Prerequisite

In this example, I will use [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/) as my cluster because I think it's simple, but the walkthrough should work with any Kubernetes solution.
I will use GitHub for my GitOps repository but Flux supports multiple source control management (SCM) providers.

So what will you need?

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start)
- [Flux cli](https://fluxcd.io/flux/installation/)
- github repository
- github token

### Setup cluster

Create a Kind cluster with [ingress support](https://kind.sigs.k8s.io/docs/user/ingress/).

```shell
cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
  - containerPort: 443
    hostPort: 443
    protocol: TCP
EOF
```

When the cluster is up, install [ingress-nginx](https://github.com/kubernetes/ingress-nginx).

```shell
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml
# Wait for ingress-nginx to become ready
kubectl wait --namespace ingress-nginx \
  --for=condition=ready pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
```

### Bootstrap Flux

Read the official documentation on how to bootstrap your Flux setup, as I mentioned earlier, I will use [github](https://fluxcd.io/flux/cmd/flux_bootstrap_github/).

Before bootstrapping your cluster, you need to create a github [PAT](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token)(Personal Access Token) with enough access.

I have been unable to find exactly what the minimal access needed for the token, so I just created a PAT with all the access. **You** should not do the same.

```shell
export GITHUB_TOKEN=github_pat_soooLongAndSecretPAT
flux bootstrap github --owner=<github-username> --repository=fleet-infra --private=false --personal=true --path=clusters/my-cluster
```

This will create a repository called fleet-infra in your GitHub account, install Flux to your cluster and setup an autosync to GitHub.
You should now be able to view Flux running in the `fleet-infra` namespace.

## Install operator using Flux and Kustomize

We are using Flux to package our Kustomize files through OCI, and they are built and released just as our helm solution.

There are two ways of installing the operator, either with namespace access or cluster access,
take a look at our [documentation](https://grafana-operator.github.io/grafana-operator/docs/grafana/#where-should-the-operator-look-for-grafana-resources) for more information.
Add the following file to your Flux repo under `clusters/my-cluster/grafana-operator.yaml`.

{{< readfile file="flux-gitops/grafana-operator.yaml" code="true" lang="yaml" >}}

Depending on if you want to install cluster scoped or namespace scoped, you need to change the path.

## Install operator using Kustomize

If you want to install the grafana-operator without using GitOps, you can also download the generated artifact and install it manually.
For example, you can run the following Flux command to download the artifact and unpack it. Then you can run a normal kubectl apply command.

```shell
flux pull artifact oci://ghcr.io/grafana/kustomize/grafana-operator:{{<param version>}} -output ./grafana-opreator
```

But of course we recommend that you manage your grafana-operator installation through your GitOps solution, no matter if it's Flux or some other solution.

## Install Grafana

Okay great, we got grafana-operator installed, but don't we actually want to install a Grafana instance as well?

Let's setup a Grafana instance and some dashboard trough code.

Since we already setup our kind cluster with ingress-nginx I will use one of our basic HTTP examples.

Let's not make this more complicated than it has to be, you should have your secrets in some KMS/vault/sealed-secrets, SOAP or similar solution, but definitely not checked in to your GitOps repository non-encrypted.
But this is an example, so for this time let's create our admin and password secrets manual.

```shell
cat <<EOF | kubectl apply -f -
kind: Secret
apiVersion: v1
metadata:
  name: credentials
  namespace: grafana
stringData:
  GF_SECURITY_ADMIN_PASSWORD: secret
  GF_SECURITY_ADMIN_USER: root
type: Opaque
EOF
```

Now create our Grafana instance together with a very basic dashboard.

The Grafana instance should be stored in `clusters/my-cluster/grafana/grafana.yaml`

{{< readfile file="flux-gitops/grafana/grafana.yaml" code="true" lang="yaml" >}}

The Grafana dashboard should be stored in `clusters/my-cluster/grafana/dashboard.yaml`

{{< readfile file="flux-gitops/grafana/dashboard.yaml" code="true" lang="yaml" >}}

We also need a Kustomization file to `clusters/my-cluster/grafana/kustomization.yaml` to help Flux to find the files.

{{< readfile file="flux-gitops/grafana/kustomization.yaml" code="true" lang="yaml" >}}

And finally create the Flux kustomization file (yes, the naming is a bit confusing).
Lets store it in `clusters/my-cluster/grafana.yaml`

{{< readfile file="flux-gitops/grafana.yaml" code="true" lang="yaml" >}}

After pushing the changes to your Flux repository, the changes should be applied to your cluster.

If you don't want to wait for the automatic reconcile, you can run which will trigger a reconcile of the git repo.

```shell
flux reconcile source git flux-system -n flux-system
```

You should now be able to go to the Grafana URL that you defined, in my case [http://grafana.127.0.0.1.nip.io](http://grafana.127.0.0.1.nip.io)

And you should see the simple dashboard among your dashboards.

## Conclusion

We have used very basic setup of Flux to install grafana-operator to our cluster. When that was done, we installed a basic Grafana instance and a dashboard giving us dashboards as code.

The dashboard wasn't the prettiest, but it's an easy example.
Hopefully someone will make a more in depth blog on how to use grafana-operator to bootstrap entire clusters monitoring in the future.
