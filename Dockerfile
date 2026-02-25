# Build stage using Nix
FROM nixos/nix:latest AS builder

WORKDIR /app

# Enable flakes
RUN echo "experimental-features = nix-command flakes" >> /etc/nix/nix.conf

# Copy source code
COPY . .

# Build using devenv/nix
RUN nix-shell -p go_1_23 --run "CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o org-charm ."

# Runtime stage
FROM alpine:3.19

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
