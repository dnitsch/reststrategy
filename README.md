# strategy-rest-seeder 

> WIP

Executes a series of instructions against an endpoint using a specified authentication on a given strategy.


## Problem space

Often we are faced with external products/services either self-hosted or SaaS where they require some level of configuration this often happens at CI (deploy) time, or on continous basis 

## Strategy

Strategy is a setting against which to perform one or more rest calls to ensure an idempotent update.

### `GET/POST`

FindPostStrategyFunc strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST*** this strategy should be used sparingly and only in cases where the service REST implementation does not support an update of existing item.

### `FIND/POST`

FindPostStrategyFunc strategy calls a GET endpoint and if item ***FOUND it does NOT do a POST*** this strategy should be used sparingly and only in cases where the service REST implementation does not support an update of existing item.

### `PUT/POST`

PutPostStrategyFunc is useful when the resource is created a user specified Id
the PUT endpoint DOES NOT support a creation of the resource. PUT should throw a 4XX
for the POST fallback to take effect

### `GET/PUT/POST`

GetPutPostStrategyFunc known ID and only know a name or other indicator the pathExpression must not evaluate to an empty string in order to for the PUT to be called else POST will be called as item was not present.


### `FIND/PUT/POST`

FindPutPostStrategyFunc is useful when the resource Id is unknown i.e. handled by the system.
providing a pathExpression will evaluate the response.
the pathExpression must not evaluate to an empty string in order to for the PUT to be called
else POST will be called as item was not present

### `FIND/DELETE`

### `FIND/DELETE/POST`

### `PUT`



## AuthMap

Is a map of authentication objects 


## Rest Action

```yaml
endpoint: https://postman-echo.com
strategy: FIND/PUT/POST
getEndpointSuffix: /get?json=provided&valid=true
postEndpointSuffix: /post
putEndpointSuffix: /put
findByJsonPathExpr: "$.array[?(@.name=='fubar')].id"
authMapRef: ouath1
payloadTemplate: |
    {
    "value": "$foo"
    }
variables:
    foo: bar
runtimeVars:
    someId: "$.array[?(@.name=='fubar')].id"
```

some other test here 