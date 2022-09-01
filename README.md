# Grafana Operator - Experimental Repository

This repository is the temporary home for v5 of the grafana-operator.
We will eventually merge the code from this repository back to the main repository.

As of now, we are re-writing massive chunks of the operator logic, to improve:
- Performance
- Reliability
- Maintainability
- Extensibility
- Testability
- Usability

The previous versions of the operator have some serious tech-debt issues, which effectively prevent community members that aren't massively
familiar with the project and/or its codebase from contributing features that they wish to see.
These previous versions, we're built on a "as-needed" basis, meaning that whatever was the fastest way to reach the desired feature, was the way
it was implemented. This lead to situations where controllers for different resources were using massively different logic, and features were added
wherever and however they could be made to work. 

The v5 version aims to re-focus the operator with a more thought out architecture and framework, that will work better, both for developers and users.
With certain standards and approaches, we can provide a better user experience through:
- Better designed Custom Resource Definitions (Upstream Grafana Native fields will be supported without having to whitelist them in the operator logic).
  - Thus, the upstream documentation can be followed to define the grafana-operator Custom Resources.
  - This also means a change in API versions for the resources, but we see this as a benefit, our previous mantra of maintaining a 
    seamless upgrade from version to version, limited us in the changes we wanted to make for a long time.
- A more streamlined Grafana resource management workflow, one that will be reflected across all controllers.
- Using an upstream Grafana API client (standardising our interactions with the Grafana API, moving away from bespoke logic).
- The use of a more up-to-date Operator-SDK version, making use of newer features.
- Implementing some proper testing.
- Cleaning and cutting down on code.
- Multi-instance AND Multi-namespace support!
