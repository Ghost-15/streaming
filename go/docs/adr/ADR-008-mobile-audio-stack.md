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
le client Web.

## Conséquences

### Positives
- **Écoute native Android** (just_audio + background) **et iOS** (media_kit /
  libmpv, qui décode le WebM/Opus qu'AVPlayer refuse) — avec arrière-plan,
  contrôles système et interruptions → couvre « lecture en arrière-plan » et
  « gestion des interruptions ». Le moteur est choisi au runtime derrière une
  même interface `AudioNotifier`, sans toucher l'UI.
- **Diffuseur mobile** (Android/iOS) : capture micro via `record`, poussée sur
  `POST /streams/:id/push` au même rythme (~500 ms) que le MediaRecorder web.
- **Aucune régression Web** : la version `_web.dart` et le backend sont intacts
  (il sert déjà `/audio` + `/audio/ws`).

### Négatives / limites
- **Codec du diffuseur mobile** : `record` produit de l'**AAC** (ADTS), pas du
  WebM/Opus. Un live poussé depuis un mobile est donc décodable par les
  **auditeurs mobiles** (just_audio / media_kit) mais **pas par un `<audio>`
  navigateur**. Le diffuseur Web reste la source universelle.
- **Latence** ~1 s (taille des chunks), comme sur le web.
- Les builds iOS/Android natifs ne se valident que sur **appareil réel + Mac/Xcode**.

## Alternatives rejetées
- **Transcodage serveur HLS/AAC pour iOS** : rend iOS universel mais ajoute de la
  latence + un pipeline FFmpeg lourd dans le backend de **prod** — écarté au profit
  du décodage client `media_kit`, sans aucun changement backend.
- **just_audio sur iOS** : AVPlayer ne décode ni WebM ni Opus → inutilisable pour
  ce flux ; d'où media_kit sur iOS, just_audio conservé sur Android (validé).

## Vérification
- `flutter analyze` : 0 issue ; `flutter test` : 50 tests verts ;
  `flutter build web` : OK (aucune régression) ; `flutter build apk` : OK.
- Validation runtime à faire sur **appareil réel** (Android + iPhone via Mac) :
  écoute d'un live + arrière-plan + ducking + pause appel + pause casque + reprise
  réseau, et diffusion micro mobile → auditeur mobile.
