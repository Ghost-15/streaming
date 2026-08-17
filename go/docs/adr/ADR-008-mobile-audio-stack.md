# ADR-008 — Stack audio mobile (iOS/Android) et périmètre codec

## Statut
Accepté — Sprint Final. Fait passer le client Flutter de **Web only** à
**multiplateforme (Web + iOS/Android)** pour l'écoute en direct.

## Contexte
Le sujet impose un frontend **mobile multiplateforme** avec lecture en
arrière-plan, gestion des interruptions et fluidité. L'implémentation initiale
était **Web only** : le cœur audio (écoute via `MediaSource`, capture via
`MediaRecorder`) repose sur des API navigateur (`dart:js_interop`,
`package:web`) indisponibles sur iOS/Android, où un stub renvoyait une erreur.

Il fallait fournir une lecture audio **native** sur mobile sans réécrire l'UI
ni le backend.

## Décision

### Seam d'imports conditionnels
Chaque notifier audio est exporté selon la plateforme, sans toucher au reste de
l'app (UI, providers, repositories restent communs) :
```
export 'audio_notifier_stub.dart'
    if (dart.library.io) 'audio_notifier_io.dart'   // iOS/Android/desktop
    if (dart.library.html) 'audio_notifier_web.dart'; // Web
```
`audio_notifier_io.dart` expose **exactement la même interface publique** que la
version Web (`AudioNotifier` + `AudioPlaybackState` + mêmes méthodes/getters).

### Écoute mobile — `just_audio`
`just_audio` lit le flux HTTP live `GET /streams/:id/audio` (buffering réseau
géré nativement, décodé par ExoPlayer sur Android). États mappés sur
`AudioPlaybackState` via `playerStateStream`.

### Arrière-plan & contrôles système — `just_audio_background`
`JustAudioBackground.init()` au démarrage (mobile uniquement, via un bootstrap
conditionnel) + un `MediaItem` sur la source → **lecture en arrière-plan**,
notification média et **contrôles écran verrouillé** (play/pause/stop, titre du
live + nom du diffuseur). Côté Android : foreground service `mediaPlayback` +
permissions `FOREGROUND_SERVICE*` / `WAKE_LOCK`. Côté iOS : `UIBackgroundModes:
audio`.

### Interruptions & focus audio — `audio_session`
`handleInterruptions: false` sur le player pour piloter nous-mêmes le focus :
- **Ducking** : une alerte système transitoire (notification, GPS) → le volume
  **baisse** (`_duckVolumeFactor`) puis **remonte** à la fin.
- **Pause** : un appel entrant ou une autre app prend le focus exclusif → pause,
  puis **reprise** automatique à la fin si l'utilisateur n'avait pas mis en pause.
- **Casque débranché / Bluetooth coupé** (`becomingNoisyEventStream`) → pause,
  comme une app musicale.

### Reconnexion réseau — `connectivity_plus`
Une coupure du flux (erreur `playbackEventStream`) déclenche une reconnexion en
**backoff** (0,5 → 8 s, max 8 tentatives). Le retour de connectivité (Wi-Fi ↔
mobile) relance immédiatement depuis le live edge.

### Diffuseur : reste Web
La capture micro navigateur (`MediaRecorder`) produit du WebM/Opus. Sur mobile,
le diffuseur **dégrade proprement** (message explicite) ; la diffusion reste sur
le client Web. Split assumé : **auditeur = mobile**, **diffuseur = Web**.

## Conséquences

### Positives
- Écoute native **Android** avec arrière-plan, contrôles système et interruptions
  → couvre les exigences « lecture en arrière-plan » et « gestion des interruptions ».
- **Aucune régression Web** : la version `_web.dart` est intacte, le backend n'est
  pas touché (il sert déjà `/audio` + `/audio/ws`).
- UI et state management (Provider) inchangés.

### Négatives / limites — le mur codec iOS
Le diffuseur navigateur émet du **WebM/Opus**. **AVPlayer (iOS) ne décode pas le
WebM** : l'écoute live d'un flux poussé depuis le Web est donc **limitée sur iOS**.
L'app iOS **compile, s'installe et se navigue**, mais la lecture du direct WebM
n'est pas garantie nativement. Lever cette limite demanderait soit un
**transcodage serveur HLS/AAC** (latence + pipeline FFmpeg, hors périmètre), soit
un **décodeur Opus côté client** (complexe). Documenté comme limite connue.

## Alternatives rejetées
- **Transcodage serveur HLS/AAC** : rend iOS pleinement compatible mais ajoute de
  la latence et un pipeline FFmpeg lourd — hors budget du sprint.
- **Décodeur Opus client** (`flutter_soloud` / `opus_dart`, sortie PCM custom) :
  débloque iOS mais complexité et risque élevés.
- **Diffusion micro mobile** : asymétrie de codec (Android enregistre Opus, iOS
  AAC/CAF ≠ WebM navigateur) — reportée, diffuseur maintenu sur Web.

## Vérification
- `flutter analyze` : 0 issue ; `flutter test` : suite verte (le smoke test
  instancie le player io sans planter).
- Builds ciblés `flutter build apk` / `ios` / `web` (validation sur émulateur
  Android : écoute d'un live poussé depuis le Web + arrière-plan + ducking +
  pause sur appel + pause casque + reprise réseau).
