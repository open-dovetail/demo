#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# create VM as a bastion host for a specified $ENV_NAME and $AZ_REGION
# usage: create-bastion.sh env region
# default value: ENV_NAME="dtwin", AZ_REGION="westus2"

cd "$(cd "$(dirname "${BASH_SOURCE[0]}")"; echo "$(pwd)")"

if [ -z "${1}" ]; then
  ENV_NAME="dtwin"
fi

if [ -z "${2}" ]; then
    AZ_REGION="westus2"
fi

echo "ENV_NAME: ${ENV_NAME}, AZ_REGION: ${AZ_REGION}"
source env.sh ${ENV_NAME} ${AZ_REGION}

starttime=$(date +%s)
echo "create bastion host may take a few mminutes ..."

# create resource group if it does not exist already
check=$(az group show -g ${RESOURCE_GROUP} --query "properties.provisioningState" -o tsv)
if [ "${check}" == "Succeeded" ]; then
  echo "resource group ${RESOURCE_GROUP} is already provisioned"
else
  echo "create resource group ${RESOURCE_GROUP} at ${AZ_REGION} ..."
  az group create -l ${AZ_REGION} -n ${RESOURCE_GROUP}
fi

# create bastion host if it does not exist already
check=$(az vm show -n ${BASTION_HOST} -g ${RESOURCE_GROUP} --query "provisioningState" -o tsv)
if [ "${check}" == "Succeeded" ]; then
  echo "bastion host ${BASTION_HOST} is already provisioned"
else
  echo "create bastion host ${BASTION_HOST} with admin-user ${BASTION_USER} ..."
  az vm create -n ${BASTION_HOST} -g ${RESOURCE_GROUP} --image UbuntuLTS --generate-ssh-keys --admin-username ${BASTION_USER}
  # Note: docker extension v1.2.2 does not work with Hyperledger Fabric v2.2.1, so use the older version - 1.1.1606092330
  az vm extension set -n DockerExtension --publisher Microsoft.Azure.Extensions --version 1.2.0  --vm-name ${BASTION_HOST} -g ${RESOURCE_GROUP}
fi

# update security rule for ssh from localhost
myip=$(curl ifconfig.me)
echo "set security rule to allow ssh from host ${myip}"
az network nsg rule update -g ${RESOURCE_GROUP} --nsg-name ${BASTION_HOST}NSG --name default-allow-ssh --source-address-prefixes ${myip}
az network nsg rule create -g ${RESOURCE_GROUP} --nsg-name ${BASTION_HOST}NSG --name allow_dtwin_svc  --priority 4096 --source-address-prefixes ${myip} --destination-port-ranges 7979 7980 --access Allow --protocol Tcp

echo "collect public IP of bastion host ${BASTION_HOST} ..."
pubip=$(az vm list-ip-addresses -n ${BASTION_HOST} -g ${RESOURCE_GROUP} --query "[0].virtualMachine.network.publicIpAddresses[0].ipAddress" -o tsv)
sed -i -e "s/^export BASTION_IP=.*/export BASTION_IP=${pubip}/" ./env.sh

# setup bastion host
echo "copy config files to bastion host ${BASTION_USER}@${pubip}"
scp -q -o "StrictHostKeyChecking no" ./setup-bastion.sh ${BASTION_USER}@${pubip}:setup.sh
if [ -f "../graphdb/TIB_tgdb_3.0.0_linux_x86_64.zip" ]; then
  scp -q -o "StrictHostKeyChecking no" ../graphdb/TIB_tgdb_3.0.0_linux_x86_64.zip ${BASTION_USER}@${pubip}:
fi

if [ "${check}" == "Succeeded" ]; then
  echo "skip setup for existing bastion host"
else
  echo "execute setup.sh on bastion host ${BASTION_USER}@${pubip}"
ssh -o "StrictHostKeyChecking no" ${BASTION_USER}@${pubip} << EOF
  ./setup.sh
EOF
fi

echo "Bastion host ${BASTION_HOST} created in $(($(date +%s)-starttime)) seconds."
echo "Access the bastion host by ssh:"
echo "  ssh ${BASTION_USER}@${pubip}"
