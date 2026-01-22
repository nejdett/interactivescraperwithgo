# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o collector ./cmd/collector

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

# Create app directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/collector .

# Copy sources configuration
COPY --from=builder /app/sources.json .

# Run as non-root user
RUN adduser -D -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

CMD ["./collector"]
