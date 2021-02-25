#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# set Azure environment for a specified $ENV_NAME and $AZ_REGION
# usage: source env.sh env region
# default value: ENV_NAME="dtwin", AZ_REGION="westus2"

##### usually you do not need to modify parameters below this line

# return the full path of this script
function getScriptDir {
  local src="${BASH_SOURCE[0]}"
  while [ -h "$src" ]; do
    local dir ="$( cd -P "$( dirname "$src" )" && pwd )"
    src="$( readlink "$src" )"
    [[ $src != /* ]] && src="$dir/$src"
  done
  cd -P "$( dirname "$src" )" 
  pwd
}

export ENV_NAME=${1}
export AZ_REGION=${2}

export RESOURCE_GROUP=${ENV_NAME}RG
export BASTION_HOST=${ENV_NAME}Bastion
# public IP will be updated when bastion host is created
export BASTION_IP=40.65.112.23
export BASTION_USER=${ENV_NAME}

export SCRIPT_HOME=$(getScriptDir)
