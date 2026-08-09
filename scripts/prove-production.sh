#!/usr/bin/env bash
set -euo pipefail

for command in curl jq openssl; do
  command -v "$command" >/dev/null 2>&1 || {
    echo "Missing required command: $command" >&2
    exit 1
  }
done

: "${PRODUCTION_ORIGIN:?Set PRODUCTION_ORIGIN to the public HTTPS origin}"
: "${RENDER_API_KEY:?Set RENDER_API_KEY}"
: "${RENDER_SERVICE_ID:?Set RENDER_SERVICE_ID}"
: "${RENDER_DEPLOY_ID:?Set RENDER_DEPLOY_ID}"
: "${EXPECTED_IMAGE_DIGEST:?Set EXPECTED_IMAGE_DIGEST to sha256:...}"

case "$PRODUCTION_ORIGIN" in
  https://*) ;;
  *) echo "PRODUCTION_ORIGIN must start with https://" >&2; exit 1 ;;
esac

origin="${PRODUCTION_ORIGIN%/}"
authority="${origin#https://}"
host="${authority%%/*}"
if [[ "$authority" != "$host" || -z "$host" || "$host" == "localhost" ]]; then
  echo "PRODUCTION_ORIGIN must be an HTTPS origin without a path" >&2
  exit 1
fi

evidence_dir="${EVIDENCE_DIR:-production-evidence}"
mkdir -p "$evidence_dir"
render_url="https://api.render.com/v1/services/${RENDER_SERVICE_ID}/deploys/${RENDER_DEPLOY_ID}"

echo "Waiting for Render deploy ${RENDER_DEPLOY_ID}..."
deploy_json=''
for attempt in $(seq 1 90); do
  deploy_json="$(curl --fail --silent --show-error \
    --header "Authorization: Bearer ${RENDER_API_KEY}" \
    --header "Accept: application/json" \
    "$render_url")"
  status="$(jq -er '.status' <<<"$deploy_json")"
  echo "attempt=${attempt} status=${status}"
  case "$status" in
    live) break ;;
    build_failed|update_failed|pre_deploy_failed|canceled|deactivated)
      echo "Render deploy ended with status ${status}" >&2
      exit 1
      ;;
  esac
  sleep 10
done

status="$(jq -er '.status' <<<"$deploy_json")"
[[ "$status" == "live" ]] || {
  echo "Render deploy did not become live before the timeout" >&2
  exit 1
}

jq '{id, status, trigger, createdAt, startedAt, finishedAt, image: (.image | {ref, sha})}' \
  <<<"$deploy_json" >"$evidence_dir/render-deploy.json"
actual_digest="$(jq -er '.image.sha' <<<"$deploy_json")"
if [[ "$actual_digest" != "$EXPECTED_IMAGE_DIGEST" ]]; then
  echo "Render image digest mismatch: expected ${EXPECTED_IMAGE_DIGEST}, got ${actual_digest}" >&2
  exit 1
fi

health_url="${origin}/health"
curl --fail --silent --show-error --verbose \
  --connect-timeout 20 --max-time 120 \
  "$health_url" \
  >"$evidence_dir/health.json" \
  2>"$evidence_dir/curl-tls-verbose.txt"
jq -e '.status == "ok" and .service == "streampulse-api"' \
  "$evidence_dir/health.json" >/dev/null

http_code="$(curl --silent --show-error \
  --max-redirs 0 --connect-timeout 20 --max-time 30 \
  --dump-header "$evidence_dir/http-redirect-headers.txt" \
  --output /dev/null --write-out '%{http_code}' \
  "http://${host}/health")"
case "$http_code" in
  301|302|307|308) ;;
  *) echo "Expected an HTTP-to-HTTPS redirect, got ${http_code}" >&2; exit 1 ;;
esac
redirect_location="$(awk 'BEGIN {IGNORECASE=1} /^location:/ {sub(/^[^:]+:[[:space:]]*/, ""); sub(/\r$/, ""); print; exit}' \
  "$evidence_dir/http-redirect-headers.txt")"
[[ "$redirect_location" == https://* ]] || {
  echo "Redirect location is not HTTPS: ${redirect_location}" >&2
  exit 1
}

openssl s_client -connect "${host}:443" -servername "$host" -showcerts </dev/null \
  >"$evidence_dir/certificate-chain.pem" \
  2>"$evidence_dir/openssl-handshake.txt"
openssl x509 -in "$evidence_dir/certificate-chain.pem" -noout \
  -subject -issuer -serial -dates -fingerprint -sha256 \
  >"$evidence_dir/certificate.txt"
openssl x509 -in "$evidence_dir/certificate-chain.pem" -noout -checkend 604800 >/dev/null

cat >"$evidence_dir/summary.md" <<EOF
# Production deployment evidence

- Timestamp (UTC): $(date -u +'%Y-%m-%dT%H:%M:%SZ')
- HTTPS health URL: ${health_url}
- Render service: ${RENDER_SERVICE_ID}
- Render deploy: ${RENDER_DEPLOY_ID}
- Render status: ${status}
- Image digest: ${actual_digest}
- HTTP redirect status: ${http_code}
- HTTP redirect location: ${redirect_location}
- Health payload: \`$(jq -c . "$evidence_dir/health.json")\`

The artifact also contains the sanitized Render deploy response, verbose curl
TLS negotiation, redirect headers, the leaf certificate details and chain.
EOF

echo "Production evidence written to ${evidence_dir}"
