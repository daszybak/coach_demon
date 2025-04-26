BIN := coach_demon
PKG := ./cmd/coach_demon
REPORT_ROOT ?= tests/reports
JOURNEY_DIR := $(REPORT_ROOT)/journey
INT_DIR := $(REPORT_ROOT)/integration

# --- Build & Run ---------------------------------------------------
build:
	go build -o $(BIN) $(PKG)

run: build
	./$(BIN)

# --- TESTS: Host-side commands -------------------------------------
test-integration: docker-build
	docker compose up -d mongo fetcher  # Start ALL needed dependencies
	docker compose run --rm tests make test-integration-local
	docker compose down  # Clean up everything when done

test-journey: docker-build
	docker compose up -d mongo fetcher  # Start ALL needed dependencies
	docker compose run --rm tests make test-journey-local
	docker compose down  # Clean up everything when done

test-all: docker-build
	docker compose up -d mongo fetcher  # Start ALL needed dependencies
	docker compose run --rm tests make test-all-local
	docker compose down  # Clean up everything when done

# --- TESTS: Commands running inside the "tests" container ----------
test-integration-local:
	@mkdir -p $(INT_DIR)
	gotestsum --format pkgname \
		--junitfile $(INT_DIR)/results.xml \
		-- -tags=integration ./tests/integration/... \
	| tee $(INT_DIR)/run.out
	@echo "✔ integration JUnit → $(INT_DIR)/results.xml"

test-journey-local:
	@mkdir -p $(JOURNEY_DIR)
	gotestsum --format pkgname \
		--junitfile $(JOURNEY_DIR)/results.xml \
		-- -tags=journey ./tests/journey/... \
	| tee $(JOURNEY_DIR)/run.out
	@echo "✔ journey JUnit → $(JOURNEY_DIR)/results.xml"

test-all-local:
	@echo "Running tests in container environment"
	@$(MAKE) test-integration-local
	@$(MAKE) test-journey-local

# --- Docker helpers ------------------------------------------------
docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-up-fetcher:
	@echo "Running in container - Docker operations skipped"
	docker compose up -d fetcher

docker-down-fetcher:
	docker compose stop fetcher
	docker compose rm -f fetcher

docker-build:
	docker compose build

.PHONY: build run test-journey test-integration test-all \
        test-integration-local test-journey-local test-all-local \
        docker-up docker-down docker-up-fetcher docker-down-fetcher
