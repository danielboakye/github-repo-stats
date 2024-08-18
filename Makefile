.PHONY: get-leaderboard get-commits

# Set the default port if not provided
PORT ?= 9000

APP_NAME := github-repo-stats

setup/db:
	@docker stop github-repo-stats-db; docker rm github-repo-stats-db; true;
	@docker run \
		--detach \
		--name=github-repo-stats-db \
		--env="POSTGRES_USER=user" \
		--env="POSTGRES_PASSWORD=password" \
		--env="POSTGRES_DB=github_repo_stats_db" \
		--publish 5432:5432 \
		--health-cmd="pg_isready -U user || exit 1" \
		--health-interval=10s \
		--health-timeout=5s \
		--health-retries=5 \
		postgres:latest
	@echo "Waiting for database to be ready..."
	@until [ "$$(docker inspect --format='{{json .State.Health.Status}}' github-repo-stats-db)" == "\"healthy\"" ]; do \
		sleep 2; \
		echo "Waiting..."; \
	done
	@echo "Database is ready!"
	@docker cp scripts/ddl/github_repo_stats_ddl.sql github-repo-stats-db:/github_repo_stats_ddl.sql
	@docker exec -i github-repo-stats-db psql -U user -d github_repo_stats_db -f /github_repo_stats_ddl.sql
	@echo "db is ready!"

build:
	@echo "Building $(APP_NAME)..."
	@go build

start: build
	@echo "Running $(APP_NAME) on port $(PORT)..."
	POSTGRES_URL="postgres://user:password@localhost:5432/github_repo_stats_db?sslmode=disable" ./$(APP_NAME) -port=$(PORT) -since=$(SINCE)


# Default targets if not specified
REPO ?= chromium/chromium
LIMIT ?= 5

get-leaderboard:
	@echo "Retrieving top $(LIMIT) leaderboard on port $(PORT)..."
	@./scripts/get-leaderboard.sh $(LIMIT) $(PORT)

get-commits:
	@echo "Retrieving $(LIMIT) commits for <$(REPO)> on port $(PORT)..."
	@./scripts/get-commits.sh $(REPO) $(LIMIT) $(PORT)
 
