# Define variables
GO_BIN=./ecr-cleaner
DOCKER_IMAGE_TAG=ecr-cleaner:latest

# Default target
all: build docker

# Build the Go binary
build:
	@echo "Building Go binary..."
	go build -o ../bin/$(GO_BIN) main.go

# Build the Docker image
docker:
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE_TAG) .

# Push Docker image to the repository
push:
	@echo "Pushing Docker image to repository..."
	docker push $(DOCKER_IMAGE_TAG)

# Build and push Docker image
build-and-push: build docker push

# Run the application locally
run:
	@echo "Running the Go binary..."
	./$(GO_BIN)

# Display help
help:
	@echo "Makefile for Go project"
	@echo "Usage:"
	@echo "  make build          Build the Go binary"
	@echo "  make docker         Build the Docker image"
	@echo "  make clean          Remove 
