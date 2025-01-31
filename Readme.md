# Alphabill Block Explorer

## Requirements

* Docker

## Starting Block Explorer

Run `docker compose up --build` in project root directory. This will start the block explorer and MongoDB containers.
The explorer will start fetching blocks from all the configured partition nodes and a REST API server will be started at the configured address (http://localhost:9666 by default).

## Configuration

Configuration can be provided using env variables in `docker-compose.yml` or changing `cmd/config.yaml`.
If both the yaml file and env variables are provided, then env variables will take precedence over those in the yaml file.
Env variables must have the prefix `BLOCK_EXPLORER`, eg `BLOCK_EXPLORER_DB_URL`.
Below is a list of all the config parameters with examples and explanations:

```
BLOCK_EXPLORER_NODES_0_URL=dev-ab-money-archive.abdev1.guardtime.com/rpc - archive node to read blocks from
BLOCK_EXPLORER_NODES_0_BLOCK_NUMBER=100 - first block number to fetch, must be > 0
BLOCK_EXPLORER_NODES_1_URL=dev-ab-tokens-archive.abdev1.guardtime.com/rpc
BLOCK_EXPLORER_NODES_1_BLOCK_NUMBER=100
BLOCK_EXPLORER_DB_URL=mongodb://<username>:<password>@localhost:27017 - connection string for Mongo DB
BLOCK_EXPLORER_SERVER_ADDRESS=localhost:9666 - address of the REST API server
```

## Rest API

Documentation of REST API endpoints can be found at http://localhost:9666/swagger/index.html
