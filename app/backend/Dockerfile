# Multi-stage build for SOHOAAS Backend
# Stage 1: Build Go application
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy genkit directory
COPY ./app/backend/genkit/go ./genkit/go

# Copy go mod files from backend directory
COPY ./app/backend/go.mod ./app/backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy backend source code
COPY ./app/backend .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o sohoaas-backend \
    ./main.go

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl tzdata

# Create non-root user (will be overridden by docker-compose user mapping)
RUN addgroup -g 1000 -S sohoaas && \
    adduser -u 1000 -S sohoaas -G sohoaas

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/sohoaas-backend .

# Copy configuration files and templates
COPY --from=builder /app/prompts ./prompts/

# Copy RaC context files (Requirements-as-Code specifications)
COPY ./rac ./rac/

# Create directories for runtime data with proper permissions
RUN mkdir -p /app/generated_workflows /app/logs && \
    chown -R sohoaas:sohoaas /app

# Switch to non-root user
USER sohoaas

# Environment variables with defaults
ENV PORT=8081
ENV GIN_MODE=release
ENV MCP_SERVICE_URL=http://mcp-service:8080
ENV GENKIT_ENV=production
ENV LOG_LEVEL=info
ENV RAC_CONTEXT_PATH=./rac

# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8081/health || exit 1

# Run the application
CMD ["./sohoaas-backend"]
