#/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")"; echo "$(pwd)")"
TGDB_HOME=${TGDB_HOME:-"${HOME}/tibco/tgdb/3.0"}

if [ ! -f "${TGDB_HOME}/bin/tgdb" ]; then
  echo "Please configure TGDB_HOME env to root of the installed tgdb version"
  exit 1
fi

cd ${SCRIPT_DIR}
if [ ! -L "../lib" ]; then
  ln -s ${TGDB_HOME}/lib ../lib
fi

if [ ! -d "./data" ]; then
  mkdir ./data
  ${TGDB_HOME}/bin/tgdb -i -f -c ./tgdb.conf
fi

${TGDB_HOME}/bin/tgdb -s -c ./tgdb.conf 
