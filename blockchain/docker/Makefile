#
# Copyright (c) 2020, TIBCO Software Inc.
# All rights reserved.
#
# SPDX-License-Identifier: BSD-3-Clause-Open-MPI
#
MAKEFILE_THIS := $(lastword $(MAKEFILE_LIST))
SRC_PATH      := $(patsubst %/,%,$(dir $(abspath $(MAKEFILE_THIS))))
DOVETAIL_PATH := $(SRC_PATH)/../../..
FAB_PATH      ?= $(DOVETAIL_PATH)/hyperledger/fabric-samples
CC_DEPLOY     := $(FAB_PATH)/chaincode
PORT          := 7979

# uncomment this line to build for Flogo Enterprise
#FE            := fe

.PHONY: all
all: start-dovetail build deploy

.PHONY: start-dovetail
start-dovetail:
	docker-compose -f docker-dovetail.yaml up -d

.PHONY: build
build:
	docker exec dovetail bash -c './contract-to-flow contract/shipping-contract.json shipping.json $(FE)'
	docker exec dovetail bash -c './build-cc contract/shipping.json'
	docker exec dovetail bash -c './contract-to-rest contract/shipping-contract.json shipping_rest.json $(FE)'
	docker exec dovetail bash -c 'GOOS=darwin GOARCH=amd64 ./build-app contract/shipping_rest.json'

.PHONY: deploy
deploy:
	cp contract/shipping_cc_1.0.tar.gz $(CC_DEPLOY)
	cp ../cc-init.sh $(FAB_PATH)/test-network/cc-init-shipping.sh
	cp ../cc-test.sh $(FAB_PATH)/test-network/cc-test-shipping.sh

.PHONY: start
start:
	cd $(FAB_PATH)/test-network && ./network.sh up createChannel -ca
	../gen-certs.sh

.PHONY: cc-init
cc-init: $(FAB_PATH)/test-network/cc-init-shipping.sh
	docker exec cli bash -c './cc-init-shipping.sh'

.PHONY: cc-test
cc-test: $(FAB_PATH)/test-network/cc-test-shipping.sh
	docker exec cli bash -c './cc-test-shipping.sh'

.PHONY: shutdown
shutdown:
	cd $(FAB_PATH)/test-network && ./network.sh down

.PHONY: build-rest
build-rest:
	cp ../shipping_rest_fe.json contract/shipping_rest.json
	docker exec dovetail bash -c 'GOOS=darwin GOARCH=amd64 ./build-app contract/shipping_rest.json'

.PHONY: run
run:
	# Flogo rest trigger config for CORS headers are required to specify auth header.  It can be tested using, e.g.,
	# curl -i -X OPTIONS -H "Access-Control-Request-Method: POST" -H "Access-Control-Request-Headers: Content-Type,Accept,Authorization" -H "Origin: http://localhost:7979" -H "Authorization: Basic VXNlcjE6" -H "Content-Type: application/json" -d '{"uid":"4dd6d722229ae46f","transactionType":"transfer"}' http://52.229.51.17:7979/shipping/packagetimeline
	FLOGO_APP_PROP_RESOLVERS=env FLOGO_APP_PROPS_ENV=auto PORT=$(PORT) CHAINCODE=shipping_cc FLOGO_LOG_LEVEL=DEBUG FLOGO_SCHEMA_SUPPORT=true FLOGO_SCHEMA_VALIDATION=false CRYPTO_PATH=$(FAB_PATH)/test-network/organizations REST_TRIGGERCORS_ALLOW_CREDENTIALS=true REST_TRIGGERCORS_ALLOW_HEADERS="Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, X-Requested-With, Accept, Accept-Language, Authorization" nohup contract/shipping_rest_app 2>&1 &

.PHONY: test
test:
	echo "prepare package transactions ..."
	curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","eventTime":"2021-02-12 09:57:25","latitude":40.8077,"longitude":-74.0692,"packageDetail":"{ \"uid\": \"677e9dcf985c8902\", \"from\": { \"street\": \"E 16th St.\", \"postal-code\": \"11212\" }, \"to\": { \"street\": \"E Florence Ave\", \"postal-code\": \"90001\" }, \"content\": { \"product\": \"PfizerVaccine\", \"description\": \"COVID-19 vaccine\", \"producer\": \"Pfizer\", \"count\": 100, \"start-lot-number\": \"A00001Y\", \"end-lot-number\": \"A00100Y\"}}"}' http://localhost:$(PORT)/shipping/pickuppackage
	curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","eventTime":"2021-02-12 20:30:25","toCarrier":"SLS","latitude":39.7392,"longitude":-104.9903}' http://localhost:$(PORT)/shipping/transferpackage
	curl -u slsadm@org2: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","eventTime":"2021-02-12 20:30:40","carrier":"NLS","latitude":39.7392,"longitude":-104.9903}' http://localhost:$(PORT)/shipping/transferpackageack
	curl -u slsadm@org2: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","eventTime":"2021-02-13 10:30:40","latitude":33.9416,"longitude":-118.4085}' http://localhost:$(PORT)/shipping/deliverpackage
	sleep 3
	echo "query package transactions ..."
	curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902"}' http://localhost:$(PORT)/shipping/getpackagebyuid
	curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902"}' http://localhost:$(PORT)/shipping/packagetimeline
	echo "prepare temperature updates ..."
	curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","periodStart":"2021-02-12 09:57:25","containerID":"e6b1c21e124125cb","eventTime":"2021-02-12 15:57:25","minValue":-75,"maxValue":-65,"inViolation":false}' http://localhost:$(PORT)/shipping/updatetemperature
	curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","periodStart":"2021-02-12 21:57:25","containerID":"e6b1c21e124125cb","eventTime":"2021-02-12 21:59:25","minValue":-50,"maxValue":-45,"inViolation":true}' http://localhost:$(PORT)/shipping/updatetemperature
	curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","periodStart":"2021-02-12 22:07:25","containerID":"e6b1c21e124125cb","eventTime":"2021-02-13 09:59:25","minValue":-70,"maxValue":-60,"inViolation":false}' http://localhost:$(PORT)/shipping/updatetemperature
	sleep 3
	echo "query package environment ..."
	curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902"}' http://localhost:$(PORT)/shipping/packageenvironment
	# use the following to test modified service API for user CA - shipping_rest_fe.json
	# curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","transactionType":"transfer"}' http://localhost:$(PORT)/shipping/verifytransaction
	# curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"677e9dcf985c8902","periodStart":"2021-02-12 21:57:25"}' http://localhost:$(PORT)/shipping/verifytemperature
