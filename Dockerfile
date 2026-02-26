# Test stage - runs tests before building final image
FROM golang:1.26-alpine AS tester

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

# Copy go mod files first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Run tests and vet
RUN go test -v ./... && go vet ./...

# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git

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

# Export stage - minimal stage for extracting just the binary
FROM scratch AS binary
COPY --from=builder /app/org-charm /org-charm

# Runtime stage
FROM alpine:3.19 AS runtime

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates
# 
# Copy the binary from builder
COPY --from=builder /app/org-charm /app/org-charm

# Create directory for org files
RUN mkdir -p /data

# Expose SSH port (use 2222 internally, map to 22 externally with -p 22:2222)
EXPOSE 2222

# Run the server
ENTRYPOINT ["/app/org-charm"]
CMD ["-port", "2222", "-dir", "/data"]
