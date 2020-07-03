# Deploying the Grafana operator in minikube
Grafana operator is a relatively small and non-resource intensive operator, it can be deployed independently on minikube, which makes it very convenient for testing and development purposes, Minikubes can be quickly disposed of and restarted without much downtime. it runs locally and provides almost all the functionality required to run this operator "out of the box"

## Prerequisites
- [Minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- Kubectl / OC - For managing the minikube cluster
- Local clone of this repository

## Running the operator "Locally"
"Locally" in this case refers to running the operator in the CLI i.e. not as a deployment on the minikube cluster. for that see the next section.

1. start the minikube cluster using `minikube start (--vm-driver=<driver>)` the part in the parentheses is optional. for more info read back through the minikube product documentation


## Running the operator on minikube cluster


