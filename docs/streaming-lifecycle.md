# Cycle de vie, timeouts et annulation du flux

## Portée

Le plan de données audio utilise HTTP derrière le reverse proxy HTTPS Render :

- `POST /api/v1/streams/:id/push` reçoit les blobs WebM/Opus ordonnés de
  `MediaRecorder` (4 Mio maximum par blob) ;
- `PUT /api/v1/streams/:id/audio` conserve le contrat de corps continu pour
  FFmpeg et les producteurs non-Web ;
- `GET /api/v1/streams/:id/audio` renvoie le live public utilisé par Flutter Web ;
- `GET /api/v1/streams/:id/listen` fournit la variante authentifiée non-Web ;
- le Hub effectue un fan-out non bloquant vers un channel borné par connexion.

Le `POST /listen` et le `POST /leave` restent des événements métier courts pour
le client Web. Les connexions GET alimentent les gauges d’auditeurs audio réels.

## Live réutilisable et sessions isolées

La ligne `streams` représente désormais un live persistant que son diffuseur
peut reprendre avec le même identifiant. Une diffusion effective est identifiée
séparément par `active_session_id`, renouvelé à chaque création ou reprise :

- `GET /api/v1/streams/mine` liste les lives du diffuseur, actifs ou arrêtés ;
- `PUT /api/v1/streams/:id/start` reprend un live existant ;
- `PUT /api/v1/streams/:id/stop` termine uniquement la session active ;
- `DELETE /api/v1/streams/:id` supprime définitivement un live possédé ;
- chaque envoi audio et chaque `Stop` porte `X-Stream-Session-ID`.

Le Hub conserve la session autorisée même lorsqu'aucun publisher n'est encore
ouvert. `Stop` la révoque atomiquement avec la fermeture des auditeurs. Un chunk
retardé qui avait déjà passé le contrôle BDD ne peut donc pas rouvrir le flux.
De même, un `Stop` retardé d'un ancien onglet ne peut pas arrêter la reprise.
Après un redémarrage du serveur, la première requête valide réhydrate cette
autorisation depuis la BDD ; une reprise explicite est seule autorisée à
remplacer une session révoquée.

## Propagation de l’annulation

| Événement | Effet |
| --- | --- |
| Auditeur ferme la connexion | `Request.Context()` est annulé, le client est retiré du Hub, son channel est fermé et le leave BDD s’exécute avec un contexte détaché borné |
| Diffuseur ferme le corps PUT | EOF ferme le publisher et observe la durée de session |
| Diffuseur Web arrête | le dernier blob est vidé, puis `/stop` ferme le publisher chunké et tous les auditeurs |
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
- deux auditeurs publics reçoivent le même blob envoyé par `POST /push` ;
- le cache du segment d’initialisation et son fan-out sont atomiques ;
- une session arrêtée ne peut ni rouvrir un publisher ni envoyer un chunk ;
- une reprise renouvelle la session tout en conservant l'identifiant du live ;
- l’annulation auditeur ramène le delta BDD join/leave à zéro ;
- `go test -race ./...` vérifie les accès concurrents.

## Limites connues

Le Hub est en mémoire : une réplique API ne partage pas ses publishers avec une
autre. En production mono-nœud, c’est cohérent et simple. Un déploiement
multi-répliques exigera une affinité `stream_id` ou un bus média externe
(NATS/Redis Streams spécialisé, WebRTC SFU, Icecast, etc.).
