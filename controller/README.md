# RestStrategy Controller

[![Go Report Card](https://goreportcard.com/badge/github.com/dnitsch/reststrategy/controller)](https://goreportcard.com/report/github.com/dnitsch/reststrategy/controller)

Custom Controller used to listen to changes on a specific object (CRD -> RestStrategy)

Everytime a set of changes is done on the types a client needs to be regenerated and unit tests updated.

## Useful resources

Kubernetes maintained [sample-controller](https://github.com/kubernetes/sample-controller) repo is a great reference.

Includes the below image, highlighting the area of responsibility between the client (client-go in this case the most mature, and best suited for concurrency) and user code (custom controller)

![CustomController!](https://raw.githubusercontent.com/kubernetes/sample-controller/ff730d68ab4ec1f5e502609829847a7e6c78c57f/docs/images/client-go-controller-interaction.jpeg)

## Helper resources

Debug controllers [video](https://morioh.com/p/b730fcc35f39)

A VSCode launch.json is shared in the controller dir.

## Notes

Build locally and test in cluster.

`docker build --build-arg REVISION=abcd1234 --build-arg VERSION=0.6.2 -t dnitsch/reststrategy-controller .`

## Deployment

Run in minikube - installation instructions [here](https://minikube.sigs.k8s.io/docs/start/).

You can deploy just the CRD when testing changes and run controller locally and have updated `kubecontext` e.g.:

```json
"args": [
    "--kubeconfig",
    "${env:HOME}/.kube/config",
    "--controllercount",
    "2",
    "--namespace",
    "reststrategy",
    "--loglevel",
    "debug"
]
```

## Secret Token

As Part of the orchstrator it does a token replace before calling the relevant service endpoint

> This way we are not storing secrets in etcd and calling a `kubectl describe reststrategy` will only fetch back token from the CRD stored in ETCD. 

The controller is using the [configmanager](https://github.com/dnitsch/configmanager) to perform token replacement, so if you are running in EKS/AKS/GKE - it is highly recommended you store any secrets in the Cloud provided secrets storage like AWS SecretsManager and *ensure your deployment* of the controller has valid pod identity to be able to perform the relevant retrieve operation.

e.g. in AWS `secretsmanager:GetSecret`

## Unit testing

To generate a JUNIT style report we are using this [package](https://github.com/jstemmer/go-junit-report)

sample usage see >> Makefile

`go test ``go list ./... | grep -v */generated/`` -coverprofile .coverage`
