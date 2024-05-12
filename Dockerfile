# Use a multi-stage build for efficiency
FROM golang:1.22.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy project source code
COPY . .

# Build the application, with 'main' as the name for the compile binary file
RUN go build -o main ./cmd/main.go

# Use a smaller alpine image for the final container
FROM alpine:latest

# Copy the compiled binary
COPY --from=builder /app/main /app/main

# Install MongoDB driver
RUN apk add --no-cache libc++ libstdc++

# Copy the configuration file
COPY config.yaml ./config.yaml

# Exposr REST API port
EXPOSE 8081

# Expose gRPC port 
EXPOSE 8083

# Set working directory
WORKDIR /app

# Run the application
CMD ["./main"]