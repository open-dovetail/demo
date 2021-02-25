#!/bin/bash
# Copyright © 2018. TIBCO Software Inc.
#
# This file is subject to the license terms contained
# in the license file that is distributed with this file.

# Execute this script on bastion host to initialize the host

sudo apt-get update

echo "install Golang 1.14.10"
GO_ZIP=go1.14.10.linux-amd64.tar.gz
curl -O https://storage.googleapis.com/golang/$GO_ZIP
sudo tar -xf $GO_ZIP -C /usr/local
mkdir -p ~/go/{bin,pkg,src}
echo "export GOPATH=$HOME/go" >> .profile
echo "export PATH=$HOME/go/bin:/usr/local/go/bin:$PATH" >> .profile
rm -f $GO_ZIP

# setup for dovetail
echo "setup dovetail"
mkdir open-dovetail
cd open-dovetail
git clone https://github.com/open-dovetail/fabric-chaincode.git
git clone https://github.com/open-dovetail/fabric-client.git
git clone https://github.com/open-dovetail/demo.git

# install fabric binary for chaincode packaging
mkdir hyperledger
cd hyperledger
curl -sSL http://bit.ly/2ysbOFE | bash -s -- 2.2.1 1.4.9
cd fabric-samples/test-network
sed -i -e "s/--remove-orphans//g" network.sh
sed -i -e "s/\$IMAGETAG/latest/g" network.sh
cd docker
sed -i -e "s/\${COMPOSE_PROJECT_NAME}_test/docker_test/g" *.yaml
sed -i -e "s/\$IMAGE_TAG/latest/g" *.yaml

cd $HOME
sudo apt-get install unzip
sudo apt-get install make

# this is required by TGDB
sudo apt-get install -y libffi-dev

if [ -f "./TIB_tgdb_3.0.0_linux_x86_64.zip" ]; then
  unzip ./TIB_tgdb_3.0.0_linux_x86_64.zip
  # find libffi lib location: ldconfig -p | grep libffi
  # then create link in tgdb/3.0/lib
  ln -s /usr/lib/x86_64-linux-gnu/libffi.so.6 ./tgdb/3.0/lib/libffi.so.5
fi

cd $HOME/open-dovetail/fabric-chaincode/scripts
./setup.sh
