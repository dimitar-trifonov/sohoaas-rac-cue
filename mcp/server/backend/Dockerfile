# Multi-stage build for Go service proxies
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy source code
COPY . .

# providers dependencies
RUN cd /app/providers/workspace && go mod download
# workflow dependencies 
RUN cd /app/workflow && go mod download

# Download dependencies
RUN cd /app && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o service-proxies .

# Final stage
FROM alpine:latest

# Install ca-certificates and curl for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/service-proxies .

# Create directory for credentials
RUN mkdir -p /app/credentials && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run
CMD ["./service-proxies"]
