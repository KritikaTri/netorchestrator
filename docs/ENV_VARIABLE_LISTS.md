# Environment Variables - Consistency Report

**Date:** 2025-10-30

---

## 1) Variables that exist in `.env` but NOT referenced in compose

✅ **NONE** - All variables in `.env` are referenced in `docker-compose.monitoring.yml`

**Variables in `.env`:**
- `DATABASE_PASSWORD` - ✅ Referenced as `${DATABASE_PASSWORD}`
- `JWT_SECRET` - ✅ Referenced as `${JWT_SECRET}`
- `INFLUX_TOKEN` - ⚠️ **Not directly referenced** (hardcoded value exists but doesn't use this var)

**Note:** While `INFLUX_TOKEN` exists in `.env`, the compose file hardcodes `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=netorchestrator-super-secret-token` instead of using `${INFLUX_TOKEN}`. The values match but the reference is missing.

---

## 2) Variables referenced in compose but MISSING in `.env`

### Database Variables (Backend expects these)
- `DATABASE_HOST` - ❌ Missing (hardcoded: `host.docker.internal`)
- `DATABASE_PORT` - ❌ Missing (hardcoded: `5433`)
- `DATABASE_USER` - ❌ Missing (hardcoded: `netorch_admin`)
- `DATABASE_DBNAME` - ❌ Missing (hardcoded: `netorch_mvp`)
- `DATABASE_PASSWORD` - ✅ Present in `.env`
- `DATABASE_SSLMODE` - ❌ Missing (default used: "disable")

### Redis Variables (Backend expects these)
- `REDIS_HOST` - ❌ Missing (hardcoded: `host.docker.internal`)
- `REDIS_PORT` - ❌ Missing (hardcoded: `6379`)
- `REDIS_DB` - ❌ Missing (hardcoded: `0`)
- `REDIS_PASSWORD` - ❌ Missing (default: "")

### NATS Variables (Backend expects these)
- `NATS_URL` - ❌ Missing (hardcoded: `nats://nats:4222`)

### InfluxDB Variables (Backend expects these)
- `INFLUXDB_URL` - ❌ Missing (default: "http://localhost:8086")
- `INFLUXDB_TOKEN` - ⚠️ **Has `INFLUX_TOKEN` in .env** (naming mismatch)
- `INFLUXDB_ORG` - ❌ Missing (hardcoded: `netorchestrator`)
- `INFLUXDB_BUCKET` - ❌ Missing (hardcoded: `metrics`)
- `INFLUXDB_INIT_PASSWORD` - ❌ Missing (hardcoded: `netorchestrator123`)
- `INFLUXDB_INIT_USERNAME` - ❌ Missing (hardcoded: `admin`)

### Grafana Variables
- `GRAFANA_ADMIN_PASSWORD` - ❌ Missing (hardcoded: `netorchestrator123`)
- `GF_SECURITY_ADMIN_PASSWORD` - ❌ Missing (hardcoded: `netorchestrator123`)

### PostgreSQL Service Variables (In compose but not referenced)
- `POSTGRES_PASSWORD` - ❌ Missing (should use `${DATABASE_PASSWORD}` but hardcoded: `SecurePass2025`)
- `POSTGRES_USER` - ❌ Missing (hardcoded: `netorch_admin`)
- `POSTGRES_DB` - ❌ Missing (hardcoded: `netorch_mvp`)

### Server Variables (In compose but could be env vars)
- `SERVER_HOST` - ❌ Missing (hardcoded: `0.0.0.0`)
- `SERVER_PORT` - ❌ Missing (hardcoded: `8080`)

---

## 3) Variables referenced in backend but MISSING in both `.env` and compose

### Database Config (Has defaults in `config.go`)
- `DATABASE_HOST` - ❌ Missing (default: "localhost")
- `DATABASE_PORT` - ❌ Missing (default: 5432)
- `DATABASE_USER` - ❌ Missing (default: "netorchestrator")
- `DATABASE_DBNAME` - ❌ Missing (default: "netorchestrator")
- `DATABASE_SSLMODE` - ❌ Missing (default: "disable")

### Redis Config (Has defaults)
- `REDIS_HOST` - ❌ Missing (default: "localhost")
- `REDIS_PORT` - ❌ Missing (default: 6379)
- `REDIS_PASSWORD` - ❌ Missing (default: "")
- `REDIS_DB` - ❌ Missing (default: 0)

### InfluxDB Config (Has defaults)
- `INFLUXDB_URL` - ❌ Missing (default: "http://localhost:8086")
- `INFLUXDB_ORG` - ❌ Missing (default: "netorchestrator")
- `INFLUXDB_BUCKET` - ❌ Missing (default: "network_metrics")
- `INFLUXDB_TOKEN` - ⚠️ **Has `INFLUX_TOKEN` in .env** (naming mismatch, backend expects `INFLUXDB_TOKEN`)

### NATS Config (Has default)
- `NATS_URL` - ❌ Missing (default: "nats://localhost:4222")

### JWT Config (Has defaults)
- `JWT_SECRET` - ✅ Present in `.env`
- `JWT_EXPIRATION` - ❌ Missing (default: 3600)

### Security Config (Has defaults)
- `SECURITY_JWT_SECRET` - ⚠️ **Compose uses `${JWT_SECRET}`** (works via viper)
- `SECURITY_RATE_LIMIT_RPS` - ❌ Missing (default: 100)
- `SECURITY_RATE_LIMIT_BURST` - ❌ Missing (default: 10)
- `SECURITY_ENABLE_AUDIT_LOGGING` - ❌ Missing (default: true)
- `SECURITY_ENCRYPTION_KEY` - ❌ Missing (default: "netorchestrator-encryption-key-32-byte")
- `SECURITY_ENABLE_CORS` - ❌ Missing (default: true)
- `SECURITY_ALLOWED_ORIGINS` - ❌ Missing (default: ["*"])
- `SECURITY_ALLOWED_METHODS` - ❌ Missing (default: ["GET", "POST", "PUT", "DELETE", "OPTIONS"])
- `SECURITY_ALLOWED_HEADERS` - ❌ Missing (default: ["Origin", "Content-Type", "Accept", "Authorization", "X-API-Key"])

### Server Config (Has defaults)
- `SERVER_HOST` - ❌ Missing (default: "0.0.0.0") - Set in compose but not as env var
- `SERVER_PORT` - ❌ Missing (default: "8080") - Set in compose but not as env var
- `SERVER_READ_TIMEOUT` - ❌ Missing (default: 30)
- `SERVER_WRITE_TIMEOUT` - ❌ Missing (default: 30)
- `SERVER_IDLE_TIMEOUT` - ❌ Missing (default: 120)

### Logging Config (Has defaults)
- `LOGGING_LEVEL` - ❌ Missing (default: "info")
- `LOGGING_FORMAT` - ❌ Missing (default: "json")

### Services Config (Has defaults)
- `SERVICES_NETWORK_SERVICE_HOST` - ❌ Missing (default: "localhost")
- `SERVICES_NETWORK_SERVICE_PORT` - ❌ Missing (default: 8081)
- `SERVICES_MONITORING_SERVICE_HOST` - ❌ Missing (default: "localhost")
- `SERVICES_MONITORING_SERVICE_PORT` - ❌ Missing (default: 8082)
- `SERVICES_ORCHESTRATION_SERVICE_HOST` - ❌ Missing (default: "localhost")
- `SERVICES_ORCHESTRATION_SERVICE_PORT` - ❌ Missing (default: 8083)
- `SERVICES_VALIDATION_SERVICE_HOST` - ❌ Missing (default: "localhost")
- `SERVICES_VALIDATION_SERVICE_PORT` - ❌ Missing (default: 8084)

### Kafka Config (Has defaults, but not used)
- `KAFKA_BROKERS` - ❌ Missing (default: ["localhost:9092"])
- `KAFKA_GROUP_ID` - ❌ Missing (default: "netorchestrator")

---

## 📊 Summary Counts

### Category 1: Variables in `.env` but not in compose
- **Count:** 0 (all referenced) ✅

### Category 2: Variables in compose but missing in `.env`
- **Count:** 15+ variables
- **Critical Missing:** 10 variables (Database, Redis, NATS, InfluxDB core config)
- **Service Missing:** 5+ variables (Grafana, PostgreSQL service vars)

### Category 3: Variables in backend but missing in both
- **Count:** 30+ variables
- **Note:** Most have defaults in `config.go`, so not critical for basic operation
- **Critical (should be in env):** Database, Redis, NATS, InfluxDB, JWT, Security settings

---

## ⚠️ Critical Issues

1. **PostgreSQL Password:** Hardcoded in compose instead of using `${DATABASE_PASSWORD}`
2. **InfluxDB Token:** Naming mismatch (`INFLUX_TOKEN` in .env vs `INFLUXDB_TOKEN` in backend)
3. **Many variables hardcoded:** Should use env vars for flexibility
4. **Missing critical vars:** Database host/port/user, Redis host/port, NATS URL

---

## ✅ Recommendations

1. **Add to `.env`:** All database, Redis, NATS, InfluxDB variables
2. **Update compose:** Replace hardcoded values with `${VAR}` references
3. **Fix naming:** Standardize `INFLUX_TOKEN` vs `INFLUXDB_TOKEN`
4. **Add Grafana password:** To `.env` for security

