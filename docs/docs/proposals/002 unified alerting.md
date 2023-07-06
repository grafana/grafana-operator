---
title: "Support for Grafana's Unified Alerting"
linkTitle: "Unified Alerting Support"
---

## Summary

Introduce the ability to configure Grafana's new alerting system using Grafana Operator.

This document contains the complete design as well as an implementation proposal required for extending Grafana Operator to support Unified Alerting.

Grafana Operator should support the full breadth of unified alerting when complete, however the design and implementation can progress iteratively with individual components becoming available independently of oneanother. 

The following components are considered in scope for this proposal:
- Alert rules
- Contact points
- Notification policies
- Mute timings
- Notification templates

## Info
_status: Design_

### Related Issues/Pull Requests

- [Issue 911](https://github.com/grafana-operator/grafana-operator/issues/911)

### Related Documentation

- [Grafana Unified Alerting](https://grafana.com/docs/grafana/latest/alerting/)
- [Grafana Alerting Provisioning API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/)
- [Grafana Alerting API (unstable)](https://editor.swagger.io/?url=https://raw.githubusercontent.com/grafana/grafana/main/pkg/services/ngalert/api/tooling/post.json)

## Motivation

Grafana's unified alerting provides numerous benefits over using a seperately managed AlertManager instance, not least providing an intuitive UI to visualise (and manage) the configuration. 

Grafana Operator can augment these benefits by providing a method for securely managing the configuration in a declarative way, allowing a unified config-as-code for all Grafana settings.

## Current State

The grafana operator currently does not support this feature set. 

## Proposal

The proposal covers the design of new custom resources and seperately a possible implementation plan. 

### Design

For each component an individual design is prepared, the extraction of items that are shareable is still possible.

#### Alert Rules
_status: Unstarted_ | [API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/#provisioned-alert-rule)

#### Contact points
_status: Unstarted_ | [API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/#embedded-contact-point)

#### Notification policies
_status: Unstarted_ | [API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/#route)

#### Mute timings
_status: Unstarted_ | [API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/#mute-time-interval)

#### Notification templates
_status: Unstarted_ | [API](https://grafana.com/docs/grafana/latest/developers/http_api/alerting_provisioning/#notification-template-content)


### Implementation Plan

#### API Availability

Grafana considers the provisioning of these components to be part of the stable API, as such there is no blocker regarding unstable APIs on Grafana's side. 

The general alerting API (not provisioning) is considered unstable, however this API covers features that are not required for the proposed design. Lastly, the legacy alerting API is deprecated and is not relevant to the unified alerting feature. 

#### Sequencing

Each component of the unified alerting feature can be implemented in a standalone fashion, allowing for a iterative approach that provides value with the first implemented component.

##### Lower Priority Components
As `alert rules` can be defined directly in prometheus in a k8s-native way utilising [prometheus-operator](https://prometheus-operator.dev/) and forwarding of alerts from prometheus to the Grafana built-in AlertManager should be a valid option this component is considered to provide lower value compared to the other components.

While `notification templates` provide value by allowing customisation of how alerts look, this feature is less of a requirement in building a functional alerting system compared to the other components.

## Verification

- Create integration tests for the initial provisioning of each component.
- Create integration tests for drift detection.

## Considered Alternatives

### Prometheus-Operator

Utilising [prometheus-operator](https://prometheus-operator.dev/) to manage grafana's unified alerting as an "external prometheus/alertmanager instance" is not possible as they do not support this feature. Future support from their side for this feature is unlikely as Prometheus/AlertManager itself does not provide an applicable management API (relying on file-based configuration). While alternative prometheus implementations (i.e. Cortex) do provide a more full-featured API, support from the OSS prometheus-operator is unlikely. 

While direct support for Grafana's unified alerting is unlikely, the design of the relevant components can be utilised in conjunction with Grafana's specification as inspiration for the proposed design.
