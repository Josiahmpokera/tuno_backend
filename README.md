# Upatu Backend (Tuno)

This is the backend for the Rotational Group Savings & Messaging Platform (Upatu).

## Technology Stack

- **Language**: Go (Golang)
- **API**: REST (Gin) + gRPC (Optional)
- **Real-Time**: WebSocket (Gorilla)
- **Messaging Protocol**: Protobuf
- **Database**: PostgreSQL
- **Cache / PubSub**: Redis
- **Architecture**: Event-driven

## Prerequisites

- Go 1.20+
- Docker & Docker Compose

## Setup

1. **Clone the repository**:
   ```bash
   git clone <repository_url>
   cd tuno_backend
   ```

2. **Environment Variables**:
   Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

3. **Start Infrastructure**:
   Start PostgreSQL and Redis using Docker Compose:
   ```bash
   make docker-up
   ```

4. **Run the Application**:
   ```bash
   make run
   ```

## Project Structure

- `cmd/api`: Main entry point for the API server.
- `internal/config`: Configuration management.
- `internal/db`: Database and Redis connection logic.
- `internal/domain`: Domain models and interfaces.
- `internal/handler`: HTTP handlers (Gin).
- `internal/middleware`: HTTP middleware.
- `internal/repository`: Data access layer.
- `internal/service`: Business logic.
- `internal/websocket`: WebSocket hub and client handling.
- `pkg/logger`: Structured logging using Zap.
- `proto`: Protocol Buffer definitions.

## API Endpoints

- `GET /health`: Health check endpoint.
- `GET /ws`: WebSocket endpoint.

## Development

To tidy up dependencies:
```bash
make tidy
```

To stop infrastructure:
```bash
make docker-down
```
