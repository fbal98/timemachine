#!/bin/bash

echo "Cleaning up any existing containers..."
docker stop timemachine-local 2>/dev/null || true
docker rm timemachine-local 2>/dev/null || true

echo "Building Docker image..."
docker build -t timemachine .

echo "Starting container..."
docker run -p 8080:8080 \
  --env-file .env \
  -v $(pwd)/messages.json:/app/messages.json \
  --name timemachine-local \
  timemachine

# The container will keep running. To stop it, you can use:
# docker stop timemachine-local 