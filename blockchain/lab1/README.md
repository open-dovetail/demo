# LAB-1: Add transaction to retrieve packages by a specified sender

This lab describes how you can add a new blockchain transaction by editing the [contract](./contract.json), and then build and deploy the updated chaincode to the Fabric test-network, and also build an updated HTTP client service that you can submit blockchain requests to.

This new transaction internally uses a Fabric chaincode API [GetStateByPartialCompositeKey](https://github.com/hyperledger/fabric-chaincode-go/blob/master/shim/stub.go) to fetch blockchain states that matches a composite key. A `composite key` is similar to an index on a database table. Dovetail has made it easy to define and create the index, and use the index to retrieve matching blockchain states.

## Add new transaction to the smart contract

You can use a JSON file editor, e.g., [vsCode](https://code.visualstudio.com/download) to edit the [contract](./contract.json) as follows, or if you want to quickly see the result, you can copy the [solution](./solution/contract.json) over the `contract.json` in this folder and skip to the next section for build and test.

Add the following transaction definition to [contract.json](./contract.json) under the section of `transactions`, e.g., after line 434.

```json
        {
          "name": "getPackagesBySender",
          "parameters": [{
            "name": "sender",
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
            "description": "retrieve list of packages by specified sender",
            "actions": [{
                "activity": "#get",
                "name": "get_1",
                "ledger": {
                  "$ref": "#/components/schemas/package"
                },
                "config": {
                  "compositeKeys": {
                    "sender~uid": [
                      "sender",
                      "uid"
                    ]
                  }
                },
                "input": {
                  "mapping": {
                    "data": {
                      "sender": "=$flow.parameters.sender"
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

This new transaction is named `getPackagesBySender`, which accepts a parameter `sender` and returns an array of `packageKeyValue` that has already been defined under the section `components/schemas`. In the rule actions, a `#get` activity is used to query the blockchain state of `package` ledger by using a composite key `sender~uid` that is composed of 2 package attributes, `sender` and `uid`. The activity input data is mapped to the value of the parameter `sender`. The result of the query activity is then mapped to the returned data of the transaction.

For this query to work, we also need to edit the package creation transaction to insert a `sender~uid` key for all new packages. In this smart contract, a new package is created by activity `put_2` in the `pickupPackage` transaction, i.e., line 98. Edit this activity by adding the following lines after line 103.

```json
                "config": {
                  "compositeKeys": {
                    "sender~uid": [
                      "sender",
                      "uid"
                    ]
                  }
                },
```

This will make the activity to create a `sender~uid` key for new `package`s on the ledger.

## Build and test the new transaction

Build chaincode and deploy it to Fabric test-network.

```bash
make
make start
make cc-init
```

The new transaction can be invoked directly by using the `cli` docker container. The test commands are included in [cc-test.sh](./cc-test.sh), which contains a test message at the end of the file that invokes the new transaction `getPackagesBySender`, i.e.,

```bash
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"getPackagesBySender","Args":["John"]}'
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

The HTTP requests for invoking blockchain transactions are listed in the [Makefile](./Makefile) under the `test` task. Note that the last test message is for the new transaction `getPackageBySender`, i.e.,

```bash
curl -u User1: -X POST -H 'Content-Type: application/json' -d '{"sender":"John"}' http://localhost:$(PORT)/shipping/getpackagesbysender
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
