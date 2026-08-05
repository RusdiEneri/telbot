# ============ Build Stage ============
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o telbot .

# ============ Runtime Stage ============
FROM alpine:latest

RUN apk add --no-cache ca-certificates tzdata bash

WORKDIR /app

# Create default persistent data directory
RUN mkdir -p /data

COPY --from=builder /app/telbot /app/telbot
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh

ENV TELBOT_DATA_DIR=/data
ENV TZ=Asia/Jakarta

VOLUME ["/data"]

# Expose OTP Webhook port (optional)
EXPOSE 8080

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["--bot"]
