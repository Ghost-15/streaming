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

Prérequis : k6, FFmpeg, curl, une API HTTPS, un stream live et un JWT Diffuseur
pour la source. Le scénario utilise par défaut la route publique `/audio`, ce
qui représente le lecteur Web réel et évite de réutiliser artificiellement un
seul utilisateur soumis au rate limit sur 500 connexions. Pour tester la route
authentifiée `/listen`, fournir un JWT User avec `LISTENER_TOKEN`.

Les connexions sont établies sur une rampe déterministe de 10 secondes afin de
ne pas confondre le backlog TCP de la machine cliente avec la capacité de
diffusion. Toutes restent ouvertes : le palier demandé est donc bien atteint
simultanément pendant le flux. La rampe peut être ajustée avec
`START_RAMP_SECONDS` pour un environnement de production documenté.

Terminal A :

```bash
ffmpeg -re -stream_loop -1 -i sample.mp3 -codec copy -f mp3 - \
  | curl --no-buffer -X PUT \
      -H "Authorization: Bearer $BROADCASTER_TOKEN" \
      -H "Content-Type: audio/mpeg" \
      --data-binary @- \
      "$BASE_URL/streams/$STREAM_ID/audio"
```

Sans fichier audio, `loadtest/publisher.js` produit un flux binaire déterministe
au débit par défaut de 128 kbit/s, suffisant pour mesurer le transport (il n'est
pas destiné à être décodé par un lecteur) :

```bash
k6 run \
  -e BASE_URL="$BASE_URL" \
  -e STREAM_ID="$STREAM_ID" \
  -e STREAM_SESSION_ID="$STREAM_SESSION_ID" \
  -e BROADCASTER_TOKEN="$BROADCASTER_TOKEN" \
  loadtest/publisher.js
```

Terminal B :

```bash
cd go
k6 run \
  -e LISTENERS=10 \
  -e BASE_URL="$BASE_URL" \
  -e STREAM_ID="$STREAM_ID" \
  loadtest/stream.js
```

Sous PowerShell, le runner collecte simultanément k6, les métriques runtime et,
si son URL loopback est fournie, pprof :

```powershell
cd go
.\loadtest\run-tier.ps1 `
  -Listeners 10 `
  -BaseUrl $env:BASE_URL `
  -StreamId $env:STREAM_ID `
  -MetricsBearerToken $env:METRICS_BEARER_TOKEN `
  -PprofBaseUrl http://127.0.0.1:6060
```

Répéter avec `-Listeners 100`, puis `500`. Arrêter FFmpeg après au moins 60 s ;
les requêtes se terminent après le timeout d’inactivité du serveur. Le runner
attend ensuite le repos, calcule CPU p95, RSS max, taux de drops, checks k6 et
goroutines, puis génère `loadtest/results/summary.md`.

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
.\loadtest\capture-pprof.ps1 -CpuSeconds 30 -Prefix tier-10
pprof -http=:0 .\loadtest\results\tier-10-cpu.pb.gz
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
informations sensibles. Le script produit aussi des rapports pprof `cpu-top`
et `heap-top` dont le chemin de workspace est anonymisé. Ces rapports et les
sorties JSON/CSV/log/Markdown de `loadtest/results/` sont versionnables et
constituent la preuve reproductible.
