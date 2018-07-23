#!/bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

mkdir -p $DIR/../bin
NAME=sgol-server

echo "******************"
echo "Formatting $DIR/../cmd/sgol"
cd $DIR/../cmd/sgol
go fmt
echo "Formatting github.com/spatialcurrent/sgol-server/sgol"
go fmt github.com/spatialcurrent/sgol-server/sgol
echo "Done formatting."
echo "******************"
echo "Building plugin for $NAME"
cd $DIR/../bin
go build github.com/spatialcurrent/sgol-server/cmd/sgol
if [[ "$?" != 0 ]] ; then
    echo "Error building $NAME program"
    exit 1
fi

echo "Plugin at $(realpath $DIR/../bin/sgol)"
