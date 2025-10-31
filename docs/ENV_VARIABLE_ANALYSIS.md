# Environment Variables Analysis

**Date:** 2025-10-30  
**Analysis:** Complete environment variable consistency check

---

## 📋 Variables Found

### 1. Variables in `.env` file:
- `DATABASE_PASSWORD`
- `JWT_SECRET`
- `INFLUX_TOKEN`

### 2. Variables in `ENV_EXAMPLE`:
- `DATABASE_PASSWORD`
- `JWT_SECRET`
- `INFLUX_TOKEN`

### 3. Variables Referenced in `docker-compose.monitoring.yml`:
**From .env file (${VAR}):**
- `${DATABASE_PASSWORD}` - ✅ Referenced
- `${JWT_SECRET}` - ✅ Referenced

**Hardcoded in compose (should use env vars):**
- `DATABASE_HOST=host.docker.internal` - ❌ Should be env var
- `DATABASE_PORT=5433` - ❌ Should be env var
- `DATABASE_DBNAME=netorch_mvp` - ❌ Should be env var
- `DATABASE_USER=netorch_admin` - ❌ Should be env var
- `REDIS_HOST=host.docker.internal` - ❌ Should be env var
- `REDIS_PORT=6379` - ❌ Should be env var
- `REDIS_DB=0` - ❌ Should be env var
- `NATS_URL=nats://nats:4222` - ❌ Should be env var

**PostgreSQL service (hardcoded):**
- `POSTGRES_DB=netorch_mvp` - ❌ Should reference env var
- `POSTGRES_USER=netorch_admin` - ❌ Should reference env var
- `POSTGRES_PASSWORD=SecurePass2025` - ❌ Should reference env var (should use `${DATABASE_PASSWORD}`)

**InfluxDB service (hardcoded):**
- `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=netorchestrator-super-secret-token` - ❌ Should use `${INFLUX_TOKEN}`
- `DOCKER_INFLUXDB_INIT_PASSWORD=netorchestrator123` - ❌ Should be env var
- `DOCKER_INFLUXDB_INIT_USERNAME=admin` - ❌ Should be env var
- `DOCKER_INFLUXDB_INIT_ORG=netorchestrator` - ❌ Should be env var
- `DOCKER_INFLUXDB_INIT_BUCKET=metrics` - ❌ Should be env var

**Grafana service (hardcoded):**
- `GF_SECURITY_ADMIN_PASSWORD=netorchestrator123` - ❌ Should be env var

**Prometheus:**
- No credentials needed (no env vars)

### 4. Variables Expected by Backend Config (`internal/config/config.go`):

**Viper uses `AutomaticEnv()` with key replacer: `.` → `_`**

**Database Config:**
- `DATABASE_HOST` (from `database.host`)
- `DATABASE_PORT` (from `database.port`)
- `DATABASE_USER` (from `database.user`)
- `DATABASE_PASSWORD` (from `database.password`)
- `DATABASE_DBNAME` (from `database.dbname`)
- `DATABASE_SSLMODE` (from `database.sslmode`)

**Redis Config:**
- `REDIS_HOST` (from `redis.host`)
- `REDIS_PORT` (from `redis.port`)
- `REDIS_PASSWORD` (from `redis.password`)
- `REDIS_DB` (from `redis.db`)

**InfluxDB Config:**
- `INFLUXDB_URL` (from `influxdb.url`)
- `INFLUXDB_TOKEN` (from `influxdb.token`)
- `INFLUXDB_ORG` (from `influxdb.org`)
- `INFLUXDB_BUCKET` (from `influxdb.bucket`)

**NATS Config:**
- `NATS_URL` (from `nats.url`) - ✅ Referenced in compose

**JWT Config:**
- `JWT_SECRET` (from `jwt.secret`)
- `JWT_EXPIRATION` (from `jwt.expiration`)

**Security Config:**
- `SECURITY_JWT_SECRET` (from `security.jwt_secret`) - ✅ Referenced in compose
- `SECURITY_RATE_LIMIT_RPS` (from `security.rate_limit_rps`)
- `SECURITY_RATE_LIMIT_BURST` (from `security.rate_limit_burst`)
- `SECURITY_ENABLE_AUDIT_LOGGING` (from `security.enable_audit_logging`)
- `SECURITY_ENCRYPTION_KEY` (from `security.encryption_key`)

**Server Config:**
- `SERVER_HOST` (from `server.host`) - ✅ Set in compose
- `SERVER_PORT` (from `server.port`) - ✅ Set in compose

**Logging Config:**
- `LOGGING_LEVEL` (from `logging.level`)
- `LOGGING_FORMAT` (from `logging.format`)

---

## 📊 Analysis Results

### 1. Variables in `.env` but NOT referenced in compose:
**NONE** ✅
- All variables in `.env` are referenced

### 2. Variables referenced in compose but MISSING in `.env`:

**Critical Missing:**
- `DATABASE_HOST` - Expected by backend, hardcoded in compose
- `DATABASE_PORT` - Expected by backend, hardcoded in compose
- `DATABASE_USER` - Expected by backend, hardcoded in compose
- `DATABASE_DBNAME` - Expected by backend, hardcoded in compose
- `REDIS_HOST` - Expected by backend, hardcoded in compose
- `REDIS_PORT` - Expected by backend, hardcoded in compose
- `REDIS_DB` - Expected by backend, hardcoded in compose
- `NATS_URL` - Expected by backend, hardcoded in compose
- `REDIS_PASSWORD` - Expected by backend, not in compose
- `DATABASE_SSLMODE` - Expected by backend, not in compose

**InfluxDB Missing:**
- `INFLUXDB_URL` - Expected by backend, hardcoded in compose
- `INFLUXDB_ORG` - Expected by backend, hardcoded in compose
- `INFLUXDB_BUCKET` - Expected by backend, hardcoded in compose
- `INFLUXDB_INIT_PASSWORD` - Hardcoded in compose, should be env var
- `INFLUXDB_INIT_USERNAME` - Hardcoded in compose, should be env var

**Grafana Missing:**
- `GRAFANA_ADMIN_PASSWORD` - Hardcoded in compose, should be env var

**Note:** `INFLUX_TOKEN` exists in `.env` but compose uses `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN` (hardcoded value, doesn't match)

### 3. Variables referenced in backend but MISSING in both `.env` and compose:

**Database (all have defaults in config.go):**
- `DATABASE_HOST` - Has default: "localhost"
- `DATABASE_PORT` - Has default: 5432
- `DATABASE_USER` - Has default: "netorchestrator"
- `DATABASE_DBNAME` - Has default: "netorchestrator"
- `DATABASE_SSLMODE` - Has default: "disable"

**Redis (all have defaults):**
- `REDIS_HOST` - Has default: "localhost"
- `REDIS_PORT` - Has default: 6379
- `REDIS_PASSWORD` - Has default: ""
- `REDIS_DB` - Has default: 0

**InfluxDB (all have defaults):**
- `INFLUXDB_URL` - Has default: "http://localhost:8086"
- `INFLUXDB_ORG` - Has default: "netorchestrator"
- `INFLUXDB_BUCKET` - Has default: "network_metrics"

**NATS (has default):**
- `NATS_URL` - Has default: "nats://localhost:4222"

**JWT (has defaults):**
- `JWT_EXPIRATION` - Has default: 3600

**Security (has defaults):**
- `SECURITY_RATE_LIMIT_RPS` - Has default: 100
- `SECURITY_RATE_LIMIT_BURST` - Has default: 10
- `SECURITY_ENABLE_AUDIT_LOGGING` - Has default: true
- `SECURITY_ENCRYPTION_KEY` - Has default: "netorchestrator-encryption-key-32-byte"

**Logging (has defaults):**
- `LOGGING_LEVEL` - Has default: "info"
- `LOGGING_FORMAT` - Has default: "json"

---

## ⚠️ Issues Found

### Issue 1: PostgreSQL Password Mismatch
- **Compose hardcodes:** `POSTGRES_PASSWORD=SecurePass2025`
- **.env has:** `DATABASE_PASSWORD=SecurePass2025`
- **Should:** Compose should use `${DATABASE_PASSWORD}` instead of hardcoding

### Issue 2: InfluxDB Token Mismatch
- **Compose hardcodes:** `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=netorchestrator-super-secret-token`
- **.env has:** `INFLUX_TOKEN=netorchestrator-super-secret-token`
- **Values match but:** Compose doesn't reference `${INFLUX_TOKEN}`

### Issue 3: Missing Environment Variables in .env
- Many variables are hardcoded in compose instead of using `.env`
- Backend expects these but they're not in `.env` (though they have defaults)

### Issue 4: JWT Secret Inconsistency
- **Compose uses:** `SECURITY_JWT_SECRET=${JWT_SECRET}` ✅ Correct
- **Backend expects:** `SECURITY_JWT_SECRET` or `JWT_SECRET`
- **.env has:** `JWT_SECRET` ✅ Works (viper maps both)

---

## ✅ Recommendations

### High Priority
1. **Update `docker-compose.monitoring.yml`** to use env vars instead of hardcoding:
   - `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_DBNAME`
   - `REDIS_HOST`, `REDIS_PORT`, `REDIS_DB`
   - `NATS_URL`
   - `INFLUXDB_*` variables
   - `GRAFANA_ADMIN_PASSWORD`

2. **Update `.env`** to include:
   - `DATABASE_HOST`
   - `DATABASE_PORT`
   - `DATABASE_USER`
   - `DATABASE_DBNAME`
   - `REDIS_HOST`
   - `REDIS_PORT`
   - `REDIS_DB`
   - `NATS_URL`
   - `INFLUXDB_URL`
   - `INFLUXDB_ORG`
   - `INFLUXDB_BUCKET`
   - `GRAFANA_ADMIN_PASSWORD`

3. **Fix PostgreSQL password reference:**
   - Change `POSTGRES_PASSWORD=SecurePass2025` to `POSTGRES_PASSWORD=${DATABASE_PASSWORD}`

4. **Fix InfluxDB token reference:**
   - Change `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=netorchestrator-super-secret-token` to `DOCKER_INFLUXDB_INIT_ADMIN_TOKEN=${INFLUX_TOKEN}`

---

## 📝 Summary Lists

### 1. Variables in `.env` but NOT referenced in compose:
**NONE** ✅

All variables in `.env` are referenced in compose (though some values are hardcoded instead of using env vars).

### 2. Variables referenced in compose but MISSING in `.env`:

**Critical (Backend expects these):**
- `DATABASE_HOST`
- `DATABASE_PORT`
- `DATABASE_USER`
- `DATABASE_DBNAME`
- `REDIS_HOST`
- `REDIS_PORT`
- `REDIS_DB`
- `NATS_URL`
- `REDIS_PASSWORD` (optional, default: "")
- `DATABASE_SSLMODE` (optional, default: "disable")

**InfluxDB:**
- `INFLUXDB_URL`
- `INFLUXDB_ORG`
- `INFLUXDB_BUCKET`
- `INFLUXDB_INIT_PASSWORD`
- `INFLUXDB_INIT_USERNAME`

**Grafana:**
- `GRAFANA_ADMIN_PASSWORD`

### 3. Variables referenced in backend but MISSING in both `.env` and compose:

**All have defaults in config.go, so not critical:**
- `DATABASE_HOST` (default: "localhost")
- `DATABASE_PORT` (default: 5432)
- `DATABASE_USER` (default: "netorchestrator")
- `DATABASE_DBNAME` (default: "netorchestrator")
- `DATABASE_SSLMODE` (default: "disable")
- `REDIS_HOST` (default: "localhost")
- `REDIS_PORT` (default: 6379)
- `REDIS_PASSWORD` (default: "")
- `REDIS_DB` (default: 0)
- `INFLUXDB_URL` (default: "http://localhost:8086")
- `INFLUXDB_ORG` (default: "netorchestrator")
- `INFLUXDB_BUCKET` (default: "network_metrics")
- `NATS_URL` (default: "nats://localhost:4222")
- `JWT_EXPIRATION` (default: 3600)
- `SECURITY_RATE_LIMIT_RPS` (default: 100)
- `SECURITY_RATE_LIMIT_BURST` (default: 10)
- `SECURITY_ENABLE_AUDIT_LOGGING` (default: true)
- `SECURITY_ENCRYPTION_KEY` (default: "netorchestrator-encryption-key-32-byte")
- `LOGGING_LEVEL` (default: "info")
- `LOGGING_FORMAT` (default: "json")

---

## 🔧 Quick Fix Script

To fix all issues, create an updated `.env` file with all required variables and update `docker-compose.monitoring.yml` to reference them instead of hardcoding values.

