# Déploiement Docker Hub vers Render

La production StreamPulse utilise exclusivement Render. GitHub Actions construit
l’image de l’API Go, la publie sur Docker Hub puis demande à l’API Render de
déployer le tag immuable correspondant au SHA Git. Render exécute le conteneur,
fournit le reverse proxy public et gère automatiquement HTTPS.

## 1. Créer le dépôt Docker Hub

Créer un dépôt nommé `streampulse-api`, public ou privé. Créer ensuite un
access token Docker Hub avec droit d’écriture ; ne jamais utiliser le mot de
passe du compte dans GitHub.

Dans GitHub, ouvrir `Settings > Secrets and variables > Actions`.

Variables :

| Nom | Exemple |
| --- | --- |
| `DOCKERHUB_USERNAME` | `mon-compte` |
| `DOCKERHUB_REPOSITORY` | `mon-compte/streampulse-api` |
| `API_DOMAIN` | `streampulse-api.onrender.com` ou domaine personnalisé |
| `RENDER_SERVICE_ID` | identifiant `srv-...` du Web Service |

Secrets :

| Nom | Usage |
| --- | --- |
| `DOCKERHUB_TOKEN` | Publication de l’image |
| `RENDER_API_KEY` | déploiement, lecture du digest et rollback via l’API Render |

## 2. Créer le Web Service Render

Dans Render, créer un Web Service à partir d’une image existante :

```text
docker.io/mon-compte/streampulse-api:<sha-git>
```

Si le dépôt Docker Hub est privé, enregistrer les identifiants du registre dans
Render. Configurer :

- health check : `/health` ;
- auto-deploy Render : désactivé, car GitHub Actions déploie le tag SHA exact ;
- une seule instance tant que le Hub audio reste en mémoire ;
- région proche de la base PostgreSQL/Supabase.

Ne pas fixer `PORT` : Render l’injecte et l’API lit directement cette variable.

## 3. Variables et Secret Files Render

Variables :

| Nom | Valeur |
| --- | --- |
| `APP_ENV` | `production` |
| `SUPABASE_DB_URL` | URL PostgreSQL avec `sslmode=require` |
| `CORS_ALLOWED_ORIGINS` | origine exacte du frontend |
| `JWT_PRIVATE_KEY_PATH` | `/etc/secrets/private.pem` |
| `JWT_PUBLIC_KEY_PATH` | `/etc/secrets/public.pem` |
| `METRICS_BEARER_TOKEN` | token aléatoire long |
| `PPROF_ENABLED` | `false` |
| `STREAM_MAX_DURATION` | `6h` |
| `STREAM_IDLE_TIMEOUT` | `30s` |
| `STREAM_WRITE_TIMEOUT` | `10s` |

Secret Files :

| Nom | Contenu |
| --- | --- |
| `private.pem` | clé privée JWT RS256 |
| `public.pem` | clé publique JWT RS256 |

Les clés peuvent être générées localement :

```bash
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem
```

Ne jamais committer ces deux fichiers.

## 4. Autoriser le déploiement prouvé

Créer une clé API Render limitée au workspace concerné, l’enregistrer dans le
secret GitHub `RENDER_API_KEY`, puis copier l’identifiant `srv-...` du service
dans la variable GitHub `RENDER_SERVICE_ID`. La clé est nécessaire pour lire le
digest réellement résolu par Render et pour exercer un rollback ; un deploy hook
seul ne fournit pas ces preuves.

Le workflow `.github/workflows/deploy.yml` suit cet ordre :

```text
push main
  -> tests Go/Flutter
  -> build go/Dockerfile
  -> push Docker Hub :<SHA Git>
  -> push Docker Hub :latest
  -> POST API Render avec docker.io/...:<SHA Git>
  -> attendre le statut live
  -> résoudre dans l'index OCI BuildKit le manifeste Linux/amd64 déployable
  -> comparer ce digest de manifeste au digest résolu par Render
  -> prouver health, endpoint métier public, redirection HTTP et certificat TLS
  -> en cas d'échec critique, restaurer et prouver la version live précédente
  -> prouver Prometheus, Loki et Tempo sans rollback applicatif sur erreur de lecture
  -> publier l'artefact production-evidence-<SHA Git>
```

La version live et son digest sont mémorisés avant le déploiement. Un échec
du statut Render, du digest, de `/health`, du smoke test `GET /api/v1/streams`,
de la redirection HTTPS ou du certificat déclenche le rollback automatique vers
cette version. La restauration est elle-même prouvée et le workflow reste en
échec afin de bloquer la livraison. Une erreur de lecture Grafana ne déclenche
pas ce rollback : elle invalide la preuve d'observabilité sans conclure que
l'application est défectueuse.

Lors du tout premier déploiement, l'API Render ne retourne encore aucune version
`live`. Ce cas n'est pas bloquant : le déploiement continue, mais aucun rollback
automatique n'est possible tant qu'une première version saine n'existe pas. Un
échec critique est alors signalé explicitement pour intervention manuelle.

Le job Render dépend du job Docker Hub : l’API Render n’est jamais appelée si
les tests, le build ou le push échouent. Les tags `v*` et les lancements manuels
publient l’image, mais seul un push sur `main` déclenche Render.

## 5. HTTPS et domaine

L’URL `onrender.com` fournie par Render est disponible en HTTPS. Pour un domaine
personnalisé :

1. ajouter le domaine dans les paramètres Render ;
2. créer les enregistrements DNS demandés par Render ;
3. attendre la validation et l’émission du certificat ;
4. mettre à jour `CORS_ALLOWED_ORIGINS` et `API_DOMAIN`.

Vérifications :

```bash
curl --fail --show-error https://streampulse-api.onrender.com/health
curl -I http://streampulse-api.onrender.com/health
```

Résultats attendus : `/health` renvoie 200 en HTTPS et HTTP redirige vers HTTPS.

## 6. Vérifier un déploiement

Dans GitHub Actions :

- `Verify release` doit réussir ;
- `Publish immutable image to Docker Hub` doit réussir ;
- `Deploy and prove Render production` doit réussir.

Télécharger l’artefact `production-evidence-<sha>`. Il contient la réponse Render
sanitisée (`image.ref` et `image.sha`), le `curl --verbose`, les en-têtes de
redirection, le certificat, sa chaîne et le JSON exact de `/health`. BuildKit
publie un index OCI qui contient le manifeste Linux/amd64 et les attestations
SBOM/provenance. Le job résout ce manifeste d'exécution puis échoue si son digest
diffère du digest tiré par Render.

## 7. Rollback

Chaque version reste disponible sur Docker Hub avec son SHA Git. Le workflow
manuel `Render rollback proof` demande l’identifiant d’un ancien déploiement et
la confirmation `ROLLBACK_AND_RESTORE`. Il :

1. mémorise la référence et le digest actuellement actifs ;
2. appelle `POST /v1/services/{serviceId}/rollback` ;
3. prouve le digest, le health check et TLS sur la version restaurée ;
4. redéploie automatiquement la référence qui était active avant le test ;
5. prouve cette restauration et publie les deux jeux d’artefacts.

Cette opération provoque volontairement un changement temporaire de production.
La lancer uniquement pendant une fenêtre annoncée et avec un ancien déploiement
connu comme sain.

## 8. Observabilité et pprof

Prometheus/Grafana restent disponibles dans la stack Docker locale. En
production, le Web Service API envoie directement les métriques et les traces
en OTLP vers Grafana Cloud, sans Background Worker Render payant. Les logs
partent directement de l’API vers Loki. `/metrics` reste disponible pour la
stack locale. Voir `docs/observability-rncp.md`.

pprof doit rester désactivé sur Render : les profils peuvent contenir des
informations internes et ne doivent pas être rendus publics.
