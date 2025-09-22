# GoTrade - A Golang Microservices Trading Platform

GoTrade is a simplified, event-driven trading platform built with Golang microservices. It demonstrates key architectural patterns like Clean Architecture, Transactional Outbox, and real-time communication using WebSockets.

## Architecture

The system is composed of several microservices that communicate via REST APIs and a Redis Pub/Sub system. An Nginx instance acts as an API Gateway, routing client requests to the appropriate backend service.

```
+----------------+      +-----------------+      +------------------------+
|                |      |                 |      |   Backend Services     |
|     Client     +----->+   Nginx Gateway +----->+                        |
| (JS/HTML/CSS)  |      |      (8080)     |      |   - user-service       |
|                |      |                 |      |   - instrument-service |
+-------+--------+      +--------+--------+      |   - order-service      |
        |                        |               |   - account-service    |
        |                        |               +-----------+------------+
        |                        |                           |
        |                        |                           |
+-------v--------+      +--------v--------+      +-----------v------------+
|                |      |                 |      |                        |
| WebSocket GW   <------+      Redis      <------+      order-service     |
| (Real-time)    |      |   (Pub/Sub)     |      |      (Publishes)       |
+----------------+      +-----------------+      +------------------------+
                                                           |
                                                           |
                                                 +---------v---------+
                                                 |                   |
                                                 |    PostgreSQL     |
                                                 | (Transactional DB)|
                                                 +-------------------+
```

## Prerequisites

- Docker & Docker Compose
- Go (1.21+)
- `make`

## Setup & Running

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd gotrade
    ```

2.  **Environment Variables:**
    Each service looks for a `.env` file in its root directory. For local development, create a single `.env` file in the project root.
    ```bash
    cp .env.example .env
    ```
    *Note: The `docker-compose.yml` file is configured to pass the root `.env` file to all services.*

3.  **Start the application:**
    This command will build the Docker images and start all services in detached mode.
    ```bash
    make up
    ```

4.  **Stop the application:**
    This command will stop and remove all containers, networks, and volumes.
    ```bash
    make down
    ```

## Development Commands

- `make build`: Build all service images.
- `make up`: Build and start all services.
- `make down`: Stop and remove all services and associated volumes.
- `make restart`: A convenient shortcut for `make down && make up`.
- `make logs`: Tail the logs from all running services.
- `make restart-service service=<service_name>`: Restart a specific service (e.g., `make restart-service service=order-service`).

## Service Descriptions

- **Nginx**: Acts as the API Gateway, routing all incoming traffic from port `8080` to the appropriate backend service.
- **user-service**: Manages user registration, login, and profile retrieval. It issues JWTs for authentication.
- **instrument-service**: Provides a list of tradable instruments. It includes a one-off seeder to populate the database.
- **order-service**: Handles order placement and cancellation. It uses the Transactional Outbox pattern to publish order events to Redis for real-time updates.
- **account-service**: Manages user account balances and mock fund transfers (deposits/withdrawals).
- **websocket-gateway**: Subscribes to Redis channels (e.g., for order events) and broadcasts messages to all connected WebSocket clients.
- **PostgreSQL**: The primary transactional database for all services requiring persistent storage.
- **Redis**: Used for caching and as a real-time message broker (Pub/Sub).

## API Documentation

Each service follows a RESTful API design. The primary endpoints are defined in the `main.go` file of each service.

- **User Service**: `/api/user/register`, `/api/user/login`, `/api/user/me`
- **Instrument Service**: `/api/instruments/`
- **Order Service**: `/api/orders/`
- **Account Service**: `/api/account/`, `/api/account/transfer`
- **WebSocket**: `/ws`

## Environment Variables

Configuration is managed through environment variables. See `.env.example` for a complete list of required variables.
