# Demo & use cases

What is restseeder? Why use it?
    - one off (Commercetools, Ping, ForgeRock)
    - runtime configuration e.g. an onboarding activity
`docker stop pb-app && docker rm pb-app && docker run --name=pb-app --detach -p 8090:8090 dnitsch/reststrategy-sample:latest`

Run in CLI mode (from CI or a one off set up)

`AWS_PROFILE=configmanager_demo AWS_REGION=eu-west-1 seeder/dist/seeder-darwin run -p test/pocketbase-cli-get-started.yaml`

K8s deploy CRD and controller

Deploy customCRD

`minikube kubectl -- apply -f crd/deployment.yml`
`minikube kubectl -- apply -f crd/reststrategy.yml`

testapp:
`minikube kubectl -- apply -f hack/testapp.yml`
`minikube service pocketbase -n testapp --url`

Run operator logic

`minikube kubectl -- apply -f test/crd-test-ns.yml`
`minikube kubectl -- apply -f test/crd-k8s-test.yml`

`minikube kubectl -- delete -f test/crd-k8s-test-2.yml`

## Topics to explore

- [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)
- re-useable pattern for controllers
- informerCacheResync vs. PeriodicResync if no change occured.
- configmanager built in
    - 