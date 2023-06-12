# Hugo

## Installing and running Hugo

When installing hugo install the `extended` edition, you can easily find it on there github page.
We are currently using [0.111.2](https://github.com/gohugoio/hugo/releases/tag/v0.111.2) but it probably works with other versions as well.

To develop locally you need to also follow the docsy [pre-req](https://github.com/google/docsy#prerequisites).
But in most cases it should work to just run

```shell
cd hugo
# Install npm dependencies
npm ci
# Download hugo module
hugo mod get
```

To look at your changes with hot reload.

```shell
hugo serve
```

## Detecting broken links in documentation

To detect broken links in documentation using [muffet](https://github.com/raviqqe/muffet):

1.  Make sure that hugo is running and serving documentation as specified in the abbove steps.
2.  Run the following command in the hugo directory:

```shell
make detect-broken-links
```
