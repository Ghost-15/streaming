# Preuves de déploiement production — 2026-08-17

## Livraison vérifiée

- Commit `main` : `dc5d4e4bd17d98806e7273047c933f94da153f8b`
- Run de déploiement : [32017556247](https://github.com/Ghost-15/streaming/actions/runs/32017556247), succès
- URL de santé HTTPS : <https://streampulse-api-1jxg.onrender.com/health>
- Déploiement Render : `dep-da1dn75bedkc73cnmjl0`, état `live`
- Image : `docker.io/timosleboss/streampulse-api:dc5d4e4bd17d98806e7273047c933f94da153f8b`
- Digest du manifeste Linux/amd64 : `sha256:d741f210c44555a6ad6ebda4d91ffcfd732a79687e9f3963bb6e334601e90912`
- Redirection HTTP : `301` vers `https://streampulse-api-1jxg.onrender.com/health`
- Santé : `{"service":"streampulse-api","status":"ok"}`
- Smoke métier : `GET /api/v1/streams`, tableau JSON valide
- Certificat : `CN=onrender.com`, valide du 2026-07-24 au 2026-10-22, empreinte SHA-256 `04:1E:C4:C6:9F:66:77:19:7B:4E:F1:C0:67:42:11:64:F4:B2:69:DA:27:F2:F5:30:29:2C:AB:BB:05:1F:C1:B1`
- Observabilité : 4 échantillons Prometheus, 6 entrées Loki et 4 traces Tempo

## Rollback réellement exercé

- Run de rollback/restauration : [32018300239](https://github.com/Ghost-15/streaming/actions/runs/32018300239), succès
- Cible saine : déploiement `dep-d9t8d6e417fc73dph59g`, image `1188d466999b32ba805324262fa9ad439b4714c0`
- Rollback créé : `dep-da1dovrl550s73ff01n0`, trigger `rollback`, état `live`
- Digest rollback vérifié : `sha256:db3af481c8084ab2eda39c8fdb3cc36c18401796f91a4587019eddb2bdb4cdd9`
- Restauration créée : `dep-da1dp649v7es73b6eabg`, état `live`
- Image restaurée : `dc5d4e4bd17d98806e7273047c933f94da153f8b`
- Digest restauré vérifié : `sha256:d741f210c44555a6ad6ebda4d91ffcfd732a79687e9f3963bb6e334601e90912`
- Les phases rollback et restauration ont chacune validé le digest Render, `/health`, le smoke métier, la redirection HTTP→HTTPS et le certificat TLS.

## Archives

Les ZIP sont des copies exactes des artefacts GitHub Actions, conservés également par GitHub jusqu'au 2026-11-15.

| Fichier | Artefact GitHub | SHA-256 du ZIP |
| --- | --- | --- |
| `production-evidence-32017556247.zip` | `production-evidence-dc5d4e4bd17d98806e7273047c933f94da153f8b` (`9284280760`) | `63A485A758455D3C2B172B8442E7007D3A95D32DB4CC107C8E7F3C66079DFADB` |
| `rollback-evidence-32018300239.zip` | `rollback-evidence-32018300239` (`9284374844`) | `736CB2FD4E9EF54B33DD6E107476A142683164CA72CEB0E7C30C2CD952BE48C1` |

Le premier ZIP contient les réponses Render sanitisées, les preuves HTTPS/TLS, la redirection, la santé, le smoke métier et l'observabilité. Le second contient les mêmes contrôles pour le rollback puis pour la restauration.
