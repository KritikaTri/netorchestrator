PROJECT?=netorchestrator
COMPOSE_FILE?=docker-compose.monitoring.yml

.PHONY: build up up-core up-api up-monitoring logs down ps

build:
	@docker compose -f $(COMPOSE_FILE) build netorchestrator

up-core:
	@docker compose -f $(COMPOSE_FILE) up -d postgres redis influxdb

up-api:
	@docker compose -f $(COMPOSE_FILE) up -d netorchestrator

up-monitoring:
	@docker compose -f $(COMPOSE_FILE) up -d prometheus grafana alertmanager node-exporter cadvisor jaeger

up: up-core build up-api up-monitoring

logs:
	@docker compose -f $(COMPOSE_FILE) logs -f netorchestrator

ps:
	@docker compose -f $(COMPOSE_FILE) ps

down:
	@docker compose -f $(COMPOSE_FILE) down -v
