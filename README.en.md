# StreamPulse

[Documentation française](README.md)

StreamPulse is a live-audio platform composed of a Go API, a Flutter client,
PostgreSQL/Supabase and a Prometheus/Grafana observability stack.

## Delivered capabilities

- Go 1.25 Clean Architecture API with RS256 JWT and User/Broadcaster/Admin RBAC.
- Real HTTP chunked audio ingestion and fan-out to concurrent listeners.
- Context propagation, sliding stream deadlines and deterministic shutdown.
- Metrics derived from audio bytes actually read and written, not UI events.
- Reproducible 10/100/500-listener benchmark, k6 scenario and isolated pprof.
- Multi-stage non-root container image.
- Production Compose stack behind Caddy with ACME HTTPS.
- Immutable Docker Hub image tags, release gates, health check and automatic rollback.

## Local start

Requirements: Go 1.25+, Flutter stable, OpenSSL and Docker with Compose.

```powershell
Copy-Item .env.example .env
New-Item -ItemType Directory -Force secrets
openssl genrsa -out secrets/private.pem 4096
openssl rsa -in secrets/private.pem -pubout -out secrets/public.pem
[guid]::NewGuid().ToString("N") | Set-Content -NoNewline secrets/metrics_bearer_token
docker compose up -d --build
```

Set `SUPABASE_DB_URL` and `CORS_ALLOWED_ORIGINS` first. The API, Prometheus,
Grafana and loopback-only pprof endpoints listen on ports 8080, 9090, 3000 and
6060 respectively.

## Audio protocol

1. Create a live stream with `POST /api/v1/streams`.
2. The owner sends a continuous authenticated body to
   `PUT /api/v1/streams/{id}/audio` with an `audio/*` content type.
3. Each listener opens authenticated
   `GET /api/v1/streams/{id}/listen`.
4. Stop the stream with `PUT /api/v1/streams/{id}/stop`.

The GET connection owns the business join/leave lifecycle. A disconnect,
stream stop, idle deadline or SIGTERM releases its Hub channel, database count
and metrics. Slow listener queues never block the broadcaster; dropped chunks
are counted.

```bash
ffmpeg -re -stream_loop -1 -i sample.mp3 -codec copy -f mp3 - \
  | curl --no-buffer -X PUT \
      -H "Authorization: Bearer $BROADCASTER_TOKEN" \
      -H "Content-Type: audio/mpeg" \
      --data-binary @- \
      "$API_BASE_URL/streams/$STREAM_ID/audio"
```

## Verification

```bash
cd go
go test ./...
go test ./... -race
cd ..
make load-bench
```

The CI coverage gate remains above 80%. The integration suite sends a binary
payload through the real HTTP ingestion/listen handlers, verifies byte-accurate
fan-out and asserts real Prometheus ingress/egress counters.

## Production

Copy `.env.production.example` to `.env.production`, create the JWT, metrics
and Grafana secret files, point the two DNS names to the host, then run:

```bash
IMAGE_REPOSITORY=dockerhub-user/streampulse-api \
IMAGE_TAG=<full-git-sha> \
./deploy/deploy.sh
```

Caddy obtains and renews certificates automatically. Only ports 80/443 are
published by the application stack; metrics, pprof and the API container port
remain private. `.github/workflows/deploy.yml` tests the release and publishes
the SHA-tagged image to Docker Hub with SBOM/provenance. On `main`, it then
calls the Render deploy hook. When `ENABLE_VPS_DEPLOY=true`, it may also deploy
to a standalone host and roll back if the public HTTPS health check fails.

## Detailed documentation

- [Local runbook](docs/runbook-local.md)
- [Streaming lifecycle](docs/streaming-lifecycle.md)
- [Observability](docs/observability-rncp.md)
- [Load, pprof, capacity and cost](docs/performance-couts.md)
- [HTTPS production deployment (French)](docs/deploiement-https.md)
- [Architecture decisions](go/docs/adr)

## Current scaling boundary

The Hub is process-local. Production therefore runs one API replica for the
audio data plane. Horizontal scaling requires stream-ID affinity or an
external media bus/SFU; adding replicas without either would split publishers
from listeners.
