#!/bin/bash

# RAC Service Proxies - Development Startup Script
# This script starts both the Go backend and React frontend

set -e

echo "🚀 Starting RAC Service Proxies Development Environment"
echo "=================================================="

# Check if required environment variables are set
if [ -z "$GOOGLE_CLIENT_ID" ] || [ -z "$GOOGLE_CLIENT_SECRET" ]; then
    echo "❌ Error: Required environment variables not set"
    echo "Please set GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET"
    echo ""
    echo "Example:"
    echo "export GOOGLE_CLIENT_ID='your-client-id'"
    echo "export GOOGLE_CLIENT_SECRET='your-client-secret'"
    echo "export OAUTH_REDIRECT_URL='http://localhost:8080/api/auth/callback'"
    exit 1
fi

# Set default values for optional environment variables
export PORT=${PORT:-8080}
export OAUTH_REDIRECT_URL=${OAUTH_REDIRECT_URL:-"http://localhost:8080/api/auth/callback"}

echo "✅ Environment variables configured"
echo "   - GOOGLE_CLIENT_ID: ${GOOGLE_CLIENT_ID:0:20}..."
echo "   - OAUTH_REDIRECT_URL: $OAUTH_REDIRECT_URL"
echo "   - PORT: $PORT"
echo ""

# Build the Go backend
echo "🔨 Building Go backend..."
cd backend
if ! go build -o rac-server .; then
    echo "❌ Failed to build Go backend"
    exit 1
fi
echo "✅ Go backend built successfully"
echo ""

# Build the React frontend
echo "🔨 Building React frontend..."
cd ../frontend
if ! npm run build; then
    echo "❌ Failed to build React frontend"
    exit 1
fi
echo "✅ React frontend built successfully"
echo ""

# Start the backend server
echo "🚀 Starting backend server on port $PORT..."
cd ../backend
./rac-server &
BACKEND_PID=$!

# Wait a moment for the backend to start
sleep 2

# Check if backend is running
if ! curl -s http://localhost:$PORT/health > /dev/null; then
    echo "❌ Backend server failed to start"
    kill $BACKEND_PID 2>/dev/null || true
    exit 1
fi

echo "✅ Backend server started successfully"
echo ""

# Start the frontend development server
echo "🚀 Starting frontend development server..."
cd ../frontend
npm run dev &
FRONTEND_PID=$!

echo ""
echo "🎉 Development environment started successfully!"
echo "=================================================="
echo "📱 Frontend: http://localhost:5173"
echo "🔧 Backend:  http://localhost:$PORT"
echo "🔗 MCP WebSocket: ws://localhost:$PORT/mcp"
echo ""
echo "Available interfaces:"
echo "  • REST API Interface - Traditional HTTP API"
echo "  • MCP Protocol Interface - WebSocket-based MCP"
echo ""
echo "Press Ctrl+C to stop all services"

# Function to cleanup on exit
cleanup() {
    echo ""
    echo "🛑 Stopping services..."
    kill $BACKEND_PID 2>/dev/null || true
    kill $FRONTEND_PID 2>/dev/null || true
    echo "✅ All services stopped"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Wait for processes
wait
