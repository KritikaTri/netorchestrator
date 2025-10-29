#!/bin/bash

# Setup script for migrating from Docker to Podman
echo "🔄 Setting up Podman as Docker replacement..."

# Initialize Podman machine
echo "📦 Initializing Podman machine..."
podman machine init

# Start Podman machine
echo "🚀 Starting Podman machine..."
podman machine start

# Create Docker socket alias for compatibility
echo "🔗 Setting up Docker socket compatibility..."
sudo ln -sf /var/run/podman/podman.sock /var/run/docker.sock 2>/dev/null || true

# Add Docker alias to shell profile
echo "⚙️  Setting up Docker command alias..."
echo "# Podman Docker compatibility" >> ~/.zshrc
echo "alias docker='podman'" >> ~/.zshrc
echo "alias docker-compose='podman-compose'" >> ~/.zshrc

# Create systemctl alias for Linux containers
echo "alias systemctl='sudo systemctl'" >> ~/.zshrc

echo "✅ Podman setup complete!"
echo "📝 Please run 'source ~/.zshrc' or restart your terminal to use the aliases"
echo "🐳 You can now use 'docker' commands with Podman backend"

# Test the setup
echo "🧪 Testing Podman setup..."
podman --version
podman machine list