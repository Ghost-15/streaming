# ADR-007 — Protocole de diffusion audio live (relais HTTP + WebSocket)

## Statut
Accepté — Sprint Final. Remplace la première version (mono-chunk HTTP) par la
version à sessions de diffusion + variante WebSocket.

## Contexte
StreamPulse doit diffuser en direct l'audio d'un diffuseur vers plusieurs
auditeurs simultanément. Il fallait choisir un protocole pour :
1. l'**ingestion** (le diffuseur envoie sa source audio au serveur) ;
2. la **diffusion** (le serveur relaie les octets à tous les auditeurs connectés).

Options envisagées : WebRTC, HLS (segments), WebSocket binaire, HTTP chunked.

Le client de référence est un **navigateur** : capture micro via `MediaRecorder`
(WebM/Opus) qui émet un **blob toutes les ~250 ms**. Ce découpage en blobs
discrets — et non un corps HTTP continu — a guidé la décision.

## Décision
**Relais en mémoire** via le `Hub` concurrent Go (goroutines + channels, aucune
écriture DB dans le chemin chaud). Le Hub expose **trois chemins qui coexistent**,
tous alimentés par le même fan-out non bloquant :

### Ingestion (diffuseur)
- **`POST /api/v1/streams/:id/push`** *(chemin principal, navigateur)* — rôle
  Diffuseur/Admin. Un POST court par blob `MediaRecorder`. C'est le contrat par
  défaut du client Web.
- **`PUT /api/v1/streams/:id/audio`** *(source binaire continue)* — pour une
  source type FFmpeg qui streame un corps HTTP unique et long.

Les deux sont mutuellement exclusifs pour un même stream (un seul publisher).

### Diffusion (auditeur)
- **`GET /api/v1/streams/:id/audio`** *(public)* — un élément `<audio>` ne peut
  pas poser d'en-tête `Authorization`. Le serveur enregistre un `Client`, envoie
  immédiatement les **métadonnées d'initialisation WebM** en cache, puis écrit
  chaque paquet via `http.Flusher` en transfert chunked jusqu'à déconnexion.
- **`GET /api/v1/streams/:id/audio/ws`** *(WebSocket, public)* — même flux, en
  messages binaires. Un auditeur qui rejoint en cours de route est **resynchronisé
  à la prochaine frontière de Cluster WebM** pour éviter un flux tronqué.

### Sessions de diffusion (migration 010 — `active_session_id`)
Un stream est un **canal réutilisable** du diffuseur. À chaque passage en live,
une nouvelle `active_session_id` (UUID) est générée et transmise par le header
**`X-Stream-Session-ID`**. Le Hub n'accepte des chunks (`OpenOwnedChunkPublisher`
/ `BroadcastOwnedChunk`) que pour la session **autorisée** courante. Conséquence :
un chunk retardé d'une **session périmée** (ancien onglet, requête en vol après un
Stop/Restart) est **rejeté** (`ErrPublisherNotActive`) et ne peut pas ressusciter
ni corrompre la diffusion en cours.

### Cache d'initialisation WebM (cluster-aware)
Le premier blob ne contient pas forcément *que* l'en-tête. Le Hub accumule les
octets jusqu'à trouver le préfixe d'initialisation (EBML/Tracks) **avant le
premier Cluster** (`webMInitializationPrefix`, borné à 1 Mio), puis le met en
cache pour les arrivants tardifs — sans jamais confondre des données média avec
les métadonnées du décodeur.

## Conséquences

### Positives
- **Adapté au navigateur** : les POST discrets collent au découpage `MediaRecorder`
  (pas d'upload streaming navigateur, fragile) ; `<audio>` ou WebSocket côté écoute.
- **Résilient aux sessions concurrentes** : le modèle de session empêche un chunk
  périmé de réveiller un live arrêté (garde testée unitairement).
- **Backpressure gérée** : envoi non bloquant sur channel bufferisé ; un auditeur
  lent voit ses paquets **droppés** sans bloquer le stream ni les autres.
- **Arrivants tardifs** : cache d'init WebM (HTTP) + resynchronisation au prochain
  Cluster (WebSocket).

### Négatives / limites
- **Pas de transcodage / qualité adaptative** (contrairement à HLS) — hors périmètre.
- **Latence** de l'ordre de la seconde (taille des blobs `MediaRecorder`).
- Les connexions longues utilisent des **deadlines glissantes** (read/write) au lieu
  d'un `WriteTimeout` serveur absolu, incompatible avec des flux multi-heures.
- Pas de persistance du flux (pas de rediffusion) — live uniquement.

## Alternatives rejetées
- **WebRTC** : latence minimale mais complexité (signaling, STUN/TURN, SFU) hors budget.
- **HLS** : robuste et scalable mais latence de plusieurs secondes + pipeline de
  segmentation FFmpeg — trop lourd pour la démo.
- **Corps HTTP continu unique côté navigateur** : `fetch` avec `ReadableStream`
  en upload est fragile/expérimental ; les POST discrets sont plus robustes. Le
  `PUT` continu reste néanmoins offert pour les sources non-navigateur.

## Vérification
- `handler.TestStreamingE2E_MediaRecorderPushToTwoListeners` : contrat navigateur
  complet (POST /push → deux auditeurs GET /audio, dont le segment d'init).
- `handler.TestStreamingWebSocket_*` : livraison init + resynchronisation Cluster.
- `streaming.TestHubAuthorizeStreamSession`, `TestHubCloseStreamSession`,
  `TestHubBroadcastOwnedChunkErrors`, `TestHubOpenAndCloseOwnedContinuousPublisher` :
  gardes de session (chunk périmé/étranger rejeté, révocation).
- `TestStreamingIntegration_Disconnect` : libération channel + goroutine à la
  déconnexion.
