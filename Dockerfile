FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for fetch dependencies (if needed)
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o nexa ./cmd/nexa

FROM alpine:latest

WORKDIR /root/

# Install basic tools
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/nexa .
COPY --from=builder /app/config.yaml .

# Create necessary directories
RUN mkdir -p data sites

# Expose ports
# 53: DNS (UDP)
# 7000: Dashboard
# 8000: Gateway
# 8080: Admin
# 8081: Storage
# 8082: Chat
EXPOSE 53/udp 7000 8000 8080 8081 8082

CMD ["./nexa"]
