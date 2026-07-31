# StreamPulse

[English documentation](README.en.md)

StreamPulse est une plateforme de streaming audio en direct composée d’une API
Go, d’un client Flutter et d’une stack d’observabilité Prometheus/Grafana. Le
backend suit une Clean Architecture, diffuse réellement les octets d’un
diffuseur vers plusieurs auditeurs et gère les connexions longues sans fuite de
goroutine lors des déconnexions ou d’un arrêt serveur.

> **English summary:** StreamPulse is a Go + Flutter live-audio platform. It
> ships real HTTP chunked audio fan-out, JWT/RBAC, graceful cancellation,
> Prometheus/Grafana observability, reproducible load tests, an HTTPS Caddy
> reverse proxy and an immutable production deployment pipeline.

## Ce qui est livré

- API REST Go 1.25, authentification JWT RS256 et rôles User/Diffuseur/Admin.
- Streaming HTTP chunked : une ingestion `audio/*`, plusieurs auditeurs et
  éviction non bloquante des clients lents.
- Annulation transverse par `context.Context`, timeouts glissants, fermeture du
  Hub avant le graceful shutdown HTTP.
- Métriques audio issues des octets réellement ingérés/diffusés, dashboard
  Grafana et alertes Prometheus.
- Tests Go, race detector, seuil de couverture CI, tests Flutter et build web.
- Benchmark 10/100/500 auditeurs, scénario k6, CPU/heap/goroutines via pprof.
- Production Docker Compose derrière Caddy (TLS ACME), secrets en fichiers,
  image Docker Hub par SHA, health-check et rollback automatique.

## Architecture

```text
Flutter / diffuseur --HTTPS--> Caddy --HTTP interne--> API Go
                                                    ├── PostgreSQL/Supabase
                                                    ├── Hub audio en mémoire
                                                    └── /metrics (réseau interne)
Prometheus <---------------------------------------------┘
Grafana <---------- Prometheus
```

Le Hub est volontairement local au processus. Une seule réplique API doit donc
traiter l’ingestion et ses auditeurs. Pour passer à plusieurs répliques, il faut
ajouter un bus média partagé ou une affinité de session ; cette limite est
documentée dans [la note de performance et de coûts](docs/performance-couts.md).

## Démarrage local

Prérequis : Docker avec Compose, Go 1.25+, Flutter stable et OpenSSL.

```powershell
Copy-Item .env.example .env
New-Item -ItemType Directory -Force secrets
openssl genrsa -out secrets/private.pem 4096
openssl rsa -in secrets/private.pem -pubout -out secrets/public.pem
[guid]::NewGuid().ToString("N") | Set-Content -NoNewline secrets/metrics_bearer_token
docker compose up -d --build
```

Renseigner auparavant `SUPABASE_DB_URL` et `CORS_ALLOWED_ORIGINS` dans `.env`.
Services locaux :

- API : `http://localhost:8080`
- Prometheus : `http://localhost:9090`
- Grafana : `http://localhost:3000`
- pprof : `http://127.0.0.1:6060/debug/pprof/`

Le guide détaillé est dans [docs/runbook-local.md](docs/runbook-local.md).

## Protocole audio

1. Un diffuseur crée un stream avec `POST /api/v1/streams`.
2. Il ouvre `PUT /api/v1/streams/{id}/audio` avec un bearer JWT, un
   `Content-Type: audio/mpeg` (ou autre `audio/*`) et un corps binaire continu.
3. Les auditeurs ouvrent `GET /api/v1/streams/{id}/listen` avec leur bearer JWT.
4. L’arrêt via `PUT /api/v1/streams/{id}/stop`, une déconnexion TCP, un timeout
   ou SIGTERM libère les channels, compteurs et goroutines.

Exemple de source audio temps réel avec FFmpeg :

```bash
ffmpeg -re -stream_loop -1 -i sample.mp3 -codec copy -f mp3 - \
  | curl --no-buffer --request PUT \
      --header "Authorization: Bearer $BROADCASTER_TOKEN" \
      --header "Content-Type: audio/mpeg" \
      --data-binary @- \
      "$API_BASE_URL/streams/$STREAM_ID/audio"
```

Le client Flutter démarre réellement `just_audio` sur la route GET avec le JWT
en en-tête ; cette connexion réalise elle-même le join/leave métier. Les détails sont dans
[docs/streaming-lifecycle.md](docs/streaming-lifecycle.md).

## Tests et performance

```powershell
cd go
$env:GOCACHE = "$PWD\.gocache"
go test ./...
go test ./... -race
cd ..
make load-bench
```

Le test d’intégration audio vérifie le fan-out binaire et l’annulation d’un
auditeur. Le mode opératoire k6 et pprof se trouve dans
[go/loadtest/README.md](go/loadtest/README.md). Les résultats mesurés et le
modèle de capacité/coûts sont dans
[docs/performance-couts.md](docs/performance-couts.md).

## Observabilité

`/metrics` est protégé par bearer token et n’est pas exposé par Caddy en
production. Le dashboard provisionné contient :

- auditeurs réels et diffuseurs connectés ;
- débit audio entrant/sortant par stream ;
- chunks livrés et abandonnés ;
- durée des sessions ;
- latences HTTP, erreurs 5xx et déconnexions.

Voir [docs/observability-rncp.md](docs/observability-rncp.md).

## Production HTTPS

La procédure complète, de la création DNS au rollback, est dans
[docs/deploiement-https.md](docs/deploiement-https.md).

```bash
cp .env.production.example .env.production
# Renseigner les domaines, la BDD et créer les quatre fichiers secrets.
IMAGE_REPOSITORY=utilisateur-dockerhub/streampulse-api \
IMAGE_TAG=<git-sha> \
./deploy/deploy.sh
```

Le workflow `.github/workflows/deploy.yml` vérifie Go et Flutter, puis publie
l'image sur Docker Hub avec SBOM/provenance. Sur la branche `main`, il appelle
ensuite le deploy hook Render. Si `ENABLE_VPS_DEPLOY=true`, il peut également
transférer la configuration sur un serveur autonome, attendre
`https://API_DOMAIN/health` et revient au SHA précédent si le contrôle échoue.

## Documentation

- [Runbook local](docs/runbook-local.md)
- [Cycle de vie streaming](docs/streaming-lifecycle.md)
- [Observabilité RNCP](docs/observability-rncp.md)
- [Charge, pprof et coûts](docs/performance-couts.md)
- [Déploiement HTTPS](docs/deploiement-https.md)
- ADR : [go/docs/adr](go/docs/adr)

## Sécurité

Ne jamais committer `.env`, `.env.production`, les clés JWT, tokens métriques,
mot de passe Grafana ou clé SSH. En production, seuls les ports 22 (restreint),
80 et 443 sont publics ; pprof, Prometheus, PostgreSQL et le port 8080 restent
privés.

## Licence

Projet pédagogique StreamPulse. Ajouter ici la licence choisie par l’équipe
avant toute distribution externe.
