# StreamPulse

[English documentation](README.en.md)

StreamPulse est une plateforme de streaming audio en direct composée d’une API
Go, d’un client Flutter et d’une stack d’observabilité Prometheus/Grafana. Le
backend suit une Clean Architecture, diffuse réellement les octets d’un
diffuseur vers plusieurs auditeurs et gère les connexions longues sans fuite de
goroutine lors des déconnexions ou d’un arrêt serveur.

> **English summary:** StreamPulse is a Go + Flutter live-audio platform. It
> ships real HTTP chunked audio fan-out, JWT/RBAC, graceful cancellation,
> Prometheus/Grafana observability, reproducible load tests, Docker Hub image
> publication and an HTTPS Render deployment pipeline.

## Ce qui est livré

- API REST Go 1.26, authentification JWT RS256 et rôles User/Diffuseur/Admin.
- Streaming HTTP chunked : une ingestion `audio/*`, plusieurs auditeurs et
  éviction non bloquante des clients lents.
- Annulation transverse par `context.Context`, timeouts glissants, fermeture du
  Hub avant le graceful shutdown HTTP.
- Métriques audio issues des octets réellement ingérés/diffusés, dashboard
  Grafana et alertes Prometheus.
- Tests Go, race detector, seuil de couverture CI, tests Flutter et build web.
- Benchmark 10/100/500 auditeurs, scénario k6, CPU/heap/goroutines via pprof.
- Image Go publiée sur Docker Hub avec tags SHA/`latest`, SBOM et provenance.
- Déploiement automatique Render par deploy hook, reverse proxy HTTPS géré et
  health-check `/health`.

## Architecture

```text
Flutter / diffuseur --HTTPS--> Render edge --> conteneur API Go
                                                ├── PostgreSQL/Supabase
                                                └── Hub audio en mémoire

Docker Hub <-- image SHA + latest <-- GitHub Actions
Render <-------- deploy hook -------- GitHub Actions
```

Le Hub est volontairement local au processus. Une seule réplique API doit donc
traiter l’ingestion et ses auditeurs. Pour passer à plusieurs répliques, il faut
ajouter un bus média partagé ou une affinité de session ; cette limite est
documentée dans [la note de performance et de coûts](docs/performance-couts.md).

## Démarrage local

Prérequis : Docker avec Compose, Go 1.26+, Flutter stable et OpenSSL.

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
- Contrat OpenAPI : `http://localhost:8080/swagger/index.html`
- Prometheus : `http://localhost:9090`
- Grafana : `http://localhost:3000`
- pprof : `http://127.0.0.1:6060/debug/pprof/`

Le guide détaillé est dans [docs/runbook-local.md](docs/runbook-local.md).

## Contrat d’API

La description OpenAPI est générée depuis les annotations des handlers, jamais
écrite à la main :

```bash
cd go
swag init --dir ./cmd/server,./internal/handler,./internal/entity --generalInfo main.go --output docs/openapi --parseInternal --parseDepth 2
```

Le résultat est versionné dans `go/docs/openapi/` (40 opérations). Deux
garde-fous empêchent la documentation de décrocher du code :

- `TestOpenAPISpecCoversEveryRoute` échoue si une route enregistrée dans le
  routeur n’apparaît pas dans la spécification ;
  `TestOpenAPISpecHasNoStaleOperation` couvre le décalage inverse ;
- le job CI `openapi-contract` régénère la spécification et échoue si la version
  commitée diffère.

L’interface Swagger est servie sur `/swagger/index.html` et se désactive avec
`SWAGGER_ENABLED=false`. Elle n’expose que le contrat : chaque route listée reste
protégée par son propre middleware RBAC.

## Protocole audio

1. Un diffuseur crée un stream avec `POST /api/v1/streams`.
2. Le client Flutter Web capture le micro en WebM/Opus et envoie, dans l'ordre,
   chaque blob `MediaRecorder` vers `POST /api/v1/streams/{id}/push`. Les clients
   continus (FFmpeg) peuvent conserver `PUT /api/v1/streams/{id}/audio`.
3. Flutter lit le direct public sur `GET /api/v1/streams/{id}/audio` et injecte
   la réponse HTTP incrémentale dans `MediaSource`. `POST /listen` et `/leave`
   enregistrent l'historique de l'utilisateur. Le GET authentifié `/listen`
   reste disponible pour les clients non-Web.
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

Le lecteur Web conserve un buffer live borné, revient au bord du direct après
une interruption et expose lecture/pause/arrêt via la Media Session du navigateur
(touches multimédia et écran verrouillé quand le navigateur le permet). Les détails
sont dans [docs/streaming-lifecycle.md](docs/streaming-lifecycle.md).

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

## Fluidité du rendu

La contrainte des 60 FPS est vérifiée par un test d'intégration exécuté en mode
profile sur un appareil, jamais estimée :

```bash
make flutter-perf DEVICE=emulator-5554
```

Le test reconstruit la liste de directs toutes les 16 ms pendant six passes de
défilement et échoue si le p90 du fil UI dépasse 16,67 ms. Le relevé mesuré et
ses limites sont archivés dans
[evidence/performance](evidence/performance/2026-08-23-rendering/README.md).

## Observabilité

`/metrics` est protégé par bearer token. Le dashboard local provisionné contient :

- auditeurs réels et diffuseurs connectés ;
- débit audio entrant/sortant par stream ;
- chunks livrés et abandonnés ;
- durée des sessions ;
- latences HTTP, erreurs 5xx et déconnexions.

Voir [docs/observability-rncp.md](docs/observability-rncp.md).

## Production HTTPS

Le workflow `.github/workflows/deploy.yml` vérifie Go et Flutter, publie
l’image Go sur Docker Hub avec les tags SHA et `latest`, puis appelle le deploy
hook Render après un push réussi sur `main`. Render exécute le conteneur et gère
le reverse proxy HTTPS.

La configuration Docker Hub, Render, Secret Files, domaine et rollback est
détaillée dans [docs/deploiement-https.md](docs/deploiement-https.md).

## Documentation

- [Runbook local](docs/runbook-local.md)
- [Cycle de vie streaming](docs/streaming-lifecycle.md)
- [Observabilité RNCP](docs/observability-rncp.md)
- [Charge, pprof et coûts](docs/performance-couts.md)
- [Déploiement HTTPS](docs/deploiement-https.md)
- ADR : [go/docs/adr](go/docs/adr)

## Sécurité

Ne jamais committer `.env`, les clés JWT, tokens métriques ou deploy hooks.
Configurer ces valeurs dans GitHub et Render. pprof doit rester désactivé en
production.

## Licence

Projet pédagogique StreamPulse. Ajouter ici la licence choisie par l’équipe
avant toute distribution externe.
