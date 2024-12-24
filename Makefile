.PHONY: build run test clean test-email

# Build the application
build:
	go build -o bin/timemachine main.go

# Run the application
run:
	go run main.go

# Run a test email immediately
test-email:
	@echo "Sending test email..."
	@go run main.go

# Clean build artifacts
clean:
	rm -rf bin/

# Install dependencies
deps:
	go mod download

# Initialize the project (create necessary files if they don't exist)
init:
	@if [ ! -f .env ]; then \
		echo "Creating .env file..."; \
		cp .env.example .env || echo "No .env.example found. Please create .env manually."; \
	fi
	@if [ ! -f messages.json ]; then \
		echo "Creating messages.json file..."; \
		echo '["Your first message", "Your second message"]' > messages.json; \
	fi
	@go mod tidy

# Default target
all: clean build

# Help command
help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make run        - Run the application"
	@echo "  make test-email - Send a test email"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make deps       - Install dependencies"
	@echo "  make init       - Initialize project files"
	@echo "  make all        - Clean and build"
	@echo "  make help       - Show this help message" 