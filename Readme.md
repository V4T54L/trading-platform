# GoTrade - A Golang Microservices Trading Platform

GoTrade is a simplified, containerized trading platform for Equities and F&O, built with a Golang microservices backend, PostgreSQL, Redis, and a vanilla JavaScript frontend. The project emphasizes Clean Architecture principles and is designed for local development and demonstration using Docker Compose.

---

### Architecture

The system is composed of several microservices that communicate via REST APIs and a Redis Pub/Sub system for real-time events. Nginx acts as a reverse proxy and API gateway.

**High-Level Diagram:**

```
+----------------+      +-----------------+      +--------------------+
|   Client UI    |----->|      Nginx      |----->|   Backend Services |
| (JS, HTML, CSS)|      | (Reverse Proxy) |      | (Go Microservices) |
+----------------+      +-------+---------+      +----------+---------+
                                |                           |
                                |                           |
      +-------------------------+---------------------------+
      |                         |                           |
+-----v-----+           +-------v--------+          +-------v-------+
| user-service |           | order-service  |          | market-service|
+-----------+           +----------------+          +---------------+
      |                         |                           |
      |                         |                           |
+-----v-----+           +-------v--------+          +-------v-------+
| PostgreSQL|           |      Redis     |          | (WebSocket to UI) |
+-----------+           +----------------+          +---------------+

```

---

### Prerequisites

*   Docker
*   Docker Compose
*   Go (for local development outside Docker)
*   `make`

---

### Setup & Running

1.  **Clone the repository:**
    ```sh
    git clone <repository-url>
    cd gotrade
    ```

2.  **Start the application:**
    Use the Makefile to build and start all services in detached mode.
    ```sh
    make up
    ```
    The application will be available at `http://localhost:8080`.

3.  **Stop the application:**
    To stop and remove all containers, networks, and volumes:
    ```sh
    make down
    ```

---

### Development Commands

The `Makefile` provides several commands to streamline development:

*   `make build`: Build or rebuild the service images.
*   `make up`: Start all services.
*   `make down`: Stop and remove all services.
*   `make restart`: A convenient shortcut for `make down && make up`.
*   `make logs`: Tail the logs from all running services.
*   `make restart-service service=<service_name>`: Restart a specific service (e.g., `make restart-service service=user-service`).

---

### Service Descriptions

*   **Nginx:** The entry point for all traffic. Acts as a reverse proxy, routing API requests to the appropriate backend service and serving the static frontend application.
*   **user-service:** Manages user authentication, registration, and profile information. Issues JWTs for securing the platform.
*   **order-service:** Handles all trading logic, including order placement, modification, and cancellation.
*   **market-service:** Provides market data, instrument information, and real-time price updates to the client via WebSockets.
*   **PostgreSQL:** The primary relational database for persistent data like user accounts, orders, positions, and instruments.
*   **Redis:** Used for caching session data and as a high-speed message broker (Pub/Sub) for real-time event propagation between services.

---

### API Documentation

API documentation is generated using Swagger/OpenAPI. Once the services are running, you can access the Swagger UI for each service at:

*   **User Service:** `http://localhost:8080/api/user/swagger/index.html` (To be implemented)

---

### Environment Variables

Configuration for each service is managed via environment variables set within the `docker-compose.yml` file. See the respective service sections for details. A `.env` file can be used to override default values.

```
# .env (Example)
POSTGRES_USER=admin
POSTGRES_PASSWORD=secret
POSTGRES_DB=gotrade
```
