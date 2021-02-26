#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# cleanup bastion host and resource group for a specified $ENV_NAME and $AZ_REGION
# usage: cleanup-bastion.sh env region
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
echo "cleanup may take 2-3 mminutes ..."

echo "delete bastion host ${BASTION_HOST}"
az vm delete -n ${BASTION_HOST} -g ${RESOURCE_GROUP} -y

echo "delete resource group ${RESOURCE_GROUP}"
az group delete -n ${RESOURCE_GROUP} -y

echo "Cleaned up ${RESOURCE_GROUP} in $(($(date +%s)-starttime)) seconds."
