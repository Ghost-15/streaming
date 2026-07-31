# Mise en production HTTPS

Cette procédure déploie une API StreamPulse mono-nœud avec Caddy, certificats
ACME automatiques, Prometheus et Grafana. Elle ne requiert ni Kubernetes ni
Terraform.

## 1. Préparer l’hôte

Recommandation minimale pour 500 auditeurs à 128 kbit/s : Linux 64 bits,
2 vCPU, 4 Gio RAM, SSD 40 Gio et au moins 200 Mbit/s. Installer Docker Engine,
le plugin Compose, curl et tar.

Créer un utilisateur de déploiement non root autorisé à exécuter Docker, puis
un répertoire dédié :

```bash
sudo install -d -o streampulse -g streampulse /opt/streampulse
```

Firewall entrant :

- 22/TCP uniquement depuis les IP d’administration/du runner ;
- 80/TCP pour le challenge ACME et la redirection HTTPS ;
- 443/TCP et 443/UDP pour HTTPS/HTTP3 ;
- aucun accès public à 8080, 3000, 6060, 9090 ou PostgreSQL.

## 2. DNS

Créer deux enregistrements A/AAAA pointant vers l’hôte :

```text
api.example.com      -> IP du serveur
grafana.example.com  -> IP du serveur
```

Attendre leur propagation et vérifier avec `dig +short`. Caddy ne pourra
obtenir les certificats tant que les domaines ne résolvent pas vers l’hôte et
que les ports 80/443 ne sont pas joignables.

## 3. Configuration et secrets

Sur le serveur :

```bash
cd /opt/streampulse
cp .env.production.example .env.production
install -m 700 -d secrets
openssl genrsa -out secrets/private.pem 4096
openssl rsa -in secrets/private.pem -pubout -out secrets/public.pem
openssl rand -hex 32 > secrets/metrics_bearer_token
openssl rand -base64 32 > secrets/grafana_admin_password
chmod 600 .env.production secrets/*
```

Renseigner dans `.env.production` les vrais domaines, l’e-mail ACME,
`SUPABASE_DB_URL`, `CORS_ALLOWED_ORIGINS` et les métadonnées OTEL. Ne pas ajouter
de guillemets autour des valeurs.

Le serveur doit déjà pouvoir tirer l’image Docker Hub. Pour une image privée :

```bash
docker login
```

Utiliser un token Docker Hub limité à la lecture sur le serveur.

## 4. Premier déploiement manuel

```bash
cd /opt/streampulse
chmod +x deploy/deploy.sh
IMAGE_REPOSITORY=utilisateur-dockerhub/streampulse-api \
IMAGE_TAG=<sha-git-complet> \
./deploy/deploy.sh
```

Le script :

1. vérifie les fichiers requis ;
2. tire toutes les images ;
3. applique `compose.prod.yml` ;
4. attend jusqu’à 120 s le health-check HTTPS ;
5. écrit `.deployed-tag` seulement après succès ;
6. restaure automatiquement le tag précédent en cas d’échec.

## 5. Pipeline GitHub Actions

Dans `Settings > Secrets and variables > Actions`, créer au niveau du dépôt :

Variables :

- `DOCKERHUB_USERNAME` : nom du compte Docker Hub ;
- `DOCKERHUB_REPOSITORY` : par exemple `utilisateur/streampulse-api` ;
- `API_DOMAIN` : domaine API sans `https://` (requis uniquement pour le
  déploiement VPS ; le build de validation utilise sinon `api.example.com`) ;
- `ENABLE_VPS_DEPLOY` : `false` pour publier uniquement sur Docker Hub, ou
  `true` pour exécuter ensuite le déploiement SSH.

Secret :

- `DOCKERHUB_TOKEN` : access token Docker Hub avec droit d’écriture.
- `RENDER_DEPLOY_HOOK_URL` : URL secrète du deploy hook créée dans Render.

### Déploiement automatique Render

Configurer dans Render un Web Service basé sur l’image Docker :

```text
utilisateur-dockerhub/streampulse-api:latest
```

Renseigner dans Render :

| Type | Nom | Valeur |
| --- | --- | --- |
| Variable | `APP_ENV` | `production` |
| Variable | `SUPABASE_DB_URL` | URL PostgreSQL avec TLS |
| Variable | `CORS_ALLOWED_ORIGINS` | domaine exact du frontend |
| Variable | `JWT_PRIVATE_KEY_PATH` | `/etc/secrets/private.pem` |
| Variable | `JWT_PUBLIC_KEY_PATH` | `/etc/secrets/public.pem` |
| Variable | `METRICS_BEARER_TOKEN` | token aléatoire long |
| Variable | `PPROF_ENABLED` | `false` |
| Secret File | `private.pem` | clé JWT privée |
| Secret File | `public.pem` | clé JWT publique |

Ne pas définir une valeur fixe pour `PORT` : Render l’injecte et l’API l’utilise
directement. Configurer le health-check Render sur `/health`.

À chaque push sur `main`, le workflow :

1. vérifie Go et Flutter ;
2. construit l’image Go ;
3. publie les tags SHA et `latest` sur Docker Hub ;
4. appelle `RENDER_DEPLOY_HOOK_URL` uniquement après la publication réussie.

Le hook est un secret : ne jamais le placer dans un fichier `.env` committé.
Pour un dépôt Docker Hub privé, configurer aussi les identifiants du registre
dans Render.

Si `ENABLE_VPS_DEPLOY=true`, créer aussi un environnement GitHub `production`
avec approbation manuelle et les secrets suivants :

- `PROD_HOST` : IP/nom SSH ;
- `PROD_USER` : utilisateur de déploiement ;
- `PROD_PATH` : `/opt/streampulse` ;
- `PROD_SSH_PRIVATE_KEY` : clé Ed25519 dédiée ;
- `PROD_KNOWN_HOSTS` : ligne de clé hôte vérifiée hors bande.

Le workflow `.github/workflows/deploy.yml` s’exécute sur `main`, sur un tag
`v*` ou manuellement. Il teste Go avec race detector, analyse/teste/build
Flutter et publie l’image Docker Hub sous le SHA complet avec SBOM et
provenance. Render est déclenché uniquement par `main`. Le transfert SSH et
l’appel du script VPS ne s’exécutent que lorsque `ENABLE_VPS_DEPLOY=true`.

## 6. Vérifications HTTPS

```bash
curl --fail --show-error https://api.example.com/health
curl -I http://api.example.com/health
openssl s_client -connect api.example.com:443 -servername api.example.com </dev/null
docker compose --env-file .env.production -f compose.prod.yml ps
docker compose --env-file .env.production -f compose.prod.yml logs --tail=100 caddy api
```

Résultats attendus :

- `/health` renvoie 200 en HTTPS ;
- HTTP redirige vers HTTPS ;
- le certificat correspond au domaine et sa chaîne est valide ;
- `/metrics` et `/debug/pprof/` renvoient 404 depuis Internet ;
- Grafana exige un login et est joignable uniquement en HTTPS.

Caddy conserve certificats et état ACME dans `caddy_data`. Le renouvellement est
automatique ; surveiller les logs Caddy et ne jamais supprimer ce volume durant
un déploiement.

## 7. Rollback et restauration

Rollback automatique : le script redéploie `.deployed-tag` si le nouveau
health-check échoue.

Rollback manuel :

```bash
cd /opt/streampulse
IMAGE_REPOSITORY=utilisateur-dockerhub/streampulse-api \
IMAGE_TAG=<ancien-sha> \
./deploy/deploy.sh
```

Sauvegarder régulièrement :

- base PostgreSQL/Supabase avec la stratégie du fournisseur ;
- volumes `grafana_data`, `prometheus_data` et `caddy_data` ;
- `.env.production` et `secrets/` dans un coffre chiffré, jamais dans Git.

Tester la restauration sur un hôte isolé. Les métriques ne sont pas une
sauvegarde métier.

## 8. Exploitation et pprof

Le dashboard Grafana surveille débit, drops, audience, 5xx et latence. pprof est
désactivé dans `compose.prod.yml`. En incident exceptionnel :

1. activer temporairement `PPROF_ENABLED=true` avec
   `PPROF_ADDR=127.0.0.1:6060` (aucun port Docker publié) ;
2. récupérer l’identifiant du conteneur API ;
3. lancer un conteneur curl éphémère dans le même namespace réseau :

   ```bash
   api_id="$(docker compose --env-file .env.production -f compose.prod.yml ps -q api)"
   docker run --rm --network "container:$api_id" curlimages/curl:8.15.0 \
     --fail http://127.0.0.1:6060/debug/pprof/profile?seconds=30 \
     > cpu.pb.gz
   ```

4. analyser le fichier hors du serveur avec `go tool pprof`, puis désactiver
   pprof et redéployer.

Ne jamais ajouter pprof au Caddyfile : ses profils peuvent exposer des données
internes.
