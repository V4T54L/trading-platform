.PHONY: all build up down restart logs test lint

# Variables
DOCKER_COMPOSE = docker-compose

# Default command
all: up

# Build all services defined in docker-compose.yml
build:
	@echo "Building Docker images..."
	$(DOCKER_COMPOSE) build

# Start all services in detached mode
up:
	@echo "Starting all services..."
	$(DOCKER_COMPOSE) up -d --build

# Stop and remove all services
down:
	@echo "Stopping and removing all services..."
	$(DOCKER_COMPOSE) down

# Restart all services
restart: down up

# View logs for all services
logs:
	@echo "Tailing logs..."
	$(DOCKER_COMPOSE) logs -f

# Restart a specific service
restart-service:
	@echo "Restarting service: ${service}..."
	$(DOCKER_COMPOSE) restart ${service}

# Run tests (placeholder)
test:
	@echo "Running tests..."
	# Example: go test ./...

# Run linter (placeholder)
lint:
	@echo "Running linter..."
	# Example: golangci-lint run

```
