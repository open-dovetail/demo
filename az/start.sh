#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# build and start demo and related services

SCRIPT_DIR="$( dirname "${BASH_SOURCE[0]}" )"
DEMO_HOME="$( dirname "${SCRIPT_DIR}" )"

# start TGDB
export TGDB_HOME=${HOME}/tgdb/3.0
cd ${DEMO_HOME}/graphdb
./start.sh

# build and start Hyperledger Fabric test-network
cd ${DEMO_HOME}/blockchain
make
make start
make cc-init

# build and run blockchain client service
make build-rest
make run

# build and run simulator
cd ${DEMO_HOME}/simulator
go build
./simulator &
