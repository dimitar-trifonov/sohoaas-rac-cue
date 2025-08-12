#!/bin/bash

# Service Proxies Startup Script
# This script starts the service proxy backend and frontend applications

set -e

echo "🚀 Starting Service Proxy Infrastructure..."
echo "=============================================="

# Check if we're in the service-proxies directory
if [ ! -f "docker-compose.yaml" ]; then
    echo "❌ Error: docker-compose.yaml not found. Please run this script from the service-proxies directory."
    exit 1
fi

# Check if credentials exist
if [ ! -f "../credentials/google-credentials.json" ]; then
    echo "⚠️  Warning: Google credentials not found at ../credentials/google-credentials.json"
    echo "   Please follow the setup instructions in ../credentials/README.md"
    echo "   The services will start but OAuth authentication will not work."
    echo ""
fi

# Build and start services
echo "🔨 Building and starting services..."
docker-compose up --build -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

# Check service health
echo ""
echo "🔍 Service Status:"
echo "=================="

# Check backend
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✅ Service Proxy Backend: Running on http://localhost:8080"
else
    echo "⚠️  Service Proxy Backend: Not responding (may still be starting)"
fi

# Check frontend
if curl -s http://localhost:3002/health > /dev/null; then
    echo "✅ Service Proxy Frontend: Running on http://localhost:3002"
else
    echo "⚠️  Service Proxy Frontend: Not responding (may still be starting)"
fi

# Check optional services
if docker-compose ps | grep -q "redis.*Up"; then
    echo "✅ Redis: Running on port 6379"
fi

if docker-compose ps | grep -q "postgres.*Up"; then
    echo "✅ PostgreSQL: Running on port 5432"
fi

echo ""
echo "🎉 Service Proxy Infrastructure Started!"
echo "========================================"
echo ""
echo "📋 Access Points:"
echo "• Backend API: http://localhost:8080"
echo "• Frontend Management: http://localhost:3002"
echo "• API Documentation: http://localhost:8080/providers"
echo ""
echo "🔧 Next Steps:"
echo "1. Visit http://localhost:8080/auth/login to authenticate with Google"
echo "2. Use the management interface at http://localhost:3002"
echo "3. Test service proxy endpoints"
echo "4. Integrate with SOHOaaS Implementation Agent"
echo ""
echo "📊 Management Commands:"
echo "• View logs: docker-compose logs -f"
echo "• Stop services: docker-compose down"
echo "• Restart services: docker-compose restart"
echo "• View status: docker-compose ps"
echo ""
echo "🔗 SOHOaaS Integration:"
echo "• Main SOHOaaS Frontend: http://localhost:3000 (when running)"
echo "• Main SOHOaaS Backend: http://localhost:3001 (when running)"
echo "• Service Proxy API for SOHOaaS: http://localhost:8080"
