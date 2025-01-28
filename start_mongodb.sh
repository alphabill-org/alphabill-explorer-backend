#!/bin/bash

# Variables
CONTAINER_NAME="mongodb-container"
MONGO_IMAGE="mongodb/mongodb-community-server:latest"
MONGO_PORT=27017
MONGO_DATA_DIR="$(pwd)/mongo-data" # Replace with your desired data directory

# Ensure the data directory exists
if [ ! -d "$MONGO_DATA_DIR" ]; then
  echo "Creating MongoDB data directory at $MONGO_DATA_DIR..."
  mkdir -p "$MONGO_DATA_DIR"
fi

# Check if the container is already running
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
  echo "MongoDB container is already running."
  exit 0
fi

# Check if the container exists but is stopped
if [ "$(docker ps -aq -f name=$CONTAINER_NAME)" ]; then
  echo "Starting existing MongoDB container..."
  docker start "$CONTAINER_NAME"
else
  # Run a new MongoDB container
  echo "Starting a new MongoDB container..."
  docker run -d \
    --name "$CONTAINER_NAME" \
    -p $MONGO_PORT:27017 \
    -v "$MONGO_DATA_DIR:/data/db" \
    "$MONGO_IMAGE"
fi

# Check if the container is running successfully
if [ "$(docker ps -q -f name=$CONTAINER_NAME)" ]; then
  echo "MongoDB container is running. Access it at localhost:$MONGO_PORT."
else
  echo "Failed to start MongoDB container."
fi
