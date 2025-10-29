# Backend Server

## Prerequisites

- Go 1.23.2 or higher
- Docker and Docker Compose

## Project Structure

```
├── cmd/            # Application entrypoints
│   └── api/        # Main API server
├── internal/       # Private application code
```

### Internal

```
.
├── controllers
│   ├── health_controller.go
│   └── user_controller.go
└── routers
    └── router.go
```

## Local Development

> Using CLI to run the Golang + Gin server
```bash
go run cmd/api/main.go
```

> To tidy
```bash
go mod tidy
```

> To get access token
```bash
POST http://localhost:8080/api/v1/auth/login
```

## CockroachDB Migration

> Using CLI to do database migration
```bash
go run ./cmd/migrate
```


