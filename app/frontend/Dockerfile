# Multi-stage build for SOHOAAS Frontend
# Stage 1: Build React application
FROM node:20-alpine as build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:alpine

# Install envsubst for environment variable substitution and curl for health checks
RUN apk add --no-cache gettext curl

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy nginx template
COPY nginx.conf.template /etc/nginx/templates/default.conf.template

# Copy built React app
COPY --from=build /app/dist /usr/share/nginx/html

# Copy and setup entrypoint script
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Environment variables with defaults
ENV NGINX_PORT=3000
ENV MCP_SERVICE_URL=http://mcp-service:8080
ENV BACKEND_SERVICE_URL=http://sohoaas-backend:8081

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

# Start nginx with environment variable substitution
ENTRYPOINT ["/docker-entrypoint.sh"]
