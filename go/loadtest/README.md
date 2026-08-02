# Test de charge audio et pprof

## 1. Benchmark reproductible du Hub

Depuis la racine :

```bash
make load-bench
```

Le benchmark diffuse des chunks de 32 Kio vers 10, 100 et 500 auditeurs, vide
chaque channel à chaque itération et publie `ns/op`, débit logique, allocations
et octets alloués. Il ne mesure pas la pile TLS ou le réseau.

## 2. Test k6 de bout en bout

Prérequis : k6, FFmpeg, curl, une API HTTPS, un stream live et deux JWT (rôle
Diffuseur pour la source, User pour les auditeurs).

Terminal A :

```bash
ffmpeg -re -stream_loop -1 -i sample.mp3 -codec copy -f mp3 - \
  | curl --no-buffer -X PUT \
      -H "Authorization: Bearer $BROADCASTER_TOKEN" \
      -H "Content-Type: audio/mpeg" \
      --data-binary @- \
      "$BASE_URL/streams/$STREAM_ID/audio"
```

Terminal B :

```bash
cd go
k6 run \
  -e LISTENERS=10 \
  -e BASE_URL="$BASE_URL" \
  -e STREAM_ID="$STREAM_ID" \
  -e LISTENER_TOKEN="$LISTENER_TOKEN" \
  loadtest/stream.js
```

Répéter avec `LISTENERS=100`, puis `500`. Arrêter FFmpeg après au moins 60 s ;
les requêtes se terminent après le timeout d’inactivité du serveur. Conserver
la sortie k6 et exporter simultanément les panels Grafana.

Critères :

- plus de 99 % de checks réussis ;
- moins de 1 % de requêtes échouées ;
- `streampulse_audio_dropped_chunks_total` reste à zéro ou sous 1 % ;
- pas de croissance monotone des goroutines/du heap après retour au repos ;
- débit sortant proche de `bitrate × auditeurs`.

## 3. Capturer pprof pendant k6

En local, Compose publie pprof uniquement sur `127.0.0.1:6060`.

PowerShell :

```powershell
cd go
.\loadtest\capture-pprof.ps1 -CpuSeconds 30
go tool pprof -http=:0 .\loadtest\results\cpu.pb.gz
```

Linux/macOS :

```bash
cd go
PPROF_BASE_URL=http://127.0.0.1:6060 sh loadtest/capture-pprof.sh
go tool pprof -http=:0 loadtest/results/cpu.pb.gz
```

En production Render, pprof reste désactivé et ne doit jamais être exposé par
le service public.

Les profils binaires sont ignorés par Git, car ils peuvent contenir des
informations sensibles. Le rapport synthétique et les commandes, eux, restent
versionnés dans `docs/performance-couts.md`.
