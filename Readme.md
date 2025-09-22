# GoTrade - A Golang Microservices Trading Platform

GoTrade is a simplified, proof-of-concept trading platform built with a Golang microservices architecture. It aims to demonstrate core trading functionalities, clean architecture principles, and modern development practices within a containerized environment.

## Architecture

The system is composed of several independent microservices that communicate via REST APIs and a Redis Pub/Sub system. An Nginx server acts as the API gateway, routing client requests to the appropriate backend service.

```
+----------------+      +----------------+      +------------------------+
|                |      |                |      |                        |
|     Client     +------>      Nginx     +------>   Backend Services      |
| (JS Frontend)  |      |  (API Gateway) |      | (Go Microservices)     |
|                |      |                |      |                        |
+----------------+      +----------------+      +-----------+------------+
                                                            |
                                                            |
                                          +-----------------+-----------------+
                                          |                 |                 |
                                          v                 v                 v
                                +-----------------+ +-----------------+ +---------------+
                                |                 | |                 | |               |
                                |  user-service   | | order-service   | | market-service|
                                |                 | |                 | | (future)      |
                                +-------+---------+ +-------+---------+ +---------------+
                                        |                   |
                                        |                   |
                                        |  +----------------+----------------+
                                        |  |                |                |
                                        v  v                v                v
                                +-----------------+ +-----------------+
                                |                 | |                 |
                                |   PostgreSQL    | |      Redis      |
                                | (Transactional) | | (Pub/Sub, Cache)|
                                +-----------------+ +-----------------+
```

## Prerequisites

- Docker & Docker Compose
- Go (1.21+)
- `make`

## Setup & Running

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd gotrade
    ```

2.  **Create an environment file:**
    Copy the `.env.example` to `.env` and customize if needed. The defaults are set up to work with the `docker-compose.yml` file.

3.  **Start the application:**
    This command will build the Docker images and start all the services in the background.
    ```sh
    make up
    ```

4.  **Stop the application:**
    This command will stop and remove all the containers, networks, and volumes.
    ```sh
    make down
    ```

## Development Commands

The `Makefile` provides several commands to streamline development:

- `make build`: Build all service images.
- `make up`: Build and start all services.
- `make down`: Stop and remove all services and associated resources.
- `make restart`: A convenient shortcut for `make down && make up`.
- `make logs`: Tail the logs from all running services.
- `make restart-service service=<service_name>`: Restart a specific service (e.g., `make restart-service service=user-service`).

## Service Descriptions

- **Nginx**: The API gateway that routes incoming HTTP requests to the appropriate microservice.
- **user-service**: Manages user registration, login, and profile data. It is responsible for issuing JWTs for authentication.
- **instrument-service**: Provides information about tradable instruments (Equities, F&O). It includes a seeder to populate the database with initial data.
- **order-service**: Handles the core trading logic, including placing and canceling orders. It uses a transactional outbox pattern to publish order events to Redis for real-time updates.
- **market-service** (Future): Will be responsible for handling real-time market data feeds.
- **PostgreSQL**: The primary relational database for persistent, transactional data like users, orders, and instruments.
- **Redis**: Used for caching and as a real-time message broker (Pub/Sub) for events like order updates.

## API Documentation

API documentation will be provided via Swagger/OpenAPI in a future update. For now, refer to the handler files in each service for endpoint definitions.

- **User Service**: `user-service/internal/handler/user_handler.go`
- **Instrument Service**: `instrument-service/internal/handler/instrument_handler.go`
- **Order Service**: `order-service/internal/handler/order_handler.go`

## Environment Variables

Configuration is managed through environment variables. A `.env.example` file is provided as a template. When running `make up`, `docker-compose` automatically loads variables from a `.env` file in the project root.
