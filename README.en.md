# StreamPulse

[Documentation française](README.md)

StreamPulse is a live-audio platform composed of a Go API, a Flutter client,
PostgreSQL/Supabase and a Prometheus/Grafana observability stack.

## Delivered capabilities

- Go 1.26 Clean Architecture API with RS256 JWT and User/Broadcaster/Admin RBAC.
- Real HTTP chunked audio ingestion and fan-out to concurrent listeners.
- Context propagation, sliding stream deadlines and deterministic shutdown.
- Metrics derived from audio bytes actually read and written, not UI events.
- Reproducible 10/100/500-listener benchmark, k6 scenario and isolated pprof.
- Multi-stage non-root container image.
- Immutable Docker Hub SHA/`latest` tags with release gates and provenance.
- Automatic Render deployment through a secret hook and managed HTTPS edge.

## Local start

Requirements: Go 1.26+, Flutter stable, OpenSSL and Docker with Compose.

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
2. Flutter Web captures WebM/Opus and sends ordered `MediaRecorder` blobs to
   authenticated `POST /api/v1/streams/{id}/push`. Continuous sources such as
   FFmpeg can still use `PUT /api/v1/streams/{id}/audio`.
3. Flutter consumes public `GET /api/v1/streams/{id}/audio` incrementally through
   `MediaSource`; short authenticated `/listen` and `/leave` requests record the
   business history. Authenticated streaming `GET /listen` remains available to
   non-Web clients.
4. Stop the stream with `PUT /api/v1/streams/{id}/stop`.

The Web player bounds its live buffer, jumps back to the live edge after a long
pause, and publishes play/pause/stop through the browser Media Session API.
A disconnect, stream stop, idle deadline or SIGTERM releases Hub resources.
Slow listener queues never block the broadcaster; dropped chunks are counted.

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

`.github/workflows/deploy.yml` tests the release, publishes the Go image to
Docker Hub with immutable SHA and `latest` tags, SBOM and provenance, then
calls the Render deploy hook after a successful push to `main`. Render runs the
container and provides the managed HTTPS edge.

Configure `DOCKERHUB_USERNAME`, `DOCKERHUB_REPOSITORY`, `DOCKERHUB_TOKEN` and
`RENDER_DEPLOY_HOOK_URL` in GitHub. Configure database, JWT Secret Files, CORS
and metrics credentials in Render. See the
[production guide](docs/deploiement-https.md).

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
