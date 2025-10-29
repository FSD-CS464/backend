# Backend Server

## Prerequisites

- Go 1.23.2 or higher
- Docker and Docker Compose
- Cockroach DB Cluster

## Project Structure

```
├── cmd/            # Application entrypoints
│   └── api/        # Main API server
├── internal/       # Private application code
```

### Internal

```
.
├── app
│   ├── config.go
│   └── server.go
├── auth
│   └── jwt.go
├── controllers
│   ├── auth_controller.go
│   ├── health_controller.go
│   ├── user_controller.go
│   └── ws_controller.go
├── db
│   └── db.go
├── middleware
│   ├── auth_jwt.go
│   ├── cors.go
│   ├── prometheus.go
│   ├── recovery.go
│   └── requestid.go
├── repository
│   ├── pet_repo.go
│   ├── types.go
│   └── user_repo.go
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

{
  "username": "demo",
  "password": "demo"
}

```

## CockroachDB Migration

> Using CLI to do database migration
```bash
go run ./cmd/migrate
```

> Seed data into database
```bash
go run ./cmd/seed
```

> To roll down goose for reset schema
```bash
go run ./cmd/migrate -down
```


