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

