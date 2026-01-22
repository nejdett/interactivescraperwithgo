# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o dashboard ./cmd/dashboard

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/dashboard .

# Copy web assets
COPY --from=builder /app/web ./web

# Run as non-root user
RUN adduser -D -u 1000 appuser
RUN chown -R appuser:appuser /root
USER appuser

EXPOSE 8080

CMD ["./dashboard"]
