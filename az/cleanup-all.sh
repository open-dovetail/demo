#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# cleanup demo and related services

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"; echo "$(pwd)")"
DEMO_HOME="$( dirname "${SCRIPT_DIR}" )"

# stop running services
procs=$(ps -ef | egrep "simulator|shipping_rest_app|tgdb" | egrep -v grep | awk '{print $2}')
for pid in $procs; do
  echo $pid
  kill -9 $pid
done

# cleanup simulator files
cd ${DEMO_HOME}/simulator
rm nohup.out
rm -R log

# stop Hyperledger Fabric test-network
cd ${DEMO_HOME}/blockchain
make shutdown
rm shipping*.gz
rm shipping*_app
rm nohup.out
rm -R keystore
