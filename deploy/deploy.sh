#!/usr/bin/env sh
set -eu

: "${IMAGE_REPOSITORY:?IMAGE_REPOSITORY is required}"
: "${IMAGE_TAG:?IMAGE_TAG is required}"

COMPOSE_FILE="${COMPOSE_FILE:-compose.prod.yml}"
ENV_FILE="${ENV_FILE:-.env.production}"
STATE_FILE="${STATE_FILE:-.deployed-tag}"

if [ ! -f "$COMPOSE_FILE" ] || [ ! -f "$ENV_FILE" ]; then
  echo "Missing $COMPOSE_FILE or $ENV_FILE" >&2
  exit 1
fi

for secret in private.pem public.pem metrics_bearer_token grafana_admin_password; do
  if [ ! -s "secrets/$secret" ]; then
    echo "Missing non-empty secrets/$secret" >&2
    exit 1
  fi
done

API_DOMAIN="$(sed -n 's/^API_DOMAIN=//p' "$ENV_FILE" | tail -n 1)"
if [ -z "$API_DOMAIN" ]; then
  echo "API_DOMAIN is missing from $ENV_FILE" >&2
  exit 1
fi

previous_tag=""
if [ -f "$STATE_FILE" ]; then
  previous_tag="$(cat "$STATE_FILE")"
fi

deploy_tag() {
  tag="$1"
  IMAGE_REPOSITORY="$IMAGE_REPOSITORY" IMAGE_TAG="$tag" \
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" pull
  IMAGE_REPOSITORY="$IMAGE_REPOSITORY" IMAGE_TAG="$tag" \
    docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" up -d --remove-orphans
}

echo "Deploying $IMAGE_REPOSITORY:$IMAGE_TAG"
deploy_tag "$IMAGE_TAG"

healthy=false
attempt=1
while [ "$attempt" -le 30 ]; do
  if curl --fail --silent --show-error --max-time 5 "https://$API_DOMAIN/health" >/dev/null; then
    healthy=true
    break
  fi
  sleep 4
  attempt=$((attempt + 1))
done

if [ "$healthy" != "true" ]; then
  echo "Health check failed for $IMAGE_TAG" >&2
  if [ -n "$previous_tag" ] && [ "$previous_tag" != "$IMAGE_TAG" ]; then
    echo "Rolling back to $previous_tag" >&2
    deploy_tag "$previous_tag"
  fi
  exit 1
fi

temporary_state="${STATE_FILE}.tmp"
printf '%s' "$IMAGE_TAG" >"$temporary_state"
mv "$temporary_state" "$STATE_FILE"
echo "Deployment healthy: https://$API_DOMAIN ($IMAGE_TAG)"
