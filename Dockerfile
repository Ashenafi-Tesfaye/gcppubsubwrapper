# Start from the official Golang image for building
FROM golang:1.21-alpine AS builder
LABEL maintainer="Ashenafi Gebreegziabhere <your-email@example.com>"

WORKDIR /app

# Install git for go mod download if needed
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o proxy ./cmd/proxy/main.go

# Use a minimal image for running
FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/Ashenafi-Tesfaye/dependency-wrapper"
WORKDIR /app

# Create non-root user
RUN adduser -D appuser

# Copy the built binary from builder
COPY --from=builder /app/proxy .

# Expose the default port
EXPOSE 8080

# Healthcheck for container orchestration
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD wget --spider -q http://localhost:8080/healthz || exit 1

# Switch to non-root user
USER appuser

# Command to run
ENTRYPOINT ["/app/proxy"]
