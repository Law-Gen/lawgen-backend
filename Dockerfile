# Start from the official Golang image for build
FROM golang:1.22-alpine AS builder

WORKDIR /app/chat-service

# Install git (for go mod) and ca-certificates
RUN apk add --no-cache git ca-certificates


# Copy go mod and sum files
COPY chat-service/go.mod chat-service/go.sum ./

# Make sure we are in the right directory for go mod download
WORKDIR /app/chat-service
RUN go mod download

# Copy the rest of the code
COPY chat-service/. ./

# Build the Go app
RUN go build -o chat-service .

# ---


# Use a minimal image for running
FROM alpine:latest

WORKDIR /app

# Install CA certificates for MongoDB Atlas TLS/SSL support
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/chat-service/chat-service .

# Copy static files if needed (e.g., index.html)
COPY chat-service/index.html ./

# Expose the port (change if needed)
EXPOSE 8080

# Run the binary
CMD ["./chat-service"]