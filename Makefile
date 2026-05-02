.PHONY: dev build test lint vet clean docker-build migrate help

.PHONY: install-tools

## ── Development ──────────────────────────────────────────────────────────────

# Start the API server locally (loads .env automatically via godotenv)
dev:
	cd go && go run ./cmd/server

# Build the production binary
build:
	cd go && CGO_ENABLED=0 go build -ldflags='-w -s' -o ../bin/streampulse ./cmd/server

# ── Tests ─────────────────────────────────────────────────────────────────────

# Run all tests with race detector + coverage report
test:
	cd go && go test ./... -race -coverprofile=../coverage.out -covermode=atomic
	go tool cover -func=coverage.out | tail -1

# Run tests and enforce 80% coverage threshold (CI gate)
test-ci:
	cd go && go test ./... -race -coverprofile=../coverage.out -covermode=atomic
	go tool cover -func=coverage.out | awk '/total/{if ($$3+0 < 80) { print "Coverage below 80%: " $$3; exit 1 }}'

# ── Quality ───────────────────────────────────────────────────────────────────

# Static analysis
vet:
	cd go && go vet ./...

# Linter — installe golangci-lint automatiquement si absent
lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@export PATH=$$PATH:$$(go env GOPATH)/bin && cd go && golangci-lint run ./...

# Install development tools used by the project (goimports, golangci-lint)
install-tools: ## Install goimports and golangci-lint into GOPATH/bin
	@echo "Installing goimports and golangci-lint..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Installed: $$(go env GOPATH)/bin/goimports $$(go env GOPATH)/bin/golangci-lint"

# ── Docker ────────────────────────────────────────────────────────────────────

docker-build:
	docker build -t streampulse-api:local ./go

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ── Database ──────────────────────────────────────────────────────────────────

# Apply all migrations in order (charge .env automatiquement)
migrate:
	@echo "Applying migrations..."
	@export $$(grep -v '^#' .env | xargs) && \
		psql $$SUPABASE_DB_URL -f migrations/001_init.sql && \
		psql $$SUPABASE_DB_URL -f migrations/002_rls.sql && \
		psql $$SUPABASE_DB_URL -f migrations/003_listen_history.sql && \
		psql $$SUPABASE_DB_URL -f migrations/004_alter_users_and_playlist_tracks.sql && \
		psql $$SUPABASE_DB_URL -f migrations/005_playlist_track_count.sql
	@echo "Done."

# ── Flutter ───────────────────────────────────────────────────────────────────

flutter-run:
	cd flutter && flutter run

flutter-test:
	cd flutter && flutter test

flutter-build-apk:
	cd flutter && flutter build apk --release

flutter-build-web:
	cd flutter && flutter build web --release

# ── Utilities ─────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ coverage.out

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
