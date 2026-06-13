# Build Stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Download dependencies separately to leverage Docker layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy application source code
COPY . .

# Compile static binary with optimizations disabled for security, CGO disabled, and stripped symbols
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o loadster main.go

# Production Stage
FROM alpine:3.19

# Install security updates and utilities for readiness/liveness checks
RUN apk update && \
    apk upgrade && \
    apk add --no-cache ca-certificates curl && \
    rm -rf /var/cache/apk/*

# Create a non-root system user and group for running the container securely
RUN addgroup -g 10001 -S loadster && \
    adduser -u 10001 -S -G loadster loadster

WORKDIR /home/loadster

# Copy the compiled statically-linked binary from builder stage
COPY --from=builder /app/loadster /usr/local/bin/loadster

# Set ownership to the non-root user
RUN chown -R loadster:loadster /home/loadster

# Run as non-root system user
USER loadster

# Expose default HTTP ports for Prometheus exporter and REST APIs
EXPOSE 9090

ENTRYPOINT ["/usr/local/bin/loadster"]
CMD ["--help"]
