#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# setup private key in test-network

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"; echo "$(pwd)")"
FAB_PATH=${1:-"${SCRIPT_DIR}/../../../../hyperledger/fabric-samples"}

KEYSTORE="${FAB_PATH}/test-network/organizations/peerOrganizations/org1.example.com/users/Admin@org1.example.com/msp/keystore"
CONFIG="${SCRIPT_DIR}/connection-profile/test-network.json"
if [ -d "${KEYSTORE}" ]; then
  for f in ${KEYSTORE}/*_sk; do
    echo $(basename $f)
    sed -i -e "s|/msp/keystore/.*|/msp/keystore/$(basename $f)\"|" ${CONFIG}
    break
  done
fi