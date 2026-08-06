# Local Runbook

## Prerequisites

- Go 1.25+
- Docker Desktop or Docker Engine with Compose
- OpenSSL
- A reachable PostgreSQL/Supabase URL for `SUPABASE_DB_URL`

On Windows, use PowerShell. On macOS/Linux, use a POSIX shell such as `bash` or
`zsh`.

## Bootstrap

Windows PowerShell:

```powershell
Copy-Item .env.example .env
New-Item -ItemType Directory -Force secrets
openssl genrsa -out secrets/private.pem 4096
openssl rsa -in secrets/private.pem -pubout -out secrets/public.pem
[guid]::NewGuid().ToString("N") | Set-Content -NoNewline secrets/metrics_bearer_token
```

macOS/Linux:

```bash
cp .env.example .env
mkdir -p secrets
openssl genrsa -out secrets/private.pem 4096
openssl rsa -in secrets/private.pem -pubout -out secrets/public.pem
openssl rand -hex 32 > secrets/metrics_bearer_token
```

Update `.env` with the local database URL and CORS origins. For Docker Compose,
keep `METRICS_BEARER_TOKEN_FILE` unset: `docker-compose.yml` injects
`/run/secrets/metrics_bearer_token` into the API and Prometheus.

## Run API Directly

Use this when you only want the Go API without the full Docker stack.

Windows PowerShell:

```powershell
cd go
$env:JWT_PRIVATE_KEY_PATH = "../secrets/private.pem"
$env:JWT_PUBLIC_KEY_PATH = "../secrets/public.pem"
$env:METRICS_BEARER_TOKEN_FILE = "../secrets/metrics_bearer_token"
go run ./cmd/server
```

macOS/Linux:

```bash
cd go
export JWT_PRIVATE_KEY_PATH="../secrets/private.pem"
export JWT_PUBLIC_KEY_PATH="../secrets/public.pem"
export METRICS_BEARER_TOKEN_FILE="../secrets/metrics_bearer_token"
go run ./cmd/server
```

## Start The Stack

Windows PowerShell:

```powershell
docker-compose up -d --build
docker-compose ps
```

macOS/Linux:

```bash
docker compose up -d --build
docker compose ps
```

If your macOS/Linux setup only provides the legacy binary, replace
`docker compose` with `docker-compose`.

Services:

- API: http://localhost:8080
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000
- pprof (loopback only): http://127.0.0.1:6060/debug/pprof/

Grafana provisions the RNCP dashboard automatically from
`go/infra/grafana/dashboards/streampulse.json`. See
`docs/observability-rncp.md` for the dashboard panels, PromQL queries, alert
rules, and Grafana Cloud import guidance.

## Health And Metrics

Windows PowerShell:

```powershell
curl.exe http://localhost:8080/health
curl.exe -i http://localhost:8080/metrics
$token = Get-Content .\secrets\metrics_bearer_token
curl.exe -H "Authorization: Bearer $token" http://localhost:8080/metrics
```

macOS/Linux:

```bash
curl http://localhost:8080/health
curl -i http://localhost:8080/metrics
token="$(cat secrets/metrics_bearer_token)"
curl -H "Authorization: Bearer ${token}" http://localhost:8080/metrics
```

Expected result: `/health` returns `200`, `/metrics` without a bearer token
returns `401` when the token is configured, and `/metrics` with the token
returns Prometheus text output.

## Real Audio Smoke Test

Create a stream with a Diffuseur JWT, then keep a real-time source connected:

```bash
ffmpeg -re -stream_loop -1 -i sample.mp3 -codec copy -f mp3 - \
  | curl --no-buffer -X PUT \
      -H "Authorization: Bearer $BROADCASTER_TOKEN" \
      -H "Content-Type: audio/mpeg" \
      --data-binary @- \
      "http://localhost:8080/api/v1/streams/$STREAM_ID/audio"
```

Open the stream from Flutter Web (`GET /audio` + MediaSource) or download it
with an authenticated curl through the non-Web listener route:

```bash
curl --no-buffer \
  -H "Authorization: Bearer $LISTENER_TOKEN" \
  "http://localhost:8080/api/v1/streams/$STREAM_ID/listen" \
  --output received.mp3
```

The Flutter broadcaster uses ordered `POST /api/v1/streams/$STREAM_ID/push`
requests with `Content-Type: audio/webm;codecs=opus`; the continuous PUT example
above remains useful for a backend-only smoke test.

Stopping the stream, closing either client or stopping Compose must return the
audio gauges to zero. See `docs/streaming-lifecycle.md`.

## Tests And Coverage

Windows PowerShell:

```powershell
cd go
go test ./... -race -coverprofile=../coverage.out -covermode=atomic
go tool cover -func=../coverage.out
cd ..
```

macOS/Linux:

```bash
cd go
go test ./... -race -coverprofile=../coverage.out -covermode=atomic
go tool cover -func=../coverage.out
cd ..
```

On Windows, if `-race` fails because `gcc` is not installed, run the same command
without `-race` locally. CI still runs the race detector on Ubuntu.

CI reports total Go coverage on every run. To turn the future 80% gate on,
set `COVERAGE_MIN: "80"` in `.github/workflows/ci.yml` or run locally with:

Windows PowerShell:

```powershell
make test-ci COVERAGE_THRESHOLD=80
```

macOS/Linux:

```bash
make test-ci COVERAGE_THRESHOLD=80
```

## Stop And Clean

Windows PowerShell:

```powershell
docker-compose down
Remove-Item coverage.out -ErrorAction SilentlyContinue
```

macOS/Linux:

```bash
docker compose down
rm -f coverage.out
```
