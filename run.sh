#!/bin/bash

make build

./build/abexplorer 'dev-ab-money-archive.abdev1.guardtime.com/rpc' 'localhost:9666' 1

#
#curl -H 'Content-Type: application/json' \
#    -d '{"jsonrpc":"2.0","id":1,"method":"state_getRoundNumber"}' \
#    https://money-partition.testnet.alphabill.org/rpc
