# Docker to Podman Migration Guide

## Overview
This guide helps migrate your NetOrchestrator project from Docker Desktop to Podman Desktop, ensuring compliance with Cisco's software policy.

## Why Podman?
- ✅ **Cisco Approved**: Free and compliant with Cisco policies
- ✅ **Docker Compatible**: Uses same commands and compose files
- ✅ **Rootless**: More secure by default
- ✅ **No Daemon**: Lighter resource usage

## Installation Steps

### 1. Install Podman Desktop
```bash
# Install Podman Desktop GUI
brew install --cask podman-desktop

# Install Podman CLI
brew install podman

# Install Docker Compose compatibility
brew install podman-compose
```

### 2. Run Setup Script
```bash
./scripts/setup-podman.sh
```

### 3. Initialize and Start Podman
```bash
# Initialize Podman machine (VM for containers)
podman machine init

# Start the machine
podman machine start

# Verify installation
podman --version
podman machine list
```

## Usage Changes

### Docker Commands → Podman Commands
| Docker Command | Podman Equivalent | Notes |
|---------------|------------------|-------|
| `docker run` | `podman run` | Identical syntax |
| `docker build` | `podman build` | Identical syntax |
| `docker-compose up` | `podman-compose up` | Use podman-compose |
| `docker ps` | `podman ps` | Identical output |
| `docker images` | `podman images` | Identical output |

### Project-Specific Commands

#### Start Services
```bash
# Using podman-compose
podman-compose -f podman-compose.yml up -d

# Or using docker-compose syntax (with alias)
docker-compose up -d
```

#### Test APIs
```bash
# Use the new Podman-compatible test script
./test-podman-api.sh
```

#### Build Application
```bash
# Build Go application
go build -o bin/api-gateway ./cmd/api-gateway

# Or use Podman to build in container
podman build -t netorchestrator:latest .
```

## Docker Desktop Removal

After confirming Podman works correctly:

### 1. Stop Docker Desktop
- Quit Docker Desktop application
- Disable "Start Docker Desktop when you log in"

### 2. Uninstall Docker Desktop
```bash
# Uninstall Docker Desktop
brew uninstall --cask docker

# Or download the uninstaller from Docker's website
```

### 3. Clean Up Docker Files (Optional)
```bash
# Remove Docker data (optional - keeps your images/containers)
rm -rf ~/.docker
```

## Troubleshooting

### Common Issues

#### Podman Machine Not Starting
```bash
podman machine stop
podman machine rm
podman machine init --cpus 2 --memory 4096
podman machine start
```

#### Port Conflicts
```bash
# Check what's using the port
lsof -i :5432

# Stop conflicting services
brew services stop postgresql
```

#### Permission Issues
```bash
# Reset Podman
podman machine reset
podman machine init
```

### Verification Commands
```bash
# Check Podman status
podman machine list
podman system info

# Test container run
podman run hello-world

# Test compose
podman-compose --version
```

## Benefits After Migration

1. **Compliance**: ✅ Meets Cisco software policies
2. **Security**: 🔒 Rootless containers by default
3. **Performance**: ⚡ No background daemon
4. **Compatibility**: 🔄 Same commands and workflows
5. **Cost**: 💰 Free to use

## Support Resources

- [Podman Desktop Documentation](https://podman-desktop.io/docs)
- [Podman CLI Documentation](https://docs.podman.io)
- [Docker to Podman Migration](https://podman.io/getting-started/migration)

## Project Files Updated

- ✅ `podman-compose.yml` - Podman-optimized compose file
- ✅ `scripts/setup-podman.sh` - Automated setup script  
- ✅ `test-podman-api.sh` - Updated test script
- ✅ `PODMAN_MIGRATION.md` - This documentation