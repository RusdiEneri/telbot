# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install dependencies required for build
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o telbot .

# Production Runtime Stage
FROM alpine:latest

# Install CA certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Create default persistent data directory
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /app/telbot /app/telbot

# Set default Environment Variables
ENV TELBOT_DATA_DIR=/data
ENV TZ=Asia/Jakarta

# Declare persistent volume mount point
VOLUME ["/data"]

# Expose OTP Webhook port (optional)
EXPOSE 8080

# Run in --bot mode by default
ENTRYPOINT ["/app/telbot", "--bot"]
