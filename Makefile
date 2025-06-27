# ========================
# Project Configuration
# ========================
APP_NAME        := entain-server
COMPOSE_FILE    := build/docker-compose.yml
DOCKER_COMPOSE  := docker-compose -f $(COMPOSE_FILE)

# ========================
# Lifecycle Commands
# ========================

.PHONY: up
up:  ## Build and start all containers
	$(DOCKER_COMPOSE) up -d --build

.PHONY: down
down:  ## Stop and remove all containers
	$(DOCKER_COMPOSE) down

.PHONY: rebuild-app
rebuild-app:  ## Rebuild only the app container
	$(DOCKER_COMPOSE) build app

.PHONY: logs
logs:  ## Tail logs from the app container
	$(DOCKER_COMPOSE) logs -f app

# ========================
# Testing & Linting
# ========================

.PHONY: test-script
test-script:  ## Run functional test script
	bash ./scripts/test.sh

# ========================
# Cleanup
# ========================

.PHONY: clean
clean:  ## Remove generated binary
ifeq ($(OS),Windows_NT)
	@cmd /C "if exist $(APP_NAME).exe del /f $(APP_NAME).exe"
else
	rm -f $(APP_NAME)
endif

.PHONY: test
test:  ## Run local unit tests (if any)
	go test $(TEST_ARGS) ./test/...