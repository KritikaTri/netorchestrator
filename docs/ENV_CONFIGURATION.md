# Environment Configuration Guide

## Overview

NetOrchestrator now supports permanent configuration through `.env` files, eliminating the need to set environment variables manually each time.

## 🚀 Quick Setup

### 1. Create .env File

Create a `.env` file in the project root:

```bash
# NetOrchestrator Environment Configuration
OPENAI_API_KEY=your-openai-api-key-here
OPENAI_MODEL=gpt-4

# Optional: Database overrides
# DB_HOST=localhost
# DB_PORT=5432
# DB_NAME=netorchestrator

# Optional: Server configuration
# SERVER_PORT=8080
# LOG_LEVEL=info
```

### 2. Start Services

Simply run the startup script - no manual environment variables needed:

```bash
./start-all.sh
```

The system will automatically:
- Load the `.env` file
- Configure OpenAI integration
- Start all services with AI capabilities enabled

## 📋 Configuration Sources (Priority Order)

1. **Environment Variables** (highest priority)
2. **.env File** (medium priority)  
3. **config.yaml** (default values)
4. **Code Defaults** (fallback)

## 🔧 Configuration Options

### OpenAI Configuration

```bash
# .env file
OPENAI_API_KEY=sk-proj-...              # Your OpenAI API key
OPENAI_MODEL=gpt-4                      # Model to use (gpt-4, gpt-3.5-turbo)
```

Or in `config.yaml`:
```yaml
openai:
  api_key: ""           # Leave empty to load from env
  model: "gpt-4"
  temperature: 0.7
  max_tokens: 2000
  enabled: false        # Auto-enabled when API key provided
```

### Database Configuration

```bash
# .env file
DB_HOST=localhost
DB_PORT=5432
DB_NAME=netorchestrator
DB_USER=netorchestrator
DB_PASSWORD=password
```

### Server Configuration

```bash
# .env file
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
LOG_LEVEL=info
LOG_FORMAT=json
```

## 🔐 Security Best Practices

### .env File Security

1. **Never commit .env to Git**:
   ```bash
   echo ".env" >> .gitignore
   ```

2. **Set proper file permissions**:
   ```bash
   chmod 600 .env
   ```

3. **Use different .env files for environments**:
   - `.env.development`
   - `.env.production`  
   - `.env.testing`

### API Key Management

- Store API keys only in `.env` files or secure environment variables
- Use different API keys for different environments
- Monitor OpenAI usage and set billing limits
- Rotate API keys regularly

## 🧪 Testing Configuration

### Verify .env Loading

```bash
# Test that .env file is loaded
curl http://localhost:8080/health

# Check logs for OpenAI configuration
grep -i "openai" logs/api-gateway.log
```

### Test AI Functionality

```bash
# Test AI intelligence
curl -X POST http://localhost:8080/api/v1/intelligence/analyze/network \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"analysis_type": "performance", "metrics": {"cpu": 75}}'

# Test NLP
curl -X POST http://localhost:8080/api/v1/ai/provision \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"text": "Create a network with 3 routers"}'
```

## 📊 Configuration Status

The system will log configuration status at startup:

```
✅ OpenAI integration enabled (model: gpt-4, configured: true)
```

Or if not configured:
```
ℹ️  OpenAI not configured, using mock AI responses
ℹ️  To enable OpenAI: Create .env file with OPENAI_API_KEY
```

## 🔄 Configuration Changes

### Updating Configuration

1. **Modify .env file**:
   ```bash
   nano .env
   ```

2. **Restart services**:
   ```bash
   ./stop-all.sh
   ./start-all.sh
   ```

3. **Verify changes**:
   ```bash
   curl http://localhost:8080/health
   ```

### Environment Variable Precedence

Environment variables override `.env` file values:

```bash
# This will override the .env file value
export OPENAI_MODEL="gpt-3.5-turbo"
./start-all.sh
```

## 🎯 Benefits

### ✅ Permanent Configuration
- No need to set environment variables manually
- Configuration persists across restarts
- Easy to manage different environments

### ✅ Security
- API keys stored locally, not in code
- .env files can be excluded from version control
- Supports multiple configuration sources

### ✅ Flexibility  
- Override with environment variables when needed
- Support for multiple environments
- Backward compatible with existing setups

## 🚨 Important Notes

1. **Restart Required**: Configuration changes require service restart
2. **File Permissions**: Secure your .env file with `chmod 600 .env`  
3. **Git Exclusion**: Add `.env` to `.gitignore`
4. **API Limits**: Monitor OpenAI usage and costs
5. **Key Rotation**: Update API keys regularly for security
