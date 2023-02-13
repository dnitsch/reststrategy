# Rest Strategy

[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=dnitsch_reststrategy&metric=bugs)](https://sonarcloud.io/summary/new_code?id=dnitsch_reststrategy)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=dnitsch_reststrategy&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=dnitsch_reststrategy)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=dnitsch_reststrategy&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=dnitsch_reststrategy)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=dnitsch_reststrategy&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=dnitsch_reststrategy)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=dnitsch_reststrategy&metric=coverage)](https://sonarcloud.io/summary/new_code?id=dnitsch_reststrategy)

__Seeder__: [![Go Report Card](https://goreportcard.com/badge/github.com/dnitsch/reststrategy/seeder)](https://goreportcard.com/report/github.com/dnitsch/reststrategy/seeder)

__Controller__: [![Go Report Card](https://goreportcard.com/badge/github.com/dnitsch/reststrategy/controller)](https://goreportcard.com/report/github.com/dnitsch/reststrategy/controller)

Rest Strategy is a collection of packages to enable idempotent seeding of data against REST endpoints.

This repo uses workspaces and is made up of following modules/components.

- [seeder](./seeder/README.md) modules which can be used as a [library](https://pkg.go.dev/github.com/dnitsch/reststrategy/seeder) (used by the controller) or a CLI.

- [controller](./controller/README.md) k8s controller code.

- [apis](./apis/README.md) holds the types for the controller.

See the individual components for a lower level overview.

When interacting with the modules use the top level Makefile tasks.

After any change or before any push to remote

- `make build`
- `make test`

To run tests against the controller or the CLI use the `test/crd-test.yml`.

## Examples

To see some [examples](./docs/example.md)...

## Notes

General notes section

### TEMPLATING

> _INFO_

    When assigning ConfigManager tokens to the templates - it is highly recommended to use the variables section and reference that instead of inlining the tokens in payloads themselves.

e.g. prefer this - see the full example [here](test/crd-k8s-test-2.yml)

```yaml
...
payloadTemplate: |
    {"email":"qa-guy@example.com","password":"${oldPass}","passwordConfirm":"${oldPass}"}
patchPayloadTemplate: |
    {"password":"${newPass}","passwordConfirm":"${newPass}"}
variables:
    oldPass: AWSPARAMSTR:///int-test/pocketbase/admin-pwd
    newPass: AWSPARAMSTR:///int-test/pocketbase/admin-pwd
```

as opposed to the below

```yaml
...
payloadTemplate: |
    {"email":"qa-guy@example.com","password":"Password123_alwaysChange","passwordConfirm":"Password123_alwaysChange"}
patchPayloadTemplate: |
    {"password":"AWSPARAMSTR:///int-test/pocketbase/admin-pwd","passwordConfirm":"AWSPARAMSTR:///int-test/pocketbase/admin-pwd"}
variables: {}
```

IF the exchanged config value for a token includes an unescaped `$` it will try to be templated and may result in unexpected errors.

The templating function is provided already with a replaced value: `P4s$w0rd123!` for token: `AWSPARAMSTR:///int-test/pocketbase/admin-pwd`. 

> `$w0rd123` portion of the replaced value will be picked up by the envsubst lexer as a variable token and will yield no results. 

### Experiment

This is a bit of experiment with controller structures and workspaces to see if some re-usable patterns can be gleamed and used in some code generation scaffolding.

At the very least a copy and paste into a new workspace of existing structure and Make tasks -> deleting following `seeder`, `controller/pkg/rstservice`, removing k8sutils should leave you with a fairly re-useable controller pkg which can be exchanged with other top level types coming from custom/new `apis` module.
