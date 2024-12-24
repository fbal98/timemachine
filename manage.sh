#!/bin/bash

cleanup() {
    echo "Cleaning up existing containers..."
    docker stop timemachine-local 2>/dev/null || true
    docker rm timemachine-local 2>/dev/null || true
}

case "$1" in
  "start")
    echo "Building and starting container..."
    cleanup
    docker build -t timemachine .
    docker run -d -p 8080:8080 \
      --env-file .env \
      -v $(pwd)/messages.json:/app/messages.json \
      --name timemachine-local \
      timemachine
    echo "Container started! Access the UI at http://localhost:8080"
    ;;
    
  "stop")
    echo "Stopping container..."
    cleanup
    ;;
    
  "logs")
    echo "Showing container logs..."
    docker logs -f timemachine-local
    ;;
    
  "restart")
    echo "Restarting container..."
    cleanup
    docker run -d -p 8080:8080 \
      --env-file .env \
      -v $(pwd)/messages.json:/app/messages.json \
      --name timemachine-local \
      timemachine
    echo "Container restarted! Access the UI at http://localhost:8080"
    ;;
    
  "clean")
    echo "Deep cleaning..."
    cleanup
    docker rmi timemachine 2>/dev/null || true
    echo "Removed all containers and images"
    ;;
    
  *)
    echo "Usage: $0 {start|stop|logs|restart|clean}"
    echo "  start   - Build and start the container"
    echo "  stop    - Stop and remove the container"
    echo "  logs    - Show container logs"
    echo "  restart - Restart the container"
    echo "  clean   - Remove all containers and images"
    exit 1
    ;;
esac 