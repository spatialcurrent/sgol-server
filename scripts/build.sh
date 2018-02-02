#!/bin/bash
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

cd ../bin
go build github.com/spatialcurrent/sgol-server/cmd/sgol
