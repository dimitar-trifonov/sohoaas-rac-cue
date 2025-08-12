# Multi-stage build for React frontend
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Production stage
FROM nginx:alpine

# Install curl for health checks and debugging
RUN apk add --no-cache curl

# Copy built app to nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx template configuration
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Set default backend service name if not provided
ENV BACKEND_SERVICE_NAME=backend

# Expose port
EXPOSE 3002

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3002 || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
