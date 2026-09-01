# Cycle de vie, timeouts et annulation du flux

## Portée

Le plan de données audio utilise HTTP derrière le reverse proxy HTTPS Render :

- `POST /api/v1/streams/:id/push` reçoit les blobs WebM/Opus ordonnés de
  `MediaRecorder` (4 Mio maximum par blob) ;
- `PUT /api/v1/streams/:id/audio` conserve le contrat de corps continu pour
  FFmpeg et les producteurs non-Web ;
- `GET /api/v1/streams/:id/audio/ws` envoie à Flutter Web les octets WebM
  ordonnés, avec une reprise alignée sur un élément `Cluster` ;
- `GET /api/v1/streams/:id/audio` conserve le flux HTTP public de secours ;
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

Le lecteur Web utilise une connexion WebSocket persistante, mais ne considère
plus chaque blob `MediaRecorder` comme un segment autonome. Le Hub extrait du
premier blob uniquement l'initialisation EBML/Info/Tracks située avant le
premier élément `Cluster`. Un auditeur tardif reçoit ces métadonnées, puis le
serveur ignore la fin du Cluster déjà commencé et reprend au Cluster suivant.
Ainsi, aucun ancien son n'est rejoué et aucun fragment courant n'est raccordé
au milieu d'un élément WebM.

Le `MediaSource` donne toujours la priorité aux blobs entrants. Son historique
n'est nettoyé que lorsque la fenêtre dépasse 75 secondes, puis 45 secondes sont
conservées ; cela évite une opération `remove` à chaque blob. Un watchdog remet
le lecteur à environ 0,75 seconde du direct, uniquement par un saut vers l'avant,
et relance `play()` si les données continuent d'arriver alors que l'horloge audio
n'avance plus. Le lecteur ne revient jamais sur un fragment déjà joué lorsque
son horloge dépasse brièvement la fin du tampon. Une pause demandée explicitement
par l'utilisateur reste, elle, respectée.

Une fermeture WebSocket passagère déclenche une reconstruction complète du
`MediaSource`, avec un délai progressif de 500 ms à 8 s, tandis que l'état du
live est contrôlé par l'API. Cette reconstruction reprend sur un nouveau
Cluster et évite de continuer avec un parseur WebM qui aurait perdu des octets.
Un arrêt confirmé termine le `MediaSource`.

## Propagation de l’annulation

| Événement | Effet |
| --- | --- |
| Auditeur ferme la connexion | `Request.Context()` est annulé, le client est retiré du Hub, son channel est fermé et le leave BDD s’exécute avec un contexte détaché borné |
| Diffuseur ferme le corps PUT | EOF ferme le publisher et observe la durée de session |
| Diffuseur Web arrête | le dernier blob est vidé, puis `/stop` ferme le publisher chunké et tous les auditeurs |
| Stream arrêté | `Hub.CloseStream` annule l’ingestion et ferme tous les channels auditeurs |
| SIGINT/SIGTERM | Le contexte racine est annulé, le Hub est fermé avant `http.Server.Shutdown`, puis l’arrêt est borné |
| Client lent | Son channel borné ne bloque jamais le diffuseur ; la connexion est fermée et reconstruite plutôt que de continuer après une perte d’octets |
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
- un channel plein ne bloque pas le Hub et déconnecte l’auditeur lent ;
- un payload ingéré est reçu à l’identique par un auditeur HTTP ;
- deux auditeurs publics reçoivent le même blob envoyé par `POST /push` ;
- le cache WebM exclut le premier Cluster et donc le son du début ;
- une reprise tardive attend une frontière Cluster, même coupée entre uploads ;
- une session arrêtée ne peut ni rouvrir un publisher ni envoyer un chunk ;
- une reprise renouvelle la session tout en conservant l'identifiant du live ;
- l’annulation auditeur ramène le delta BDD join/leave à zéro ;
- `go test ./...` couvre le Hub, les sessions et les handlers audio ; le mode
  `-race` peut être ajouté sur un environnement Go avec CGO activé.

## Limites connues

La spécification `MediaRecorder` ne garantit pas qu'un blob produit par
`timeslice` soit lisible isolément. L'alignement WebM corrige les arrivées
tardives sur les navigateurs WebM/Opus pris en charge, mais une évolution vers
WebRTC avec SFU reste la cible recommandée pour une diffusion multi-navigateurs,
multi-répliques et plusieurs heures avec gestion native de la gigue et des
pertes réseau. Les routes REST de cycle de vie resteraient inchangées.

Le Hub est en mémoire : une réplique API ne partage pas ses publishers avec une
autre. En production mono-nœud, c’est cohérent et simple. Un déploiement
multi-répliques exigera une affinité `stream_id` ou un bus média externe
(NATS/Redis Streams spécialisé, WebRTC SFU, Icecast, etc.).
