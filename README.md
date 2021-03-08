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

For example, assuming that you have downloaded Go from [here](https://golang.org/dl/), and extracted the package in `/usr/local/go`, you may setup the following environment varialbes for Go:

```bash
export GOROOT=/usr/local/go
mkdir $HOME/go
export GOPATH=$HOME/go
export PATH=$GOROOT/bin:$GOPATH/bin:$PATH
```

You can then use the following commands to install everything in `$HOME` directory:

```bash
unzip TIB_tgdb_3.0.0_macosx_x86_64.zip -d $HOME
export TGDB_HOME=$HOME/tgdb/3.0
mkdir $HOME/open-dovetail
cd $HOME/open-dovetail
git clone https://github.com/open-dovetail/fabric-chaincode.git
git clone https://github.com/open-dovetail/fabric-client.git
git clone https://github.com/open-dovetail/demo.git
cd fabric-chaincode/scripts
./setup.sh
```

The `setup.sh` will install Hyperledger Fabric binary and samples in `$HOME/open-dovetail/hyperledger/fabric-samples`.

If this is a new installation of `TGDB`, run the following command to test the installation and accept the license:

```bash
cd $TGDB_HOME/bin
./tgdb -i -f -c ./tgdb.conf
```

On some Linux environment, e.g., Ubuntu, if you see the following error when initializing TGDB:

```text
Could not find a viable libffi. Need to have libffi installed to run the TIBCO graph database server!
```

you can find the `libffi` location and then create a symbolic link as follows:

```bash
# find the location of libffi.so.x
ldconfig -p | grep libffi

# create symbolic link to the result, e.g., /usr/lib/x86_64-linux-gnu/libffi.so.6
ln -s /usr/lib/x86_64-linux-gnu/libffi.so.6 $TGDB_HOME/lib/libffi.so.5
```

## Start the backend components of the demo

Start all backend components by executing the script [demo/az/start-all.sh](https://github.com/open-dovetail/demo/blob/master/az/start-all.sh), and test components as described in [README.md](https://github.com/open-dovetail/demo/blob/master/az/README.md).

The [README.md](https://github.com/open-dovetail/demo/blob/master/az/README.md) also describes how to create and setup a Linux VM in Azure, and start all the components in the VM. The same startup script [start-all.sh](https://github.com/open-dovetail/demo/blob/master/az/start-all.sh) works both locally on a laptop, or on an Azure Linux VM.

## Cleanup all demo processes

When the test is complete, you can use the following script to shutdown and cleanup all the demo processes:

```bash
cd $HOME/open-dovetail/demo/az
./cleanup-all.sh
```
