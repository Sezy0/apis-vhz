#!/bin/bash
# VinzHub REST API - Quick Deploy Script

echo "🚀 VinzHub REST API Deployment"
echo "================================"

# Stop existing container
echo "⏹️  Stopping existing container..."
docker stop vinzhub-api 2>/dev/null || true
docker rm vinzhub-api 2>/dev/null || true

# Build new image
echo "🔨 Building Docker image..."
docker build -t vinzhub-api .

# Run container
echo "▶️  Starting container..."
docker run -d \
    --name vinzhub-api \
    --restart unless-stopped \
    -p 8080:8080 \
    --env-file .env \
    vinzhub-api

# Wait and check
sleep 3
echo ""
echo "✅ Deployment complete!"
echo ""

# Health check
echo "🏥 Health check:"
curl -s http://localhost:8080/api/v1/health || echo "⚠️  API not responding yet, wait a moment"

echo ""
echo "📋 Container status:"
docker ps | grep vinzhub-api

echo ""
echo "📝 View logs with: docker logs -f vinzhub-api"
