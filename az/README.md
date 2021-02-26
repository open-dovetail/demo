# Setup Microsoft Azure

The scripts in this section will setup a `bastion` host that you can login and start a Hyperledger Fabric network. The configuration file [env.sh](./env.sh) specifies the default configuration.

## Configure Azure account login

Install [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest) as described by the link.

Once your Azure account is setup, you can login by typing the command:

```bash
az login
```

Enter your account info in a pop-up browser window. Note that you may lookup your account details by using the [Azure Portal](https://portal.azure.com), although it is not absolutely necessary since we use only `Azure CLI` scripts.

## Create bastion VM and login

Create the bastion VM with all defaults:

```bash
cd /path/to/open-dovetail/demo/az
./create-bastion.sh create
```

This script accepts 2 parameters for you to specify a different Azure environment, e.g.,

```bash
./create-bastion.sh  dtwin westus2
```

would create an AKS cluster with name prefix of `dtwin`, at the Azure location of `westus2`.

Wait 2 minutes for the bastion VM is up, it will print a line, such as:

```bash
ssh fab@40.65.112.23
```

You can use this command to login to the `bastion` VM instance and start the Hyperledger Fabric test-network. Note that the `ssh` keypair for accessing the `bastion` host is in your `$HOME/.ssh` folder, and named as `id_rsa.pub` and `id_rsa`. The script will generate a new keypair if these files do not exist already.

Note also that the scripts have set the security group such that the `bastion` host can be accessed by only your workstation's current IP address. If your IP address changes, you'll need to login to Azure to update the security rule, or simply re-run the script:

```bash
cd ./az
az login
./create-bastion.sh dtwin westus2
```

Note in the script `create-bastion.sh`, we loaded Microsoft docker extension `v1.2.0`. Do not use the `v1.2.2` because it does not work with Hyperledger Fabric v2.2.1 test-network. Besides, the `network.sh` script does not shutdown the test-network because this version of `docker-compose` does not support the option flag `--remove-orphans`. You can fix this issue by removing this flag in the script `network.sh`.

## Start simulator and related services

Log on to the `bastion` host, e.g., (your real host IP will be different):

```bash
ssh fab@40.65.112.23
```

Set `$TGDB_HOME` to where the TIBCO graph DB is installed.  Then, start all required services scripted in [start-all.sh](./start-all.sh)

```bash
cd /path/to/demo/az
export TGDB_HOME=$HOME/tgdb/3.0
./start-all.sh
```

The blockchain client `shipping_rest_app` service will listen on `http://40.65.112.23:7979`. The `simulator` service will listen on `http://40.65.112.23:7980`. To get access to these services, the demo presenter can provide the presenter's IP address, i.e., the output from `curl ifconfig.me`, and add it to the Azure security rule.

If the bastion VM is at IP address `40.65.112.23`, you can create a package using the sample data [package.json](../simulator/package.json):

```bash
cd /path/to/demo/simulator
curl -X PUT -H "Content-Type: application/json" -d @package.json http://40.65.112.23:7980/packages/create
```

If the returned package UID is `2f850cc1cd8e670a`, you can use the following APIs to process the package and fetch the results:

```bash
# invoke simulator APIs
curl -X PUT -H "Content-Type: application/json" http://40.65.112.23:7980/packages/pickup?uid=2f850cc1cd8e670a
curl -X GET -H "Content-Type: application/json" http://40.65.112.23:7980/packages/timeline?uid=2f850cc1cd8e670a
```

Verify Blockchain transactions using the following APIs

```bash
# query package transactions on blockchain
curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"2f850cc1cd8e670a"}' http://40.65.112.23:7979/shipping/packagetimeline

# query package environment on blockchain
curl -u iot: -X POST -H 'Content-Type: application/json' -d '{"uid":"2f850cc1cd8e670a"}' http://40.65.112.23:7979/shipping/packageenvironment

# query blockchain transaction/temperature records with user signing CA
curl -u User1: -X POST -H 'Content-Type: application/json' -d '{"uid":"2f850cc1cd8e670a","transactionType":"transfer"}' http://40.65.112.23:7979/shipping/verifytransaction
curl -u User1: -X POST -H 'Content-Type: application/json' -d '{"uid":"2f850cc1cd8e670a","periodStart":"2021-02-25T14:30:31Z"}' http://40.65.112.23:7979/shipping/verifytemperature

# query package content on blockchain
curl -u nlsadm: -X POST -H 'Content-Type: application/json' -d '{"uid":"2f850cc1cd8e670a"}' http://40.65.112.23:7979/shipping/getpackagebyuid
```

## Clean up all Azure resources

You can exit from the `bastion` host, and clean up every thing created in Azure when they are no longer used, i.e.,

```bash
cd ./az
./cleanup-bastion.sh dtwin westus2
```

This will clean up the bastion VM and resource group created in the previous step. Make sure that you supply the same parameters as that of the previous `create-bastion.sh create` command if they are different from the default values.
