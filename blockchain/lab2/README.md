# LAB-2: Add transaction to retrieve packages by name of contained product

This lab describes how you can add a new blockchain transaction by editing the [contract](./contract.json), and then build and deploy the updated chaincode to the Fabric test-network, and also build an updated HTTP client service that you can submit blockchain requests to.

This new transaction internally uses a Fabric chaincode API [GetQueryResult](https://github.com/hyperledger/fabric-chaincode-go/blob/master/shim/stub.go) to fetch blockchain states returned by a couchDB query. A [CouchDB query](https://docs.couchdb.org/en/latest/api/database/find.html) is a query statement in JSON format that can filter the resultset by the values of nested elements in a JSON document. Dovetail has made it easy to retrieve blockchain states by a CouchDB query statement, or the so called `rich query`.

## Add new transaction to the smart contract

You can use a JSON file editor, e.g., [vsCode](https://code.visualstudio.com/download) to edit the [contract](./contract.json) as follows, or if you want to quickly see the result, you can copy the [solution](./solution/contract.json) over the `contract.json` in this folder and skip to the next section for build and test.

Add the following transaction definition to [contract.json](./contract.json) under the section of `transactions`, e.g., after line 493.

```json
        {
          "name": "getPackagesByProduct",
          "parameters": [{
            "name": "product",
            "schema": {
              "type": "string"
            }
          }],
          "returns": {
            "type": "array",
            "items": {
              "$ref": "#/components/schemas/packageKeyValue"
            }
          },
          "rules": [{
            "description": "query list of packages by specified product",
            "actions": [{
                "activity": "#get",
                "name": "get_1",
                "ledger": {
                  "$ref": "#/components/schemas/package"
                },
                "config": {
                  "query": {
                    "selector": {
                      "content.product": "$product"
                    }
                  }
                },
                "input": {
                  "mapping": {
                    "data": {
                      "product": "=$flow.parameters.product"
                    }
                  }
                }
              },
              {
                "activity": "#actreturn",
                "input": {
                  "mapping": {
                    "status": "=$activity[get_1].code",
                    "message": "=$activity[get_1].message",
                    "returns": "=$activity[get_1].result"
                  }
                }
              }
            ]
          }]
        },
```

This new transaction is named `getPackagesByProduct`, which accepts a parameter `product` and returns an array of `packageKeyValue` that has already been defined under the section `components/schemas`. In the rule actions, a `#get` activity is used to query the blockchain state of `package` ledger by using a rich-query statement that matches the package attribute `content.product` to a specified product name. The activity input data is mapped to the value of the parameter `product`. The result of the query activity is then mapped to the returned data of the transaction.

To make the query run efficiently, we need to define CouchDB indices as shown under the folder [META-INF](./META-INF), which we have defined [indexProduct.json](./META-INF/statedb/couchdb/indexes/indexProduct.json) for the new transaction.

## Build and test the new transaction

Build chaincode and deploy it to Fabric test-network.

```bash
make
make start
make cc-init
```

Note that the [Makefile](./Makefile) starts the Fabric test-network with an option `-s couchdb`, which is necessary to use rich queries, i.e.,

```bash
cd $(FAB_PATH)/test-network && ./network.sh up createChannel -ca -s couchdb
```

The new transaction can be invoked directly by using the `cli` docker container. The test commands are included in [cc-test.sh](./cc-test.sh), which contains a test message at the end of the file that invokes the new transaction `getPackagesByProduct`, i.e.,

```bash
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"getPackagesByProduct","Args":["PfizerVaccine"]}'
```

Execute the direct tests using the following command

```bash
make cc-test
```

## Build HTTP client service and test remote requests

Create and build HTTP client service.

```bash
make build-client
make run
```

The HTTP requests for invoking blockchain transactions are listed in the [Makefile](./Makefile) under the `test` task. Note that the last test message is for the new transaction `getPackageByProduct`, i.e.,

```bash
curl -u User1: -X POST -H 'Content-Type: application/json' -d '{"product":"PfizerVaccine"}' http://localhost:$(PORT)/shipping/getpackagesbyproduct
```

Execute the tests for the HTTP client service using the following command

```bash
make test
```

## Shutdown

`Ctrl+C` to stop the HTTP client service. Then shutdown the Fabric test-network.

```bash
make shutdown
```
