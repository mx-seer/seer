# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 go build -o seer ./cmd/seer

# Runtime stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/seer .
COPY --from=builder /build/config.example.yaml ./config.yaml

# Create data directory
RUN mkdir -p /app/data

EXPOSE 8080

CMD ["./seer"]
