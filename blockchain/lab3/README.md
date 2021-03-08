# LAB-3: Generate chaincode events

This lab describes how you can add chaincode events to blockchain transactions by editing the [contract](./contract.json), and then build and deploy the updated chaincode to the Fabric test-network.

The transaction `updateTemperature` will be edited such that an event will be emitted for IoT measurements that violated the temperature threshold of the contained product. An event listener can monitor the committed blocks and act on the emitted chaincode events. This new event activity internally uses a Fabric chaincode API [SetEvent](https://github.com/hyperledger/fabric-chaincode-go/blob/master/shim/stub.go) to define the event name and payload schema. Dovetail has made it easy to define and create chaincode events.

## Add new event activity to the smart contract

You can use a JSON file editor, e.g., [vsCode](https://code.visualstudio.com/download) to edit the [contract](./contract.json) as follows, or if you want to quickly see the result, you can copy the [solution](./solution/contract.json) over the `contract.json` in this folder and skip to the next section to build and test it.

Chaincode events are implemented by a Flogo activity [setevent](https://github.com/open-dovetail/fabric-chaincode/tree/master/activity/setevent), so first add the following line to the `imports` section of the [contract](./contract.json), e.g., on line 10.

```bash
    "github.com/open-dovetail/fabric-chaincode/activity/setevent",
```

Add the following event activity definition to the `updateTemperature` transaction in [contract.json](./contract.json), e.g., after line 644.

```json
              {
                "activity": "#setevent",
                "name": "setevent_1",
                "input": {
                  "sample": {
                    "payload": {
                      "key": "",
                      "violationType": "",
                      "periodStart": "",
                      "periodEnd": "",
                      "minValue": 0,
                      "maxValue": 0
                    }
                  },
                  "mapping": {
                    "name": "=string.concat($activity[put_1].result[0].value.measurementType, \" violation\")",
                    "payload": {
                      "key": "=$activity[put_1].result[0].key",
                      "violationType": "=$activity[put_1].result[0].value.measurementType",
                      "periodStart": "=$activity[put_1].result[0].value.periodStart",
                      "periodEnd": "=$activity[put_1].result[0].value.eventTime",
                      "minValue": "=$activity[put_1].result[0].value.minValue",
                      "maxValue": "=$activity[put_1].result[0].value.maxValue"
                    }
                  }
                }
              },
```

This new activity defines the schema of the event payload by using a sample JSON object, which specifies the data structure and element type of the payload, and it maps the event name and payload attributes to results from the previous activities. You may specify JSON schema of the payload directly, although a sample JSON object is usually easier for most developers.

## Build and test the new transaction

Build chaincode and deploy it to Fabric test-network.

```bash
make
make start
make cc-init
```

Execute tests in `cli` docker container using the following command

```bash
make cc-test
```

## Browse blockchain by using blockchain-explorer

You can use the Hyperledger [blockchain explorer](https://github.com/hyperledger/blockchain-explorer) to view the blockchain transactions. Use the following command to start the explorer service.

```bash
make start-explorer
```

Browse the Fabrtic test-network at <http://localhost:8080>, login as the default user `exploreradmin` and password `exploreradminpw`.

However, the explorer does not show the chaincode events that this lab has created. You will need an event listner to capture and verify the events. A Flogo trigger [eventlistener](https://github.com/dovetail-lab/fabric-client/tree/master/trigger/eventlistener) exists, but it has not been merged into the `open-dovetail` repository, and thus it does not support Open-source Flogo UI yet.

## Build HTTP client service and test remote requests

Create and build HTTP client service.

```bash
make build-client
make run
```

Execute the tests on the HTTP client service using the following command

```bash
make test
```

## Shutdown

`Ctrl+C` to stop the HTTP client service. Then shutdown the Fabric test-network.

```bash
make shutdown
```
