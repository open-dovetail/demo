# Build Blockchain Contract Using Prebuilt Docker Image

For developers who is interested in quick startup with Dovetail applications without setting up local development environment, the following steps illustrates the build and test process by using prebuilt Docker images.

## Install Docker and Hyperledger Fabric test-network

Follow the instructions of <https://docs.docker.com/get-docker/> to install Docker and docker compose.

Install the following version of Hyperledger Fabric test-network in an empty open-dovetail folder:

```bash
mkdir -p open-dovetail/hyperledger
cd open-dovetail/hyperledger
curl -sSL http://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9
```

This should install a test-network in the folder `open-dovetail/hyperledger/fabric-samples/test-network`, which we'll use to test the demo contract.

For the following steps to work without editing any script path, you should clone the `demo` project in the same `open-dovetail` folder, i.e.,

```bash
cd open-dovetail
git clone https://github.com/open-dovetail/demo.git
```

## Build contract chaincode and services

The demo contract [shipping-contract.json](./contract/shipping-contract.json) is the only artifact for this demo application, which is used to generate deployable chaincode package for Hyperledger Fabric, as well as a service application that provide REST APIs for receiving blockchain requests.

```bash
cd open-dovetail/demo/blockchain/docker
make start-dovetail
make build
```

Refer to the [Makefile](./Makefile) for details of these commands, which are explained as follows:

The step `make start-dovetail` starts a docker container to execute all Dovetail build commands, i.e.,

```bash
docker-compose -f docker-dovetail.yaml up -d
```

The step `make build` executes the following scripts to build a Hyperledger Fabric chaincode package `shipping_cc_1.0.tar.gz` and an executable `shipping_rest_app` for REST API server.

```bash
# generate Flogo model for Fabric chaincode - shipping.json
docker exec dovetail bash -c './contract-to-flow contract/shipping-contract.json shipping.json'
# compile chaincode model and assemble deployable package - shipping_cc_1.0.tar.gz
docker exec dovetail bash -c './build-cc contract/shipping.json'
# generate Flogo model for REST service app - shipping_rest.json
docker exec dovetail bash -c './contract-to-rest contract/shipping-contract.json shipping_rest.json'
# compile the REST app into executable for Mac - shipping_rest_app
docker exec dovetail bash -c 'GOOS=darwin GOARCH=amd64 ./build-app contract/shipping_rest.json'
```

The REST service must be built for a specified Hyperledger Fabric network, which are defined by `config.yaml` and optionally `entity_matchers.yaml` for local networks. When the configuration files are not specified, it is built using the local test-network config files in <https://github.com/open-dovetail/fabric-client/tree/master/test-network>. The above command builds an REST service executable for Mac. The environment variables `GOOS` and `GOARCH` must be changed accordingly for other platforms.

## View or edit Flogo model using Flogo-UI

The generated Flogo models `shipping.json` and `shipping_rest.json` can be edited in Flogo UI, before they are compiled and deployed.

Use the following command to start the Flogo UI service that is pre-configured with Dovetail extensions:

```bash
docker run -it -p 3303:3303 yxuco/flogo-ui eula-accept
```

Open the UI <http://localhost:3303> in a web browser, and import the models `shipping.json` and `shipping_rest.json` to view or edit. Export the models if they are edited.

## Deploy chaincode to test-network and run on-network smoke test

```bash
make deploy
make start
make cc-init
make cc-test
```

The step `make deploy` copies the contract package `shipping_cc_1.0.tar.gz` and test scripts to the Hyperledger Fabric test-network.

The step `make start` executes Hyperledger Fabric commands to start a local test-network.

The step `make cc-init` executes the Hyperledger Fabric scripts of [cc-init.sh](../cc-init.sh) to install, approve, and commit the contract package `shipping_cc_1.0.tar.gz` on the test-network.

The step `make cc-test` invokes blockchain transactions by executing smoke test script [cc-test.sh](../cc-test.sh) in a `cli` docker container that is directly connected to the test-network.

## Start and test REST client service

```bash
make run
make test
```

The step `make run` would start the REST service with important environment variables that specify listen port, chaincode, user crypto and authorization, and HTTP CORS support etc.

The step `make test` executes end-to-end tests by using the HTTP requests in the [Makefile](./Makefile).

## Shutdown

To cleanup all the demo processes, you can find and kill the REST service app, and execute `make shutdown` to shutdown the test-network of Hyperledger Fabric.
