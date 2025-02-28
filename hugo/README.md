# Hugo

## Installing and running Hugo

When installing hugo install the `extended` edition at version `v0.134.0`. You
can find it on the [Hugo release
page](https://github.com/gohugoio/hugo/releases/tag/v0.134.0) or build it with a
working Go toolchain:

```shell
CGO_ENABLED=1 go install -tags extended github.com/gohugoio/hugo@v0.134.0
```

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
