.PHONY: dev build test lint vet clean docker-build migrate flutter-dev flutter-run flutter-test flutter-build-apk flutter-build-web load-bench load-k6 profile-cpu help

COVERAGE_THRESHOLD ?= 80

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

# Run tests and enforce 80% coverage threshold (CI gate).
# Mock packages (test helpers) are excluded from the coverage measurement.
test-ci:
	cd go && go test $$(go list ./... | grep -v '/mock$$') -coverprofile=../coverage.out -covermode=atomic
	go tool cover -func=coverage.out | awk -v min="$(COVERAGE_THRESHOLD)" '/total/{if ($$3+0 < min) { print "Coverage below " min "%: " $$3; exit 1 } else { print "Coverage OK: " $$3 " >= " min "%" }}'

# Reproducible in-process fan-out benchmark at 10, 100 and 500 listeners.
load-bench:
	cd go && go test ./internal/infrastructure/streaming -run '^$$' -bench '^BenchmarkHubBroadcast$$' -benchmem -benchtime=2s

# End-to-end public audio test; requires STREAM_ID plus a live publisher.
# Set LISTENER_TOKEN to exercise the authenticated /listen route instead.
load-k6:
	cd go && k6 run -e LISTENERS=$${LISTENERS:-10} -e BASE_URL=$${BASE_URL} -e STREAM_ID=$${STREAM_ID} -e LISTENER_TOKEN=$${LISTENER_TOKEN} loadtest/stream.js

# Capture CPU, heap and goroutine profiles from a local PPROF_ENABLED API.
profile-cpu:
	cd go && PPROF_BASE_URL=$${PPROF_BASE_URL:-http://127.0.0.1:6060} sh loadtest/capture-pprof.sh

# ── Quality ───────────────────────────────────────────────────────────────────

# Static analysis
vet:
	cd go && go vet ./...

# Linter — installe golangci-lint automatiquement si absent
lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@export PATH=$$PATH:$$(go env GOPATH)/bin && cd go && golangci-lint run ./...

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
		psql $$SUPABASE_DB_URL -f migrations/005_playlist_track_count.sql && \
		psql $$SUPABASE_DB_URL -f migrations/006_user_suspend.sql && \
		psql $$SUPABASE_DB_URL -f migrations/007_favorites.sql && \
		psql $$SUPABASE_DB_URL -f migrations/008_listen_history_stream.sql && \
		psql $$SUPABASE_DB_URL -f migrations/009_listen_history_events.sql && \
		psql $$SUPABASE_DB_URL -f migrations/010_reusable_stream_sessions.sql
	@echo "Done."

# ── Flutter ───────────────────────────────────────────────────────────────────

# Read API_BASE_URL and FLUTTER_WEB_PORT from .env, pass to flutter via --dart-define
flutter-dev:
	@export $$(grep -v '^#' .env | grep -E '^(API_BASE_URL|FLUTTER_WEB_PORT)=' | xargs) && \
		cd flutter && flutter run -d chrome \
			--web-port=$${FLUTTER_WEB_PORT:-3001} \
			--dart-define=API_BASE_URL=$${API_BASE_URL:-http://localhost:8080/api/v1}

flutter-run:
	cd flutter && flutter run

flutter-test:
	cd flutter && flutter test

flutter-build-apk:
	cd flutter && flutter build apk --release

flutter-build-web:
	@export $$(grep -v '^#' .env | grep -E '^API_BASE_URL=' | xargs) && \
		cd flutter && flutter build web --release \
			--dart-define=API_BASE_URL=$${API_BASE_URL:-http://localhost:8080/api/v1}

# ── Utilities ─────────────────────────────────────────────────────────────────

clean:
	rm -rf bin/ coverage.out

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
