.PHONY: docker-compose-up
docker-compose-up: ## Starts up docker compose services
	@$(COMPOSE_CMD) up --detach

.PHONY: docker-compose-down
docker-compose-down: ## Stops docker compose services
	@$(COMPOSE_CMD) down

.PHONY: ping-postgres
ping-postgres: docker-compose-up ## Waits until postgres is ready to accept connections
	$(COMPOSE_CMD) exec -T -- postgres sh -c "until pg_isready; do sleep 1s; done"
