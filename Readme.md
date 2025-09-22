# GoTrade - A Golang Microservices Trading Platform

GoTrade is a simplified, event-driven trading platform built with a microservices architecture in Golang. It provides core functionalities for trading Equities and F&O, including account management, order placement, and real-time updates via WebSockets.

The project is designed to be run entirely within Docker containers, orchestrated by a single `docker-compose.yml` file, making setup and development straightforward.

## Architecture

The system is composed of several independent microservices that communicate with each other via REST APIs and a Redis Pub/Sub messaging system. An Nginx server acts as an API Gateway, routing all incoming client requests to the appropriate backend service.

```
+-----------------+      +------------------------+      +--------------------+
|   Web Client    |----->|   Nginx API Gateway    |----->|   User Service     |
| (JS, HTML, CSS) |      |       (Port 8080)      |      | (Auth, Profile)    |
+-----------------+      +------------------------+      +----------+---------+
       |                           |                                |
       | (WebSocket)               | (REST)                         | (PostgreSQL)
       |                           |                                |
+------v-------------+      +------v---------------+      +----------v---------+
| WebSocket Gateway  |<-----|   Order Service      |----->|   Account Service  |
+--------------------+      | (Order Management)   |      | (Funds, Balance)   |
       ^                     +----------+----------+      +----------+---------+
       |                                | (REST)                       |
       | (Redis Pub/Sub)                |                              | (PostgreSQL)
       |                                |                              |
+------v-------------+      +----------v-----------+      +----------v---------+
|       Redis        |----->|  Instrument Service  |<---->|     PostgreSQL     |
| (Pub/Sub, Cache)   |      |   (Instrument Data)  |      | (Transactional DB) |
+--------------------+      +----------------------+      +--------------------+
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

2.  **Configure Environment:**
    Copy the example environment file and customize it if needed (the defaults are suitable for local development).
    ```bash
    cp .env.example .env
    ```
    *Note: The `docker-compose.yml` file is configured to pass the root `.env` file to all services.*

3.  **Start the application:**
    Use the `Makefile` to build and start all services in detached mode.
    ```bash
    make up
    ```
    The frontend will be available at `http://localhost:8080`.

4.  **Stop the application:**
    To stop and remove all containers, networks, and volumes:
    ```bash
    make down
    ```

## Development Commands

The `Makefile` provides several commands to streamline development:

*   `make build`: Build Docker images for all services.
*   `make up`: Start all services (and rebuild if necessary).
*   `make down`: Stop and remove all services and their data.
*   `make restart`: A convenient shortcut for `make down && make up`.
*   `make logs`: Tail the logs from all running services.
*   `make restart-service service=<service_name>`: Restart a specific service (e.g., `make restart-service service=order-service`).
*   `make test`: (Placeholder) Run application tests.
*   `make lint`: (Placeholder) Run code linters.

## Service Descriptions

*   **Nginx:** Acts as the API Gateway, routing requests to the appropriate backend service.
*   **user-service**: Handles user registration, login (JWT issuance), and profile management.
*   **instrument-service**: Provides information about tradable instruments. It includes a one-off seeder to populate the database.
*   **order-service**: Manages the lifecycle of trade orders (placement, cancellation) and publishes order events to Redis.
*   **account-service**: Manages user account balances and mock fund transfers (deposits/withdrawals).
*   **websocket-gateway**: Manages client WebSocket connections and broadcasts real-time events received from Redis.
*   **PostgreSQL**: The primary transactional database for all services requiring persistent storage.
*   **Redis**: Used for caching and as a real-time message broker (Pub/Sub).

## API Documentation

API documentation is generated using Swagger/OpenAPI and is available for each service. Once the application is running, you can access the interactive Swagger UI for each service at the following endpoints:

*   **User Service:** `http://localhost:8080/api/user/swagger/index.html`
*   **Account Service:** `http://localhost:8080/api/account/swagger/index.html`
*   **Order Service:** `http://localhost:8080/api/orders/swagger/index.html`
*   **Instrument Service:** `http://localhost:8080/api/instruments/swagger/index.html`

### Key Endpoints

#### User Service (`/api/user`)
*   `POST /register`: Create a new user account.
*   `POST /login`: Authenticate a user and receive a JWT.
*   `GET /me`: (Authenticated) Get the current user's profile.

#### Account Service (`/api/account`)
*   `GET /`: (Authenticated) Get the user's account balance.
*   `POST /transfer`: (Authenticated) Deposit or withdraw funds.

#### Order Service (`/api/orders`)
*   `POST /`: (Authenticated) Place a new order.
*   `GET /`: (Authenticated) Get all open and historical orders for the user.
*   `DELETE /{orderID}`: (Authenticated) Cancel an open order.

#### Instrument Service (`/api/instruments`)
*   `GET /`: Get a list of all tradable instruments.

#### WebSocket Gateway (`/ws`)
*   `GET /ws`: Establish a WebSocket connection for real-time updates.

## Environment Variables

Configuration is managed through environment variables. See `.env.example` for a complete list of required variables.

