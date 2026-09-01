# Observability RNCP - Grafana

## Faut-il le faire directement sur Grafana Cloud ?

Non. Le dashboard doit rester versionne dans le depot pour etre rejouable,
auditable et presentable en soutenance. Grafana Cloud peut ensuite servir
d'environnement de demonstration :

1. En local, Docker Compose provisionne automatiquement le datasource
   Prometheus et le dashboard depuis `go/infra/grafana`.
2. Dans Grafana Cloud, importer le JSON
   `go/infra/grafana/dashboards/streampulse.json`.
3. Creer un datasource Prometheus compatible avec les metriques StreamPulse.
4. Reprendre les regles de `go/infra/prometheus/alerts.yml` comme base pour
   les alert rules Grafana Cloud.

Cette approche evite les clics non traces dans l'interface Cloud et permet de
montrer au jury que l'observabilite fait partie du projet.

## Collecte de production sur Render

Un Background Worker Render n'est pas necessaire. Le Web Service API exporte
directement les metriques et les traces en OTLP vers Grafana Cloud, tandis que
les logs partent directement vers Loki. `/metrics` reste disponible pour la
stack Prometheus locale.

Dans **Render > Web Service API > Environment**, configurer les valeurs donnees
par la tuile de connexion OpenTelemetry de Grafana Cloud :

- `OTEL_EXPORTER_OTLP_ENDPOINT` : endpoint OTLP se terminant normalement par
  `/otlp` ;
- `OTEL_EXPORTER_OTLP_PROTOCOL=http/protobuf` ;
- `OTEL_EXPORTER_OTLP_HEADERS` : en-tete `Authorization` fourni par Grafana ;
- `OTEL_DEPLOYMENT_ENVIRONMENT=production` ;
- `OTEL_SERVICE_NAMESPACE=streampulse` (ou le namespace retenu).

Le token associe a l'en-tete OTLP doit avoir les scopes `metrics:write` et
`traces:write`. `OTEL_METRICS_ENABLED` vaut automatiquement `true` lorsque
`APP_ENV=production`; il peut etre fixe explicitement a `true` sur Render pour
rendre l'intention visible.

L'API exporte les metriques toutes les 15 secondes avec les labels
`service="streampulse-api"` et `env="production"`. Le job de deploiement attend
un cycle d'export et execute `scripts/prove-observability.sh`. Il echoue si l'un
des trois constats manque : metrique HTTP StreamPulse, log Loki de production,
trace Tempo de production. L'artefact conserve uniquement les labels,
timestamps et identifiants de trace, jamais le contenu des logs.

Variables GitHub requises pour la lecture de preuve :

- `GRAFANA_PROMETHEUS_QUERY_URL` (base terminee par `/api/prom`) ;
- `GRAFANA_LOKI_QUERY_URL` ;
- `GRAFANA_TEMPO_QUERY_URL` (base terminee par `/tempo`).

Secrets GitHub requis : les couples username/token de lecture
`GRAFANA_PROMETHEUS_*`, `GRAFANA_LOKI_*` et `GRAFANA_TEMPO_*` utilises par le
workflow. Ils peuvent provenir d'une seule Access Policy Grafana Cloud avec les
scopes de lecture adaptes.

## Dashboard livre

Dashboard : `StreamPulse - RNCP Observability`

Fichier source :

- `go/infra/grafana/dashboards/streampulse.json`

Panels metier :

- Online users : `streampulse_online_users`
- Active streams : `streampulse_active_streams`
- Total listeners : `sum(streampulse_listeners_per_stream)`
- Stream starts - last hour : `increase(streampulse_stream_start_total[1h])`
- Listeners by stream : `streampulse_listeners_per_stream`
- Audience and active streams over time

Panels audio reels :

- Audio ingest bitrate : octets recus sur la route d'ingestion
- Audio egress bitrate : octets remis aux auditeurs HTTP
- Dropped audio chunks : backpressure des clients lents
- Connected audio publishers
- Real audio bitrate by stream
- Audio health : chunks livres et abandonnes

Panels techniques :

- API latency p50 / p95 / p99
- 5xx error rate
- Requests per second by status
- 5xx errors and disconnects per minute
- Slowest API routes p95

## Metriques exposees par l'API

L'API expose les metriques Prometheus sur `/metrics`, protege par bearer token
si `METRICS_BEARER_TOKEN` ou `METRICS_BEARER_TOKEN_FILE` est configure.
En production, les memes mesures sont aussi envoyees directement en OTLP a
Grafana Cloud, sans service Alloy payant.

| Metrique | Type | Usage |
| --- | --- | --- |
| `streampulse_online_users` | Gauge | Utilisateurs uniques actuellement connectes a un stream |
| `streampulse_active_streams` | Gauge | Streams live en cours |
| `streampulse_listeners_per_stream{stream_id}` | Gauge | Listeners actifs par stream |
| `streampulse_stream_start_total` | Counter | Nombre total de streams demarres |
| `streampulse_listener_disconnect_total` | Counter | Nombre total de deconnexions listeners |
| `streampulse_audio_ingest_bytes_total{stream_id}` | Counter | Octets effectivement recus du diffuseur |
| `streampulse_audio_egress_bytes_total{stream_id}` | Counter | Octets effectivement ecrits dans les reponses HTTP auditeurs |
| `streampulse_audio_chunks_total{stream_id,direction}` | Counter | Chunks audio traites en entree/sortie |
| `streampulse_audio_dropped_chunks_total{stream_id}` | Counter | Chunks abandonnes pour clients lents |
| `streampulse_audio_broadcasters{stream_id}` | Gauge | Sources audio reellement connectees |
| `streampulse_audio_chunk_size_bytes` | Histogram | Taille des chunks recus |
| `streampulse_audio_listener_session_duration_seconds` | Histogram | Duree des connexions audio auditeur |
| `streampulse_audio_broadcaster_session_duration_seconds` | Histogram | Duree des connexions d'ingestion |
| `streampulse_api_request_duration_seconds` | Histogram | Latence HTTP, debit et erreurs par route/methode/status |

## Alerting simple

Les alertes locales sont definies dans :

- `go/infra/prometheus/alerts.yml`

Regles livrees :

- `StreamPulseHigh5xxRate` : taux 5xx > 5% pendant 5 minutes
- `StreamPulseHighLatencyP95` : latence p95 > 500 ms pendant 5 minutes
- `StreamPulseDisconnectSpike` : plus de 20 deconnexions listeners par minute
  pendant 5 minutes
- `StreamPulseAudioDrops` : plus de 1 % de chunks audio abandonnes pendant 2 minutes
- `StreamPulsePublisherWithoutListeners` : source connectee sans auditeur pendant 10 minutes

Ces seuils sont volontairement simples pour la soutenance. Ils montrent les
risques principaux : indisponibilite API, degradation de performance et rupture
d'experience d'ecoute.

## Verification locale

Demarrer la stack :

```powershell
docker-compose up -d --build
```

Verifier les cibles Prometheus :

```powershell
curl.exe http://localhost:9090/-/ready
```

Ouvrir Grafana :

- URL : http://localhost:3000
- Dashboard : dossier `StreamPulse`, dashboard `StreamPulse - RNCP Observability`

Verifier les alertes Prometheus :

- http://localhost:9090/alerts

## Requetes PromQL utiles pour la soutenance

Taux d'erreur 5xx :

```promql
sum(rate(streampulse_api_request_duration_seconds_count{status=~"5.."}[5m]))
/
clamp_min(sum(rate(streampulse_api_request_duration_seconds_count[5m])), 0.001)
```

Latence p95 :

```promql
histogram_quantile(
  0.95,
  sum by (le) (rate(streampulse_api_request_duration_seconds_bucket[5m]))
)
```

Listeners totaux :

```promql
sum(streampulse_listeners_per_stream)
```

Deconnexions par minute :

```promql
rate(streampulse_listener_disconnect_total[5m]) * 60
```
