# Charge, pprof, capacité et coûts

## Résultat de référence

Mesure effectuée le 31 juillet 2026 avec Go 1.25, Windows amd64, sur un
Intel Core i7-7660U 2,50 GHz (4 threads). Commande :

```bash
go test ./internal/infrastructure/streaming \
  -run '^$' -bench '^BenchmarkHubBroadcast$' \
  -benchmem -benchtime=2s
```

| Auditeurs | Temps/chunk 32 Kio | Débit logique | Mémoire/op | Allocations/op |
| ---: | ---: | ---: | ---: | ---: |
| 10 | 13,774 µs | 23,790 MB/s | 32,768 o | 1 |
| 100 | 19,213 µs | 170,548 MB/s | 32,768 o | 1 |
| 500 | 43,718 µs | 374,766 MB/s | 32,768 o | 1 |

Le débit est « logique » : Go compte les octets remis aux 500 channels, pas les
octets réellement sortis de la carte réseau. La preuve importante est
l’allocation constante : un seul clone immuable de 32 Kio par chunk, et non un
clone par auditeur. Le réseau/TLS doit être validé séparément par le scénario
k6 documenté dans `go/loadtest/README.md`.

## Budget réseau

Pour un encodage constant à 128 kbit/s :

| Auditeurs simultanés | Sortie requise | Données pour 1 h | Données si 24/7 pendant 30 j |
| ---: | ---: | ---: | ---: |
| 10 | 1,28 Mbit/s | 576 Mo | 0,415 To |
| 100 | 12,8 Mbit/s | 5,76 Go | 4,15 To |
| 500 | 64 Mbit/s | 28,8 Go | 20,74 To |

Formules :

```text
débit_sortie_bit_s = bitrate_audio_bit_s × auditeurs
volume_octets = débit_sortie_bit_s ÷ 8 × durée_secondes
auditeurs_max = bande_passante_utile_bit_s ÷ bitrate_audio_bit_s
```

Une marge de 30 % est réservée au TLS, aux retransmissions et aux autres
services. Avec une interface à 200 Mbit/s, le budget utile est donc 140 Mbit/s,
soit environ 1 093 auditeurs à 128 kbit/s ou 546 à 256 kbit/s. Cette limite
réseau arrive bien avant la limite du fan-out en mémoire mesurée ci-dessus.

## Comparatif de coût self-hosted

Exemple indicatif en région Paris, prix publics consultés le 31 juillet 2026,
hors taxes :

| Poste | Hypothèse | Coût mensuel |
| --- | --- | ---: |
| VM | Scaleway BASIC2-A2C-4G, 2 vCPU, 4 Gio, 200 Mbit/s | 16,79 € |
| IPv4 flexible | 0,004 €/h × 730 h | 2,92 € |
| Block Storage 5K | 40 Gio × 0,000130 €/Gio/h × 730 h | 3,80 € |
| API/Prometheus/Grafana | Conteneurs sur la même VM | inclus compute |
| Egress Instance | inclus dans le prix Instance selon la grille | 0 € |
| **Total infrastructure applicative** | hors BDD/Supabase, sauvegardes et TVA | **23,51 €** |

Sources officielles : [tarifs Instances Scaleway](https://www.scaleway.com/en/pricing/virtual-instances/)
et [tarifs Block Storage Scaleway](https://www.scaleway.com/en/pricing/storage/).
Les tarifs changent : revalider le devis dans la console avant une soutenance
ou une mise en production.

Ce tableau reste un comparatif de capacité et non l’architecture actuellement
déployée : la production utilise Render. Son coût réel doit être relevé dans le
dashboard et sur la [grille officielle Render](https://render.com/pricing), en
ajoutant notamment le volume sortant calculé ci-dessus. À 500 auditeurs
permanents en 128 kbit/s, les 20,74 To/mois rendent le coût de bande passante
déterminant.

En self-hosted, le coût marginal n’est pas proportionnel aux auditeurs tant que la bande
passante de l’Instance est incluse et reste sous 200 Mbit/s. Au-delà, il faut
monter de gamme, répartir les streams ou placer la diffusion derrière un
service média/CDN ; recalculer alors le coût d’egress réel.

## Protocole de preuve k6 + pprof

Pour chaque palier 10/100/500 :

1. démarrer une source `ffmpeg -re` à bitrate connu ;
2. lancer `go/loadtest/stream.js` pendant au moins 60 s ;
3. capturer 30 s de CPU, puis heap et goroutines ;
4. exporter les métriques Grafana et la sortie k6 ;
5. arrêter la source et attendre `STREAM_IDLE_TIMEOUT` ;
6. vérifier le retour des gauges auditeurs/publisher à zéro et l’absence de
   goroutines résiduelles.

Les profils ne sont pas committés. Le tableau ci-dessous est à remplir avec une
preuve exécutée sur l’hôte de production, car un benchmark local ne peut pas
inventer CPU, RAM, TLS ou latence réseau de cette machine.

| Palier | CPU p95 | RSS max | Drops | k6 checks | Goroutines après repos |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 10 | à mesurer | à mesurer | à mesurer | à mesurer | à mesurer |
| 100 | à mesurer | à mesurer | à mesurer | à mesurer | à mesurer |
| 500 | à mesurer | à mesurer | à mesurer | à mesurer | à mesurer |

## Seuils de décision

- drops audio > 1 % pendant 2 min : incident, vérifier clients lents et CPU ;
- bande passante > 70 % soutenue : augmenter la capacité avant le palier suivant ;
- heap ou goroutines ne revenant pas au niveau de repos : suspicion de fuite ;
- CPU > 70 % soutenu : profiler avant de scaler ;
- une seconde réplique n’est pas sûre sans routage par `stream_id`, car le Hub
  n’est pas distribué.
