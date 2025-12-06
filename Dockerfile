# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled (required for webp package)
RUN CGO_ENABLED=1 GOOS=linux go build -o /build/mdserve main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/mdserve .

# Copy necessary directories
COPY --from=builder /build/config ./config
COPY --from=builder /build/content ./content
COPY --from=builder /build/templates ./templates
COPY --from=builder /build/assets ./assets
COPY --from=builder /build/user-static ./user-static

# Create directory for generated files
RUN mkdir -p .static

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./mdserve"]

