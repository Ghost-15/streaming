#!/usr/bin/env bash
set -euo pipefail

for command in curl jq; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Missing required command: $command" >&2
    exit 1
  }
done

: "${GRAFANA_PROMETHEUS_QUERY_URL:?Set the Grafana Cloud Prometheus query base URL}"
: "${GRAFANA_PROMETHEUS_USERNAME:?Set the Grafana Cloud Prometheus username}"
: "${GRAFANA_PROMETHEUS_API_KEY:?Set the Grafana Cloud Prometheus read token}"
: "${GRAFANA_LOKI_QUERY_URL:?Set the Grafana Cloud Loki base URL}"
: "${GRAFANA_LOKI_USERNAME:?Set the Grafana Cloud Loki username}"
: "${GRAFANA_LOKI_API_KEY:?Set the Grafana Cloud Loki read token}"
: "${GRAFANA_TEMPO_QUERY_URL:?Set the Grafana Cloud Tempo query base URL}"
: "${GRAFANA_TEMPO_USERNAME:?Set the Grafana Cloud Tempo username}"
: "${GRAFANA_TEMPO_API_KEY:?Set the Grafana Cloud Tempo read token}"

evidence_dir="${EVIDENCE_DIR:-production-evidence/observability}"
mkdir -p "$evidence_dir"
now="$(date +%s)"
start="$((now - 1800))"

prometheus_json="$(curl --fail --silent --show-error --get \
  --user "${GRAFANA_PROMETHEUS_USERNAME}:${GRAFANA_PROMETHEUS_API_KEY}" \
  --data-urlencode 'query=up{service="streampulse-api",env="production"}' \
  "${GRAFANA_PROMETHEUS_QUERY_URL%/}/api/v1/query")"
prometheus_samples="$(jq -er '[.data.result[] | select(.value[1] == "1")] | length' <<<"$prometheus_json")"
(( prometheus_samples > 0 )) || {
  echo "No healthy production Prometheus target found" >&2
  exit 1
}
jq '{status, resultType: .data.resultType, healthy_samples: [.data.result[] | select(.value[1] == "1") | {metric, timestamp: .value[0]}]}' \
  <<<"$prometheus_json" >"$evidence_dir/prometheus.json"

loki_json="$(curl --fail --silent --show-error --get \
  --user "${GRAFANA_LOKI_USERNAME}:${GRAFANA_LOKI_API_KEY}" \
  --data-urlencode 'query={service="streampulse-api",env="production"}' \
  --data-urlencode "start=${start}000000000" \
  --data-urlencode "end=${now}000000000" \
  --data-urlencode 'limit=100' \
  --data-urlencode 'direction=backward' \
  "${GRAFANA_LOKI_QUERY_URL%/}/loki/api/v1/query_range")"
loki_entries="$(jq -er '[.data.result[].values[]] | length' <<<"$loki_json")"
(( loki_entries > 0 )) || {
  echo "No production Loki entry found in the last 30 minutes" >&2
  exit 1
}
jq '{status, stream_count: (.data.result | length), entry_count: ([.data.result[].values[]] | length), streams: [.data.result[] | {labels: .stream, newest_timestamp_ns: .values[0][0]}]}' \
  <<<"$loki_json" >"$evidence_dir/loki.json"

tempo_json="$(curl --fail --silent --show-error --get \
  --user "${GRAFANA_TEMPO_USERNAME}:${GRAFANA_TEMPO_API_KEY}" \
  --data-urlencode 'q={ resource.service.name = "streampulse-api" && resource.deployment.environment = "production" }' \
  --data-urlencode "start=${start}" \
  --data-urlencode "end=${now}" \
  --data-urlencode 'limit=20' \
  "${GRAFANA_TEMPO_QUERY_URL%/}/api/search")"
tempo_traces="$(jq -er '(.traces // []) | length' <<<"$tempo_json")"
(( tempo_traces > 0 )) || {
  echo "No production Tempo trace found in the last 30 minutes" >&2
  exit 1
}
jq '{trace_count: ((.traces // []) | length), traces: [(.traces // [])[] | {traceID, rootServiceName, rootTraceName, startTimeUnixNano, durationMs}]}' \
  <<<"$tempo_json" >"$evidence_dir/tempo.json"

cat >"$evidence_dir/summary.md" <<EOF
# Production observability evidence

- Timestamp (UTC): $(date -u +'%Y-%m-%dT%H:%M:%SZ')
- Prometheus healthy production samples: ${prometheus_samples}
- Loki production entries in the last 30 minutes: ${loki_entries}
- Tempo production traces in the last 30 minutes: ${tempo_traces}

The JSON files retain labels, timestamps and trace identifiers, but deliberately
exclude log messages and all authentication material.
EOF

echo "Production observability evidence written to ${evidence_dir}"
