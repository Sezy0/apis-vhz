# VinzHub REST API

A production-ready REST API with realtime WebSocket support, built with Go.

## Features

- ✅ RESTful API with versioning (`/api/v1`)
- ✅ WebSocket support for realtime events
- ✅ Clean architecture with strict layer separation
- ✅ Interface-based dependencies (easy to swap implementations)
- ✅ Docker-ready with multi-stage builds
- ✅ Horizontal scaling ready (Redis abstractions)
- ✅ Graceful shutdown
- ✅ Health check endpoints

## Architecture

```
Transport Layer (HTTP/WebSocket)
        ↓
  Business Layer (Services)
        ↓
Infrastructure Layer (Repository/Cache/EventBus)
```

### Layer Responsibilities

| Layer | Responsibility |
|-------|----------------|
| **Handler** | HTTP parsing, validation, response formatting |
| **Service** | Business logic, orchestration |
| **Repository** | Data access, storage abstraction |
| **Cache** | Caching abstraction (memory/Redis) |
| **EventBus** | Pub/sub for async processing |

## Quick Start

### Using Docker

```bash
# Development
docker-compose up --build

# Production (with Redis)
REDIS_PASSWORD=your_password docker-compose -f docker-compose.prod.yml up -d
```

### Local Development

```bash
# Install dependencies
go mod download

# Run the API
go run ./cmd/api
```

## API Endpoints

### Health

```bash
# Liveness probe
curl http://localhost:8080/api/v1/health

# Readiness probe
curl http://localhost:8080/api/v1/ready
```

### Users

```bash
# List users
curl http://localhost:8080/api/v1/users

# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com"}'

# Get user
curl http://localhost:8080/api/v1/users/{id}

# Update user
curl -X PUT http://localhost:8080/api/v1/users/{id} \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/users/{id}
```

### WebSocket

```bash
# Connect using wscat
wscat -c ws://localhost:8080/api/v1/ws

# Subscribe to users channel
> {"type": "subscribe", "channel": "users"}

# You'll receive events when users are created/updated/deleted
```

## Configuration

All configuration is done via environment variables. See `.env.example` for available options.

## Project Structure

```
├── cmd/api/              # Application entrypoint
├── internal/
│   ├── config/           # Configuration loading
│   ├── domain/           # Domain models and errors
│   ├── service/          # Business logic
│   ├── repository/       # Data access layer
│   ├── cache/            # Cache abstraction
│   ├── event/            # Event bus abstraction
│   └── transport/
│       ├── http/         # HTTP handlers, middleware, router
│       └── ws/           # WebSocket hub, client, handler
├── pkg/
│   ├── apierror/         # API error types
│   └── uid/              # ID generation
├── Dockerfile            # Multi-stage Docker build
├── docker-compose.yml    # Development setup
└── docker-compose.prod.yml # Production setup with Redis
```

## Scaling

The application is designed for horizontal scaling:

1. **Cache**: Swap `MemoryCache` for `RedisCache`
2. **Event Bus**: Swap `MemoryEventBus` for `RedisEventBus`
3. **WebSocket**: Use Redis pub/sub for cross-instance messaging
4. **Database**: Implement PostgreSQL repository

## License

MIT
