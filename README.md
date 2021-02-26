# Global Logistics Services Demo

This demo uses Dovetail and TIBCO graph DB to demonstrate a digital-twin type of application for global logicstics services that manages the pickup and delivery packages by multiple carriers. The blockchain tracks key milestone events generated when environment-sensitive goods are picked up, delivered, and/or transferred between carriers, as well as IoT events that shows violation of required shipping conditions.

## Required Components

- TIBCO graph Database 3.0
- Hyperledger Fabric 2.2.1
- Golang 1.14 or above

## Installation

- Download and install TIBCO graph database, and set `$TGDB_HOME` to the installation directory, e.g., `$HOME/tibco/tgdb/3.0`.
- Download the following 3 open-dovetail repositories in an empty folder, e.g. `open-dovetail`: [fabric-chaincode](https://github.com/open-dovetail/fabric-chaincode), [fabric-client](https://github.com/open-dovetail/fabric-client), and [demo](https://github.com/open-dovetail/demo).
- Initialize the Dovetail development environment by executing the script [fabric-chaincode/scripts/setup.sh](https://github.com/open-dovetail/fabric-chaincode/blob/master/scripts/setup.sh).

## Start the backend components of the demo

Start all backend components by executing the script [demo/az/start-all.sh](https://github.com/open-dovetail/demo/blob/master/az/start-all.sh), and test components as described in [README.md](https://github.com/open-dovetail/demo/blob/master/az/README.md).

The [README.md](https://github.com/open-dovetail/demo/blob/master/az/README.md) also describes how to create and setup a Linux VM in Azure, and start all the components in the VM. The same startup script [start-all.sh](https://github.com/open-dovetail/demo/blob/master/az/start-all.sh) works both locally on a laptop, or on an Azure Linux VM.
