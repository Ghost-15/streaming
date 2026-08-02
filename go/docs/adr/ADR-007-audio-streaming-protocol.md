# ADR-007 — Protocole de diffusion audio live (HTTP chunked relay)

## Statut
Accepté — Sprint Final.

## Contexte
StreamPulse doit diffuser en direct l'audio d'un diffuseur vers plusieurs auditeurs
simultanément. Il fallait choisir un protocole de transport pour :
1. l'**ingestion** (le diffuseur envoie sa source audio au serveur) ;
2. la **diffusion** (le serveur relaie les octets audio à tous les auditeurs connectés).

Options envisagées : WebRTC, WebSocket binaire, HLS (segments), HTTP chunked.

## Décision
**Relais HTTP chunked** basé sur le `Hub` concurrent Go (goroutines + channels) :

- **Ingestion** : `POST /api/v1/streams/:id/push` (rôle Diffuseur/Admin).
  Le navigateur capture le micro via `MediaRecorder` (WebM/Opus) et envoie des
  petits chunks (~1/s). Le serveur lit le corps, met en cache le **premier chunk**
  (segment d'init WebM contenant l'en-tête EBML/Tracks) puis appelle `Hub.Broadcast`.
- **Diffusion** : `GET /api/v1/streams/:id/audio` (public — un élément `<audio>`
  ne peut pas poser d'en-tête `Authorization`). Le serveur enregistre un `Client`
  dans le `Hub`, envoie immédiatement le segment d'init en cache (pour les arrivants
  tardifs), puis écrit chaque chunk reçu via `http.Flusher` en transfert chunked,
  jusqu'à déconnexion (`ctx.Done`) — libération propre du channel et de la goroutine.
- **Comptage** : `POST /streams/:id/listen` et `/leave` enregistrent l'événement
  dans `listen_history` (join/leave) et alimentent les métriques Prometheus.

## Conséquences

### Positives
- **Simplicité + démo-abilité** : lisible par n'importe quel client HTTP (`curl`,
  `<audio>`, `just_audio`), aucune négociation ICE/SDP comme WebRTC.
- **Backpressure gérée** : `Hub.Broadcast` fait un envoi non bloquant (channel bufferisé) ;
  un auditeur lent voit ses paquets **droppés** sans bloquer le stream ni les autres.
- **Arrivants tardifs** : le cache du segment d'init WebM leur permet d'initialiser
  le décodeur du navigateur même en rejoignant en cours de route.
- **Découplé de la base** : le relais vit entièrement en mémoire dans le `Hub`
  (aucune écriture DB dans le chemin chaud).

### Négatives / limites
- **Pas de transcodage / qualité adaptative** (contrairement à HLS) — hors périmètre.
- **Latence** de l'ordre de la seconde (taille des chunks `MediaRecorder`).
- **`WriteTimeout` désactivé** côté serveur pour permettre les connexions longues
  (compensé par `ReadHeaderTimeout` + `IdleTimeout`).
- Pas de persistance du flux (pas de rediffusion) — live uniquement.

## Alternatives rejetées
- **WebRTC** : latence minimale mais complexité (signaling, STUN/TURN, SFU) hors budget.
- **HLS** : robuste et scalable mais latence de plusieurs secondes + pipeline de
  segmentation FFmpeg — trop lourd pour la démo.
- **WebSocket binaire** : équivalent fonctionnel, mais le HTTP chunked se branche
  directement sur un `<audio src>` sans code client de réassemblage.

## Vérification
Test d'intégration `handler.TestStreamingIntegration_BroadcastToListeners` :
un diffuseur pousse un chunk, deux auditeurs connectés le reçoivent tous les deux ;
`TestStreamingIntegration_Disconnect` prouve la libération des ressources à la
déconnexion ; `TestStreamHandler_Audio_LateJoiner` prouve la remise du segment d'init.
