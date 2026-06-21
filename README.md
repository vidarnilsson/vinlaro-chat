# Chat App

A learning project covering: Go, PostgreSQL, MinIO, Kafka, WebSockets, and JWT auth.

## Prerequisites

- [Go 1.22+](https://go.dev/dl/)
- [Docker + Docker Compose](https://docs.docker.com/get-docker/)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [golang-migrate](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install golang-migrate
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Getting started

### 1. Start infrastructure

```bash
docker-compose up -d
```

This starts:
- **Postgres** on `localhost:5432`
- **MinIO** on `localhost:9000` (console at http://localhost:9001)
- **Kafka** on `localhost:9092`
- **Kafka UI** at http://localhost:8080

### 2. Run database migrations

```bash
cd backend
migrate -path db/migrations \
        -database "postgres://chat:chat@localhost:5432/chat?sslmode=disable" \
        up
```

### 3. Generate sqlc code

```bash
cd backend
sqlc generate
```

This reads your SQL queries in `db/queries/` and generates type-safe Go code into `internal/db/`.

### 4. Install Go dependencies

```bash
cd backend
go mod tidy
```

### 5. Run the API server

```bash
cd backend
go run ./cmd/api
```

API is available at http://localhost:8000

### 6. Run the Kafka worker (separate terminal)

```bash
cd backend
go run ./cmd/worker
```

## API endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Get JWT token |
| GET | `/api/channels` | Yes | List channels |
| POST | `/api/channels` | Yes | Create channel |

## Testing with curl

```bash
# Register
curl -X POST http://localhost:8000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8000/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"password123"}'

# List channels (use token from login response)
curl http://localhost:8000/api/channels \
  -H "Authorization: Bearer <your-token>"
```

## Project structure

```
chat-app/
├── docker-compose.yml          # All infrastructure
├── backend/
│   ├── cmd/
│   │   ├── api/main.go         # HTTP server entrypoint
│   │   └── worker/main.go      # Kafka consumer entrypoint
│   ├── internal/
│   │   ├── auth/               # JWT + bcrypt
│   │   ├── handler/            # HTTP handlers
│   │   ├── middleware/         # JWT middleware
│   │   ├── db/                 # sqlc generated code (don't edit)
│   │   ├── messaging/          # Kafka producer/consumer (next step)
│   │   └── storage/            # MinIO client (next step)
│   ├── db/
│   │   ├── migrations/         # SQL migration files
│   │   └── queries/            # SQL queries for sqlc
│   └── config/                 # App config from env vars
└── frontend/                   # Coming later
```

## What's next

1. WebSocket hub for real-time messaging
2. Kafka producer in the API (publish messages)
3. Kafka consumer in the worker (persist + broadcast)
4. MinIO integration for file attachments
5. Frontend in React
```
