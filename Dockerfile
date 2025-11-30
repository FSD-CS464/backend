# --- Stage 1: build the Go binary ---
FROM golang:1.22-alpine AS builder

ENV GOTOOLCHAIN=auto

WORKDIR /app

# Install CA certs + git just in case
RUN apk add --no-cache ca-certificates git

# Go deps
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the API binary (adjust ./cmd/api if needed)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/api ./cmd/api

# --- Stage 2: minimal runtime image ---
FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/api /app/api

EXPOSE 8080

CMD ["/app/api"]
