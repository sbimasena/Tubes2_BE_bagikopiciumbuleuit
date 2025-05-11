# Use official Golang image as the build stage
FROM golang:1.24-alpine AS builder

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of your source code
COPY . .

# Build the Go app (output: /app/server)
RUN go build -o server .

# Final stage: use minimal image
FROM alpine:latest

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose the port (change if needed)
EXPOSE 8080

# Run the binary
CMD ["./server"]
