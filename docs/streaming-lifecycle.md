# Cycle de vie, timeouts et annulation du flux

## Portée

Le plan de données audio utilise HTTP derrière le reverse proxy HTTPS Render :

- `PUT /api/v1/streams/:id/audio` reçoit un corps binaire chunked ;
- `GET /api/v1/streams/:id/listen` renvoie le même média en continu ;
- le Hub effectue un fan-out non bloquant vers un channel borné par connexion.

Le `POST /listen` reste un événement métier court. Seule la connexion GET
alimente les gauges d’auditeurs audio réels.

## Propagation de l’annulation

| Événement | Effet |
| --- | --- |
| Auditeur ferme la connexion | `Request.Context()` est annulé, le client est retiré du Hub, son channel est fermé et le leave BDD s’exécute avec un contexte détaché borné |
| Diffuseur ferme le corps PUT | EOF ferme le publisher et observe la durée de session |
| Stream arrêté | `Hub.CloseStream` annule l’ingestion et ferme tous les channels auditeurs |
| SIGINT/SIGTERM | Le contexte racine est annulé, le Hub est fermé avant `http.Server.Shutdown`, puis l’arrêt est borné |
| Client lent | Son channel borné ne bloque jamais le diffuseur ; le chunk est abandonné et métriqué |
| BDD lente après déconnexion | Le leave utilise un timeout de 3 s et ne retient pas la connexion audio |

Une copie immuable unique du chunk est partagée entre les channels. Le buffer
de lecture du diffuseur peut donc être réutilisé sans corruption et sans une
allocation de 32 Kio par auditeur.

## Matrice des limites

| Variable | Défaut | Protection |
| --- | ---: | --- |
| `HTTP_READ_HEADER_TIMEOUT` | 10 s | Slowloris sur les en-têtes |
| `HTTP_IDLE_TIMEOUT` | 60 s | Connexions keep-alive HTTP inactives |
| `STREAM_MAX_DURATION` | 6 h | Ingestion infinie ou oubliée |
| `STREAM_IDLE_TIMEOUT` | 30 s | Diffuseur silencieux et auditeur sans paquet |
| `STREAM_WRITE_TIMEOUT` | 10 s | Socket auditeur bloqué |
| `STREAM_MAX_INGEST_BYTES` | 8 Gio | Volume maximal par requête d’ingestion |
| `STREAM_CHUNK_SIZE` | 32 Kio | Allocation de lecture bornée |
| `STREAM_CLIENT_BUFFER` | 64 chunks | Mémoire et backpressure bornées par client |
| `SHUTDOWN_TIMEOUT` | 30 s | Durée maximale d’arrêt gracieux |

Les `ReadTimeout`/`WriteTimeout` globaux de `http.Server` restent à zéro, car ce
sont des délais absolus incompatibles avec un stream de plusieurs heures. Les
handlers posent à la place une deadline glissante avant chaque lecture/écriture.

## Invariants testés

- un stream n’accepte qu’un publisher actif ;
- `CloseStream` annule le publisher et ferme chaque auditeur ;
- `Shutdown` est idempotent et refuse les nouvelles connexions ;
- un channel plein ne bloque pas le Hub ;
- un payload ingéré est reçu à l’identique par un auditeur HTTP ;
- l’annulation auditeur ramène le delta BDD join/leave à zéro ;
- `go test -race ./...` vérifie les accès concurrents.

## Limites connues

Le Hub est en mémoire : une réplique API ne partage pas ses publishers avec une
autre. En production mono-nœud, c’est cohérent et simple. Un déploiement
multi-répliques exigera une affinité `stream_id` ou un bus média externe
(NATS/Redis Streams spécialisé, WebRTC SFU, Icecast, etc.).
