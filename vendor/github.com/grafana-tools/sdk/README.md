<!--*- mode:markdown -*-->

# Grafana SDK [![Go Report Card](https://goreportcard.com/badge/github.com/grafana-tools/sdk)](https://goreportcard.com/report/github.com/grafana-tools/sdk)

SDK for Go language offers a library for interacting with
[Grafana](http://grafana.org) server from Go applications.  It
realizes many of
[HTTP REST API](http://docs.grafana.org/reference/http_api) calls for
administration, client, organizations. Beside of them it allows
creating of Grafana objects (dashboards, panels, datasources) locally
and manipulating them for constructing dashboards programmatically.
It would be helpful for massive operations on a large set of
Grafana objects.

It was made foremost for
[autograf](https://github.com/grafana-tools/autograf) project but
later separated from it and moved to this new repository because the
library is useful per se.

## Library design principles

1. SDK offers client functionality so it covers Grafana REST API with
   its requests and responses as close as possible.
1. SDK maps Grafana objects (dashboard, row, panel, datasource) to
   similar Go structures but not follows exactly all Grafana
   abstractions.
1. It doesn't use any logger, instead API functions could return errors
   where it need.
1. Prefere no external deps except Go stdlib.
1. Cover SDK calls with unit tests.

## Examples [![GoDoc](https://godoc.org/github.com/grafana-tools/sdk?status.svg)](https://godoc.org/github.com/grafana-tools/sdk)

```go
	board := sdk.NewBoard("Sample dashboard title")
	board.ID = 1
	row1 := board.AddRow("Sample row title")
	row1.Add(sdk.NewGraph("Sample graph"))
	graph := sdk.NewGraph("Sample graph 2")
	target := sdk.Target{
		RefID:      "A",
		Datasource: "Sample Source 1",
		Expr:       "sample request 1"}
	graph.AddTarget(&target)
	row1.Add(graph)
	grafanaURL := "http://grafana.host"
	c := sdk.NewClient(grafanaURL, "grafana-api-key", sdk.DefaultHTTPClient)
	response, err := c.SetDashboard(board, false)
	err != nil {
		fmt.Printf("error on uploading dashboard %s", board.Title)
	} else {
		fmt.Printf("dashboard URL: %v", grafanaURL+*resp.URL)
	}
```

The library includes several demo apps for showing API usage:

* [backup-dashboards](cmd/backup-dashboards) — saves all your dashboards as JSON-files.
* [backup-datasources](cmd/backup-datasources) — saves all your datasources as JSON-files.
* [import-datasources](cmd/import-datasources) — imports datasources from JSON-files.
* [import-dashboards](cmd/import-dashboards) — imports dashboards from JSON-files.

You need Grafana API key with _admin rights_ for using these utilities.

## Installation [![Build Status](https://travis-ci.org/grafana-tools/sdk.svg?branch=master)](https://travis-ci.org/grafana-tools/sdk)

Of course Go development environment should be set up first. Then:

	go get github.com/grafana-tools/sdk

Dependency packages have included into
distro. [govendor](https://github.com/kardianos/govendor) utility used
for vendoring.  The single dependency now is:

	go get github.com/gosimple/slug

The "slugify" for URLs is a simple task but this package used in
Grafana server so it used in the SDK for the compatibility reasons.

## Status of REST API realization [![Coverage Status](https://coveralls.io/repos/github/grafana-tools/sdk/badge.svg?branch=master)](https://coveralls.io/github/grafana-tools/sdk?branch=master)

Work on full API implementation still in progress. Currently
implemented only create/update/delete operations for dashboards and
datasources. State of support for misc API parts noted below.

| API                    | Status                    |
|------------------------|---------------------------|
| Authorization          | API tokens and Basic Auth |
| Dashboards             | partially                 |
| Datasources            | +                         |
| Organization (current) | partially                 |
| Organizations          | partially                 |
| Users                  | partially                 |
| User (actual)          | partially                 |
| Snapshots              | -                         |
| Frontend settings      | -                         |
| Admin                  | partially                 |

There is no exact roadmap, the library tries to reflect latest Grafana API.

I still have interest to this library development but not always have
time for it. So I gladly accept new contributions. Drop an issue or
[contact me](grafov@gmail.com).

## Licence

Distributed under Apache v2.0. All rights belong to the SDK
authors. There is no authors list yet, you can see the full list of
the contributors in the git history. Official repository is
https://github.com/grafana-tools/sdk

## Collection of Grafana tools in Golang

* [github.com/nytm/go-grafana-api](https://github.com/nytm/go-grafana-api) — a golang client of Grafana project currently that realizes parts of the REST API, used for the Grafana Terraform provider.
* [github.com/adejoux/grafanaclient](https://github.com/adejoux/grafanaclient) — API to manage Grafana 2.0 datasources and dashboards. It lacks features from 2.5 and later Grafana versions.
* [github.com/mgit-at/grafana-backup](https://github.com/mgit-at/grafana-backup) — just saves dashboards localy.
* [github.com/raintank/memo](https://github.com/raintank/memo) — send slack mentions to Grafana annotations.
* [github.com/retzkek/grafctl](https://github.com/retzkek/grafctl) — backup/restore/track dashboards with git.
