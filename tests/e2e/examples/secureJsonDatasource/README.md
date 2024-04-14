# Testing GrafanaDatasource secureJsonData

This test creates a GrafanaDatasource with a reference
to a secret (which is normally created by a serviceAccount)
and makes sure it's inserted correctly into
grafana.

## Step 00

This step creates a number of resources:
- Grafana (to create a new grafana)
- GrafanaDatasource (with secureJsonData and a secret)
- A thanos emulator pod, using netcat, with a service

## Step 01

This step starts a pod which query the grafana to test it's datasource,
which in turn forces the grafana to query thanos.

## Step 02

Verify in the log that grafana is happy with the response from
the datasource.

## Step 03

Verify in the log that grafana sent the authorization header with
the token.
