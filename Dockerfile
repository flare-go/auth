# Build stage
FROM golang:1.23.0-alpine3.19 AS builder

# Set the working directory
WORKDIR /build

# Set Go environment variables
ENV GOPATH /go
ENV GOCACHE /go-build

# Install necessary build tools
RUN apk update && apk add --no-cache git

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code and configuration files
COPY . .

# Compile the application
# -ldflags="-w -s" removes debugging information and symbol tables to reduce the size of the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -a -installsuffix cgo -o auth ./cmd/api

# Final stage
FROM alpine:3.19

# Add metadata
LABEL maintainer="Your Name <your.email@example.com>"
LABEL version="1.0"
LABEL description="Auth Service for GoFlare"

# Set the working directory
WORKDIR /app

# Copy the binary and configuration files from the build stage
COPY --from=builder /build/auth .
COPY --from=builder /build/config.yaml .
COPY --from=builder /build/firebase-service-account.json .
COPY --from=builder /build/casbin.conf .

# Create a non-root user
RUN adduser -D -g '' appuser

# Ensure the application user has permission to read the config files
RUN chown -R appuser:appuser /app

# Switch to the non-root user
USER appuser

# Specify the container start command
CMD ["./auth"]

# Declare the port the container will use
EXPOSE 50051
