# Deploying the Grafana operator in minikube
Grafana operator is a relatively small and non-resource intensive operator, it can be deployed independently on minikube, which makes it very convenient for testing and development purposes, Minikubes can be quickly disposed of and restarted without much downtime. it runs locally and provides almost all the functionality required to run this operator "out of the box"

## Prerequisites
- [Minikube installed](https://kubernetes.io/docs/tasks/tools/install-minikube/)
- Kubectl / OC - For managing the minikube cluster
- Local clone of this repository

## Running the operator "Locally"
"Locally" in this case refers to running the operator in the CLI i.e. not as a deployment on the minikube cluster. for that see the next section.

1. Start the minikube cluster using `minikube start (--vm-driver=<driver>)` the part in the parentheses is optional. for more info read back through the minikube product documentation.
2. To prepare the local instance of grafana run `make cluster/prepare/local`, this will create a new instance of grafana in the grafana namespace (to edit this edit the NAMESPACE field in the makefile).
3. Run the operator with `make code/run` or `operator-sdk up local --namespace=grafana"
4. The ingress readiness probe will most likely fail, after a few seconds, from a new terminal window run `minikube addonsenable ingress`
5. The operator should now see the ingress and proceed with deployment. Optionally to confirm that the ingress is reachable `kubectl get ingress -n grafana` and ensure that the address is not empty.

To stop the operator `ctrl + c` in the terminal with the operator running.
## Running the operator on the minikube cluster
1. Start the minikube cluster using `minikube start (--vm-driver=<driver>)` the part in the parentheses is optional. for more info read back through the minikube product documentation.
2. Run the deployment with `make operator/deploy`, this will deploy the operator on the cluster and will be ready in a few seconds.

to stop the operator run `make operator/stop`, this will delete the operator deployment from the cluster.

to re-deploy the operator run `make operator/deploy`
## Cleanup
To cleanup the cluster from Grafana resources (only in the grafana namespace):
Run `make cluster/cleanup`

To completely remove the minikube cluster and all associated resources:
Run `minikube stop` and `minikube delete` after the minikube cluster has stopped.



