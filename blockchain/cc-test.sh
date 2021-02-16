#!/bin/bash
#
# Copyright (c) 2020, TIBCO Software Inc.
# All rights reserved.
#
# SPDX-License-Identifier: BSD-3-Clause-Open-MPI
#
# shipping_cc tests executed from cli docker container of the Fabric test-network

. ./scripts/envVar.sh
CCNAME=shipping_cc

setGlobals 1
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/test-network/organizations/peerOrganizations/org1.example.com/users/nlsadm@org1.example.com/msp

ORDERER_ARGS="-o orderer.example.com:7050 --ordererTLSHostnameOverride orderer.example.com --tls --cafile $ORDERER_CA"
ORG1_ARGS="--peerAddresses peer0.org1.example.com:7051 --tlsRootCertFiles $PEER0_ORG1_CA"
ORG2_ARGS="--peerAddresses peer0.org2.example.com:9051 --tlsRootCertFiles $PEER0_ORG2_CA"

# pickup package
echo "pickup package by user 'nlsadm@org1' ..."
peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"pickupPackage","Args":["677e9dcf985c890d","2021-02-12 09:57:25","40.8077","-74.0692","{ \"uid\": \"677e9dcf985c890d\", \"from\": { \"street\": \"E 16th St.\", \"postal-code\": \"11212\" }, \"to\": { \"street\": \"E Florence Ave\", \"postal-code\": \"90001\" }, \"content\": { \"product\": \"PfizerVaccine\", \"description\": \"COVID-19 vaccine\", \"producer\": \"Pfizer\", \"count\": 100, \"start-lot-number\": \"A00001X\", \"end-lot-number\": \"A00100X\"}}"]}'

# transfer package
echo "transfer package by user 'nlsadm@org1' ..."
peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"transferPackage","Args":["677e9dcf985c890d","2021-02-12 20:30:25","SLS","39.7392","-104.9903"]}'

# transfer package ack
echo "transfer package ack by user 'slsadm@org2' ..."
setGlobals 2
export CORE_PEER_ADDRESS=peer0.org2.example.com:9051
export CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/test-network/organizations/peerOrganizations/org2.example.com/users/slsadm@org2.example.com/msp

peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"transferPackageAck","Args":["677e9dcf985c890d","2021-02-12 20:30:40","NLS","39.7392","-104.9903"]}'
#peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"deliverPackage","Args":["677e9dcf985c890d","2021-02-13 10:30:40","33.9416","-118.4085"]}'

# get package by uid
sleep 5
echo "get package by uid ..."
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"getPackageByUID","Args":["677e9dcf985c890d"]}'

# get package transaction
echo "get package transaction for 'transferAck'..."
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"getPackageTransaction","Args":["677e9dcf985c890d","transferAck"]}'

# get package timeline
echo "get package timeline ..."
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"packageTimeline","Args":["677e9dcf985c890d"]}'

# update package temperature
setGlobals 1
export CORE_PEER_ADDRESS=peer0.org1.example.com:7051
export CORE_PEER_MSPCONFIGPATH=/etc/hyperledger/test-network/organizations/peerOrganizations/org1.example.com/users/iot@org1.example.com/msp

echo "update temperature ..."
peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"updateTemperature","Args":["677e9dcf985c890d","2021-02-12 09:57:25","e6b1c21e124125cb","2021-02-12 15:57:25","-75","-65","false"]}'
peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"updateTemperature","Args":["677e9dcf985c890d","2021-02-12 21:57:25","e6b1c21e124125cb","2021-02-12 21:59:25","-50","-45","true"]}'
peer chaincode invoke $ORDERER_ARGS -C mychannel -n $CCNAME $ORG1_ARGS $ORG2_ARGS -c '{"function":"updateTemperature","Args":["677e9dcf985c890d","2021-02-12 22:07:25","e6b1c21e124125cb","2021-02-13 09:59:25","-70","-60","false"]}'

# get package environment
sleep 5
echo "get package environment ..."
peer chaincode query -C mychannel -n $CCNAME -c '{"function":"packageEnvironment","Args":["677e9dcf985c890d"]}'
