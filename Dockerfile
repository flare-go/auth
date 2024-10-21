# Start from the official Golang Alpine image as the builder stage
FROM golang:1.23.0-alpine3.19 AS builder

# Set the working directory in the container
WORKDIR /app

# Install necessary tools: git for version control and upx for binary compression
RUN apk add --no-cache git upx

# Set environment variables for Go build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies using go mod
# Use mount cache to speed up builds
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy the entire project
COPY . .

# Build argument for version, default to 1.0
ARG VERSION=1.0

# Build the Go application
# Use mount cache to speed up builds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -ldflags="-w -s -X main.Version=${VERSION}" \
    -trimpath -o auth ./cmd/api

# Compress the binary using UPX
RUN upx --best --lzma auth

RUN mkdir -p /app/configs/casbin /app/configs/firebase /app/configs/cockroach/ca

# Start a new stage from scratch for a smaller final image
FROM gcr.io/distroless/static-debian11

# Add metadata to the image
LABEL maintainer="Your Name <koopapapa@gmail.com>"
LABEL version=${VERSION}
LABEL description="Auth Service for GoFlare"

# Set the working directory in the container
WORKDIR /app

# Copy the binary
COPY --from=builder /app/auth .
COPY --from=builder /app/configs /app/configs/


# Expose port 8080
EXPOSE 8080

# Set the entrypoint to run the auth binary
ENTRYPOINT ["/app/auth"]
