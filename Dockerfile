# Test stage - runs tests before building final image
FROM golang:1.23-alpine AS tester

WORKDIR /app

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Run tests and vet
RUN go test -v ./... && go vet ./...

# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Build args for cross-compilation
ARG TARGETOS=linux
ARG TARGETARCH=amd64

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags='-w -s' -o org-charm .

# Runtime stage
FROM alpine:3.19 AS runtime

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Create non-root user and group
RUN addgroup -g 1000 orgcharm && \
    adduser -u 1000 -G orgcharm -s /bin/sh -D orgcharm

# Copy the binary from builder
COPY --from=builder /app/org-charm /app/org-charm

# Create directory for org files and set ownership
RUN mkdir -p /data && chown -R orgcharm:orgcharm /data /app

# Expose SSH port (use 2222 internally, map to 22 externally with -p 22:2222)
EXPOSE 2222

# Switch to non-root user
USER orgcharm

# Run the server
ENTRYPOINT ["/app/org-charm"]
CMD ["-port", "2222", "-dir", "/data"]
