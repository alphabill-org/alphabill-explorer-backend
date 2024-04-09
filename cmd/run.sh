#!/bin/bash

go run explorer.go 'https://money-partition.testnet.alphabill.org' 'localhost:9666' 254277

#
#curl -H 'Content-Type: application/json' \
#    -d '{"jsonrpc":"2.0","id":1,"method":"state_getRoundNumber"}' \
#    https://money-partition.testnet.alphabill.org/rpc