#!/bin/bash
#
# Docker Permission Fix Script
# This script fixes Docker permission issues for non-root users
# Run with: bash docker-permissions-fix.sh
#

set -e

echo "🔧 Fixing Docker permissions..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

CURRENT_USER=$(whoami)

# Check if user is root
if [ "$CURRENT_USER" = "root" ]; then
    echo "⚠️  Running as root. This script should be run as your regular user."
    exit 1
fi

# Add user to docker group (requires sudo)
echo "📝 Adding $CURRENT_USER to docker group..."
if ! groups "$CURRENT_USER" | grep &> /dev/null '\bdocker\b'; then
    sudo usermod -aG docker "$CURRENT_USER"
    echo "✓ User added to docker group"
else
    echo "✓ User already in docker group"
fi

# Activate docker group without logout
echo "🔄 Activating docker group in current session..."
newgrp docker <<'EOF'
echo "✓ Docker group activated for current session"
EOF

# Verify Docker socket permissions
echo "🔍 Checking Docker socket permissions..."
if [ -S /var/run/docker.sock ]; then
    ls -l /var/run/docker.sock
    echo "✓ Docker socket found"
else
    echo "❌ Docker socket not found at /var/run/docker.sock"
fi

# Test Docker access
echo "🧪 Testing Docker access..."
if docker ps &> /dev/null; then
    echo "✓ Docker access granted!"
else
    echo "⚠️  Docker access test failed. You may need to restart your terminal or run:"
    echo "   newgrp docker"
    exit 1
fi

echo ""
echo "✅ Docker permission fix completed!"
echo ""
echo "📌 Next steps:"
echo "   1. Close and reopen your terminal for changes to take effect"
echo "   2. Run: docker-compose up -d"
echo "   3. To stop: docker-compose down"
echo ""
