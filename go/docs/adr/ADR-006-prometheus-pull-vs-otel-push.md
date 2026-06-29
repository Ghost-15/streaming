# ADR-006 — Prometheus pull (client_golang) vs OTEL metrics push

## Status
Accepted — Sprint 3 (US-010).

## Context
US-010 demande un dashboard Grafana avec 4 panels et alerting sur métriques métier
(`active_streams`, `listeners_per_stream`, `stream_start_total`, `api_request_duration`).

Deux approches possibles pour exporter ces métriques :

1. **Prometheus pull** via `prometheus/client_golang` — l'API expose `/metrics`
   et Prometheus scrape périodiquement (15 s).
2. **OTEL push** via le SDK `go.opentelemetry.io/otel/metric` — l'API pousse les métriques
   au collector OTEL déjà en place (Sprint 2), qui les forwarde vers Prometheus / Grafana Cloud.

## Decision
Prometheus pull avec `client_golang` + `promauto`.

## Consequences

### Positives
- **Simplicité opérationnelle** : un seul registry global, pas de provider/meter à configurer
  par module. Le pattern `promauto.NewCounter(...)` exprime la métrique en une ligne.
- **Stack Grafana Cloud déjà alignée** : Tempo (traces) + Loki (logs) + **Prometheus** (metrics).
  Le datasource Grafana attend des séries Prometheus, pas OTLP metrics.
- **Pull = découplage** : si Prometheus tombe, l'API continue de tourner sans backpressure.
  Avec push OTEL, un collector down peut faire grossir des buffers en mémoire.
- **Discoverabilité** : `curl localhost:8080/metrics` retourne toutes les métriques en texte
  brut — utile pour debug et démo jury (RNCP C3.3).
- **Standard de facto** dans l'écosystème Go : le tooling (alertmanager, exporters, recording rules)
  est mature.

### Négatives
- **Dépendance supplémentaire** : `prometheus/client_golang` ajoute ~3 Mo au binaire distroless.
  Acceptable.
- **Pas d'unification "OTLP-everything"** : on a OTEL pour les traces mais Prometheus pour les
  métriques. Un audit RNCP pourrait demander pourquoi deux protocoles — réponse : maturité
  outillage côté Grafana + cohérence avec la majorité des projets Go production.
- **Cardinality risk** : le label `stream_id` sur `listeners_per_stream` peut exploser si on
  ne nettoie pas les séries des streams terminés. Mitigation : `DeleteLabelValues` à
  l'`Unregister` du dernier listener (à câbler quand l'impl complète du Hub arrivera, Sprint 4).

## Alternatives considered
- **OTEL metrics push** — rejeté pour les raisons ci-dessus. À reconsidérer si le projet bascule
  vers un backend OTLP-natif (Honeycomb, Dynatrace) qui n'aurait pas de scrape Prometheus.
- **StatsD / DogStatsD** — rejeté : pas de support natif Grafana Cloud, requiert un agent supplémentaire.
