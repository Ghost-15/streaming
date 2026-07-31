# Déploiement Docker Hub vers Render

La production StreamPulse utilise exclusivement Render. GitHub Actions construit
l’image de l’API Go, la publie sur Docker Hub puis appelle un deploy hook
Render. Render exécute le conteneur, fournit le reverse proxy public et gère
automatiquement HTTPS.

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

Secrets :

| Nom | Usage |
| --- | --- |
| `DOCKERHUB_TOKEN` | Publication de l’image |
| `RENDER_DEPLOY_HOOK_URL` | Déclenchement du déploiement Render |

## 2. Créer le Web Service Render

Dans Render, créer un Web Service à partir d’une image existante :

```text
docker.io/mon-compte/streampulse-api:latest
```

Si le dépôt Docker Hub est privé, enregistrer les identifiants du registre dans
Render. Configurer :

- health check : `/health` ;
- auto-deploy Render depuis Git : désactivé, car le deploy hook est utilisé ;
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

## 4. Créer le deploy hook

Dans les paramètres du service Render, créer/copier le deploy hook puis
l’enregistrer dans le secret GitHub `RENDER_DEPLOY_HOOK_URL`.

Le workflow `.github/workflows/deploy.yml` suit cet ordre :

```text
push main
  -> tests Go/Flutter
  -> build go/Dockerfile
  -> push Docker Hub :<SHA Git>
  -> push Docker Hub :latest
  -> POST RENDER_DEPLOY_HOOK_URL
  -> Render tire :latest et remplace le service
```

Le job Render dépend du job Docker Hub : le hook n’est jamais appelé si les
tests, le build ou le push échouent. Les tags `v*` et les lancements manuels
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
- `Trigger Render deployment` doit réussir.

Dans Docker Hub, vérifier les tags `latest` et le SHA complet. Dans Render,
contrôler que le digest tiré correspond à la nouvelle image, puis consulter les
logs et `/health`.

## 7. Rollback

Chaque version reste disponible sur Docker Hub avec son SHA Git. Pour revenir à
une version :

1. sélectionner dans Render l’image
   `mon-compte/streampulse-api:<ancien-sha>` ;
2. lancer un déploiement manuel ;
3. vérifier `/health` et les logs ;
4. remettre `latest` après correction si nécessaire.

Le tag SHA est la preuve immuable ; `latest` sert uniquement au déploiement
automatique courant.

## 8. Observabilité et pprof

Prometheus/Grafana restent disponibles dans la stack Docker locale. Pour la
production, exporter les métriques vers un service compatible ou utiliser
Grafana Cloud sans exposer publiquement `/metrics`.

pprof doit rester désactivé sur Render : les profils peuvent contenir des
informations internes et ne doivent pas être rendus publics.
