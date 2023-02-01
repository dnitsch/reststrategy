# Demo & use cases

## Test pre-requisites

Ensure you have the required containers running

### mock app

Using [pocketbase.io](https://pocketbase.io/) to mock a real world app or SaaS system that will need configuring in a certain way to fit within your existing systems/flows.

`docker run --name=pb-app --detach -p 8090:8090 dnitsch/reststrategy-sample:latest`

Then navigate to this [page](http://127.0.0.1:8090/_/?installer#)

enter any username/password combo this is only for local testing
test@example.com/P4s$w0rd123!

<!-- `minikube service pocketbase -n testapp --url` -->
<!-- if you are running in Minikube and testing from the outside -->

### mock oauth server

Containerised version of [this](https://github.com/axa-group/oauth2-mock-server)

`docker run --name=oauth-mock --detach -p 8080:8080 dnitsch/reststrategy-oauth-mock:latest`

See `test/*-test.yaml` for an integration style test locally.

## What is it? Why use it?

    - one off (Commercetools, Ping, ForgeRock)
    - runtime configuration e.g. an onboarding activity

`docker stop pb-app ; docker rm pb-app ; docker run --name=pb-app --detach -p 8090:8090 dnitsch/reststrategy-sample:latest`

### Run in CLI mode (from CI or a one off set up)

`AWS_PROFILE=configmanager_demo AWS_REGION=eu-west-1 seeder/dist/seeder-darwin run -p test/pocketbase-cli-get-started.yaml`

### K8s deploy CRD and controller

Deploy customCRD

`kubectl apply -f crd/deployment.yml`
`kubectl apply -f crd/reststrategy.yml`

testapp:
`kubectl -- apply -f hack/testapp.yml`
`minikube service pocketbase -n testapp --url`

Run operator logic

`kubectl apply -f test/crd-test-ns.yml`
`kubectl apply -f test/crd-k8s-test.yml`
`kubectl apply -f test/crd-k8s-test-2.yml`

`kubectl delete -f test/crd-k8s-test.yml`
`kubectl delete -f test/crd-k8s-test-2.yml`

## Topics to explore

- [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- re-useable pattern for controllers
- informerCacheResync vs. PeriodicResync if no change occured.
- configmanager built in
    - 