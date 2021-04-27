# Blockchain Component

This folder contains the blockchain smart contract for the `Global logistics Services Demo`. We build and deploy the contract on the Hyperledger Fabric test-network, and generate a rest service for other demo components to invoke the contract transactions, all without any code.

If you have not setup the local development environment, you can follow the instructions in [README.md](docker/README.md) to quickly build and test the blockchain transactions.

The following instructions assumes that the local development environment has already been configured according to the instructions in [README.md](https://github.com/open-dovetail/fabric-chaincode/blob/master/README.md).

## Build and deploy contract

The shipping contract is defined in [contract.json](./contract.json), which can be built and deployed using the following scripts.

```bash
cd /path/to/demo/blockchain
make build
make deploy
make start
make cc-init
```

Refer to the [Makefile](./Makefile) for details of these commands, which are explained as follows:

The step `make build` uses the following scripts to build a package `shipping_cc_1.0.tar.gz` that can be deployed in a Hyperledger Fabric network.

```bash
# generate flogo model shipping.json from contract.json
flogo contract2flow -e -c contract.json -o shipping.json
# build model shipping.json and package it for deployment to Hyperledger Fabric
/path/to/fabric-chaincode/scripts/build.sh shipping.json shipping_cc
```

The generated Flogo model `shipping.json` can be edited visually by using a Flogo UI with more detailed data mapping if it is not fully specified in the `contract.json`. The modified model can then be exported and compiled using the above command.

The step `make deploy` copies the contract package `shipping_cc_1.0.tar.gz` and test scripts to the Hyperledger Fabric test-network.

The step `make start` executes Hyperledger Fabric commands to start a local test-network.

The step `make cc-init` executes the Hyperledger Fabric scripts of [cc-init.sh](./cc-init.sh) to install, approve, and commit the contract package `shipping_cc_1.0.tar.gz` on the test-network.

## On-network smoke test

The following command invokes blockchain transactions by executing the script [cc-test.sh](./cc-test.sh) in a `cli` docker container that is directly connected to the test-network.

```bash
make cc-test
```

## Build and run REST client service

From the [contract.json](./contract.json), we can generate a REST client service as a gateway for other application components to interact with the blockchain contract.

```bash
cd /path/to/demo/blockchain
make build-client
make run
```

The step `make build-client` uses the following scripts to build an executable `shipping_rest_app` that supports REST APIs for invoking blockchain transactions defined in [contract.json](./contract.json).

```bash
# generate flogo model shipping_rest.json from contract.json
flogo contract2rest -e -c contract.json -o shipping_rest.json
# build model shipping_rest.json into an executable for the local platform
/path/to/fabric-client/scripts/build.sh shipping_rest.json config.yaml local_entity_matchers.yaml
```

The generated Flogo model `shipping_rest.json` can be edited visually by using a Flogo UI to add or update REST APIs that differ from the generated interface. The modified model can then be exported and compiled using the above command. For example, this demo service uses a modified model [shipping_rest_fe.json](./shipping_rest_fe.json) that supports two additional APIs for signature verifications. Use the command `make client-rest` to build the modified REST service.

The REST service must be built for a specified Hyperledger Fabric network, which are defined by `config.yaml` and optionally `entity_matchers.yaml` for local networks. We use the network configuration for a local test-network, which can be found in <https://github.com/open-dovetail/fabric-client/tree/master/test-network>. To build an executable for a different platform, you can specify the platform in environment variables for the build step, e.g., `GOOS=darwin GOARCH=amd64 ./build.sh shipping_rest.json ...`, which would build an executable for Mac.

The step `make run` would start the REST service with important environment variables that specify listen port, chaincode, user crypto and authorization, and HTTP CORS support etc.

## Test REST service

The [Makefile](./Makefile) contains HTTP requests for invoking contract transactions via the REST service. Thus, you can run end-to-end blockchain tests by using the following command:

```bash
make test
```

## Shutdown

To cleanup all the demo processes, you can find and kill the REST service app, and execute `make shutdown` to shutdown the test-network of Hyperledger Fabric.
