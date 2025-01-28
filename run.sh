#!/bin/bash

make build

./build/abexplorer './cmd/config.yaml'

#
#curl -H 'Content-Type: application/json' \
#    -d '{"jsonrpc":"2.0","id":1,"method":"state_getRoundNumber"}' \
#    https://money-partition.testnet.alphabill.org/rpc
