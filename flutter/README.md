# StreamPulse Flutter Web

Client Web du direct audio StreamPulse. Le diffuseur capture le microphone avec
`MediaRecorder` (WebM/Opus) et envoie des blobs ordonnés à l'API Go. Le lecteur
consomme la réponse HTTP chunked avec `MediaSource` et expose les commandes
système via Media Session.

## Vérification

```powershell
flutter pub get
flutter analyze
flutter test
flutter build web --release
```

Les implémentations Web sont chargées par exports conditionnels. Les tests VM
utilisent des notifiers de repli ; le build Web valide les interops navigateur
réels. Chrome/Edge est recommandé pour WebM/Opus + MediaSource. Firefox peut
lire le flux mais ne fournit pas toujours les commandes Media Session.
