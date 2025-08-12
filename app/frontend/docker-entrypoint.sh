#!/bin/sh

# Substitute only our environment variables, not nginx variables
envsubst '$NGINX_PORT $MCP_SERVICE_URL $BACKEND_SERVICE_URL' < /etc/nginx/templates/default.conf.template > /etc/nginx/conf.d/default.conf

# Start nginx
exec nginx -g "daemon off;"
