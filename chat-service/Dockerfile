# Start from the official Golang image for build
FROM golang:1.23-alpine AS builder

# Set the working directory for the build stage.
WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod and sum files from the build context
COPY go.mod go.sum ./

# Download all the go modules
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
RUN go build -o chat-service .

# ---

# Use a minimal image for running
FROM alpine:latest

# Set the working directory for the final image.
WORKDIR /app

# Install CA certificates
RUN apk add --no-cache ca-certificates

# Copy the built binary from the builder stage
COPY --from=builder /app/chat-service .

# Copy static files (if any)
# Since index.html is in the same directory as the Dockerfile
COPY index.html .

# Expose the port
EXPOSE 8080

# Run the binary
CMD ["./chat-service"]
