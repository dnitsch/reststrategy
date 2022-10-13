# Rest Strategy

Rest Strategy is a collection of packages to enable idempotent seeding of data against REST endpoints.

This module uses workspaces and is made up of following submodules/components.

- [seeder](./seeder/README.md) modules which can be used as a library (used by the controller) or a CLI.

- [controller](./controller/README.md) k8s controller code.

- [apis](./apis/README.md) holds the types for the controller.

See the individual componenets for a lower level overview.

When interacting with the modules use the top level Makefile tasks.

`make build` 

To run tests against the controller or the CLI use the `test/`

## Test pre-requisites

Ensure you have the required containers running

### mock app

Using [pocketbase.io](https://pocketbase.io/) to mock a real world app or SaaS system that will need configuring in a certain way to fit within your existing systems/flows.

`docker run --name=pb-app --detach -p 8090:8090 dnitsch/reststrategy-sample:latest`

Then navigate to this [page](http://127.0.0.1:8090/_/?installer#)

enter any username/password combo this is only for local testing
test@example.com/P4s$w0rd123!

`minikube service pocketbase -n testapp --url`

### mock oauth server

Containerised version of [this](https://github.com/axa-group/oauth2-mock-server)

`docker run --name=oauth-mock --detach -p 8080:8080 dnitsch/reststrategy-oauth-mock:latest`

if you are running in Minikube and testing from the outside

See test/*-test.yaml for an integration style test locally.

## Notes

This is a bit of experiment with controller structures and workspaces to see if some re-usable patterns can be gleamed and used in some code generation scaffolding.

At the very least a copy and paste into a new workspace of existing structure and Make tasks -> deleting following `seeder`, `controller/pkg/rstservice`, removing k8sutils should leave you with a fairly re-useable controller pkg which can be exchanged with other top level types coming from custom/new `apis` module.