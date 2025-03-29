#!/bin/bash

# Change to the script directory
SCRIPT_DIR=$(dirname "$(realpath "$0")")
cd "$SCRIPT_DIR"

go build -o ./analysis2/analysis2 ./analysis2
go build -o ./backtest/backtest ./backtest
go build -o ./dicttomap/dicttomap ./dicttomap
go build -o ./goruut/goruut ./goruut
go build -o ./phondephontest/phondephontest ./phondephontest

echo "Goruut binaries built successfuly"