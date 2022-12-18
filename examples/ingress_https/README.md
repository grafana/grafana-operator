# ingress https

Assumes that you have [cert-manager](https://github.com/cert-manager/cert-manager) running in your cluster and have a ClusterIssuer called letsencrypt.

It also assumes that you have `ingressClassName: nginx`.

You can of course have added a certificate to a secret manually.
