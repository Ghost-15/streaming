# Cahier des charges — StreamPulse

## 1. Présentation du projet

### 1.1 Contexte

StreamPulse est une plateforme de streaming audio en direct développée dans le cadre du
projet PEC2. Elle permet à des utilisateurs de diffuser en live depuis un microphone et
à des auditeurs d'écouter ces diffusions en temps réel depuis un navigateur web ou une
application mobile (iOS et Android).

Le projet couvre l'intégralité de la chaîne : backend API Go, application Flutter
multiplateforme, base de données PostgreSQL avec politiques de sécurité par ligne (RLS),
pipeline CI/CD, déploiement HTTPS sur Render et observabilité Prometheus/Grafana.

### 1.2 Objectifs

- Diffuser un flux audio en temps réel d'un émetteur vers plusieurs auditeurs.
- Gérer les comptes utilisateurs avec des rôles distincts (auditeur, diffuseur, admin).
- Permettre à chaque utilisateur de construire une bibliothèque personnelle (favoris,
  playlists).
- Alimenter un moteur de recommandations basé sur l'historique d'écoute.
- Respecter les exigences RGPD sur la collecte et la conservation des données.
- Assurer une fluidité d'interface à 60 FPS et une capacité à 500 auditeurs simultanés.

### 1.3 Périmètre

| Inclus | Hors périmètre |
| --- | --- |
| Streaming audio live WebM/Opus | Podcasts / fichiers pré-enregistrés |
| Application Flutter web + mobile (iOS, Android) | Application desktop native |
| API REST Go avec OpenAPI documenté | API GraphQL ou WebSocket temps réel |
| Gestion des comptes et rôles | Monétisation / paiement |
| Bibliothèque : favoris et playlists | Partage social externe |
| Moteur de recommandations | Algorithme ML avancé |
| Observabilité Prometheus / Grafana | APM cloud tiers payant |
| Déploiement Render + Docker Hub | Déploiement Kubernetes multi-région |

---

## 2. Acteurs et rôles

| Rôle | Description | Droits principaux |
| --- | --- | --- |
| **Anonyme** | Visiteur non authentifié | Inscription uniquement |
| **Auditeur** (`user`) | Utilisateur inscrit | Écouter les streams, gérer favoris et playlists |
| **Diffuseur** (`diffuseur`) | Auditeur avec droit de diffusion | Créer, démarrer, arrêter et supprimer ses streams |
| **Administrateur** (`admin`) | Gestionnaire de la plateforme | Tous les droits + gestion des comptes et statistiques |

Les rôles sont stockés dans la table `users.role` (ENUM PostgreSQL) et vérifiés à chaque
requête par le middleware RBAC du backend Go.

---

## 3. Spécifications fonctionnelles

### 3.1 Gestion des comptes

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Inscription | Créer un compte avec email, prénom, nom et mot de passe | Anonyme |
| Connexion | S'authentifier et recevoir un JWT RS256 | Tous |
| Déconnexion | Invalider la session côté client | Tous connectés |
| Consulter son profil | Accéder à ses informations de compte | Tous connectés |
| Modifier son mot de passe | Mettre à jour le mot de passe hashé | Tous connectés |

### 3.2 Diffusion live

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Créer un stream | Déclarer un nouveau canal de diffusion avec un titre | Diffuseur |
| Démarrer la diffusion | Activer le stream et envoyer les chunks audio (WebM/Opus) | Diffuseur |
| Arrêter la diffusion | Mettre fin au live (`status = ended`) | Diffuseur |
| Reprendre un stream | Réactiver un canal existant avec un nouvel `active_session_id` | Diffuseur |
| Supprimer un stream | Effacer définitivement un canal de diffusion | Diffuseur |
| Voir ses streams | Consulter la liste de ses canaux passés et actifs | Diffuseur |

Le protocole audio :

1. Le diffuseur crée un stream via `POST /api/v1/streams`.
2. Il envoie les blobs WebM/Opus de `MediaRecorder` vers `POST /api/v1/streams/{id}/push`
   (clients continus : `PUT /api/v1/streams/{id}/audio`).
3. Les auditeurs consomment `GET /api/v1/streams/{id}/audio` via une réponse HTTP
   incrémentale injectée dans l'API `MediaSource`.
4. L'arrêt via `PUT /api/v1/streams/{id}/stop` ou une déconnexion TCP libère les
   ressources (goroutines, channels, compteurs).

### 3.3 Écoute

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Lister les streams actifs | Voir tous les streams en cours sur l'écran d'accueil | Tous connectés |
| Rejoindre un stream | Lancer la lecture d'un live (event `join` enregistré) | Auditeur |
| Quitter un stream | Arrêter la lecture (event `leave` enregistré) | Auditeur |
| Contrôles de lecture | Lecture / pause / volume / arrêt | Auditeur |
| MiniPlayer persistant | Lecteur réduit visible en navigation entre écrans | Auditeur |

### 3.4 Bibliothèque personnelle

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Ajouter aux favoris | Marquer un stream comme favori | Auditeur |
| Retirer des favoris | Supprimer un favori | Auditeur |
| Créer une playlist | Créer une playlist nommée | Auditeur |
| Ajouter à une playlist | Ajouter un stream à une playlist existante | Auditeur |
| Consulter la bibliothèque | Voir ses playlists et favoris | Auditeur |
| File de lecture (queue) | Playlist spéciale `is_queue = true` pour la lecture enchaînée | Auditeur |

### 3.5 Recommandations

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Streams suggérés | Proposer des streams basés sur l'historique d'écoute | Auditeur |

L'historique est alimenté par les events `join` et `leave` dans `listen_history`.
Le moteur filtre les streams déjà écoutés et favorise les diffuseurs fréquents.

### 3.6 Administration

| Fonctionnalité | Description | Rôle |
| --- | --- | --- |
| Lister les utilisateurs | Voir tous les comptes avec leur rôle et leur statut | Admin |
| Modifier le rôle d'un compte | Passer un compte de `user` à `diffuseur` ou inversement | Admin |
| Suspendre un compte | Bloquer un utilisateur (`suspended_at` non null) | Admin |
| Réactiver un compte | Lever une suspension | Admin |
| Statistiques plateforme | Nombre d'utilisateurs, streams actifs, historique global | Admin |

---

## 4. Architecture technique

### 4.1 Stack

| Couche | Technologie | Version |
| --- | --- | --- |
| Application mobile et web | Flutter | 3.48+ (channel beta) |
| Backend API | Go, Clean Architecture | 1.26 |
| Base de données | PostgreSQL (Supabase) | 15 |
| Authentification | JWT RS256 | — |
| Stockage fichiers | Supabase Storage (bucket audio) | — |
| Observabilité | Prometheus + Grafana | — |
| CI/CD | GitHub Actions | — |
| Hébergement | Render (API) + Docker Hub (image) | — |

### 4.2 Organisation du code

```
streaming/
├── go/                        Backend Go (Clean Architecture)
│   ├── cmd/server/            Point d'entrée, configuration
│   ├── internal/
│   │   ├── entity/            Entités métier (User, Stream, Playlist…)
│   │   ├── usecase/           Logique applicative
│   │   ├── repository/        Accès base de données (pgx)
│   │   ├── handler/           Handlers HTTP (Gin)
│   │   └── handler/middleware/ CORS, Auth, RBAC
│   └── docs/openapi/          Spécification OpenAPI générée
├── flutter/
│   └── lib/
│       ├── api/models/        Modèles de données Flutter
│       ├── api/repositories/  Accès à l'ApiService
│       ├── notifiers/         État global (Provider + ChangeNotifier)
│       ├── screens/           Écrans de l'application
│       ├── widgets/           Composants réutilisables
│       └── config/            Router, thème, configuration
├── migrations/                11 migrations SQL séquentielles
└── docs/                      Documentation technique
```

Le diagramme de composants complet est dans [doc-uml-bpmn.md](doc-uml-bpmn.md#2-diagramme-de-composants--architecture-globale).

### 4.3 Décisions architecturales (ADR)

| ADR | Décision | Justification |
| --- | --- | --- |
| ADR-001 | Gin (net/http) plutôt qu'un framework lourd | Performance, contrôle total du routage |
| ADR-002 | Clean Architecture plate | Séparation des responsabilités sans sur-ingénierie |
| ADR-003 | Supabase (PostgreSQL + RLS) | BDD managée avec isolation des données en base |
| ADR-006 | Provider + ChangeNotifier (Flutter) | Gestion d'état simple et testable sans Riverpod |
| ADR-007 | HTTP chunked streaming (WebM/Opus) | Pas de dépendance à un CDN ou un serveur média tiers |
| ADR-008 | media_kit (iOS/Android) + Web Audio API | Décodage natif WebM/Opus multiplateforme |

---

## 5. Schémas UML & BPMN

Les schémas détaillés sont dans [doc-uml-bpmn.md](doc-uml-bpmn.md) (critère Ce3.6.1) :

- **Diagramme de classes** — modèles Flutter (`UserModel`, `StreamModel`, `PlaylistModel`,
  `TrackModel`, rôle `Role`)
- **Diagramme de composants** — architecture Flutter → Go Backend → Supabase
- **Diagramme de séquence** — cycle de vie complet d'une diffusion live
- **BPMN** — parcours utilisateur auditeur de l'ouverture de l'app à la fin d'écoute

---

## 6. Schéma de base de données

Le schéma complet (MCD et MPD) est dans [doc-mcd-mpd.md](doc-mcd-mpd.md).

Tables principales :

| Table | Rôle |
| --- | --- |
| `users` | Comptes, rôles, suspensions — RLS activé |
| `streams` | Canaux de diffusion live et archivés — RLS activé |
| `playlists` | Bibliothèques personnelles — RLS activé |
| `playlist_tracks` | Items de playlist (références `streams`) |
| `favorites` | Streams mis en favori (références `streams`) |
| `listen_history` | Historique d'écoute avec events `join`/`leave` — RLS activé |

Le schéma est versionné via 11 migrations dans `migrations/` ; il ne dépend d'aucun ORM
et peut être réjoué intégralement depuis `001_init.sql`.

---

## 7. Protection des données personnelles — RGPD (Ce3.1.4)

La politique RGPD complète est dans [doc-rgpd.md](doc-rgpd.md).

Points essentiels :

- Seules les données nécessaires au fonctionnement sont collectées (email, prénom, nom,
  historique d'écoute par action volontaire).
- Le consentement est recueilli à l'inscription et par l'action explicite de rejoindre
  un stream.
- La suppression du compte efface en cascade toutes les données associées
  (`ON DELETE CASCADE`).
- Row Level Security (RLS) garantit l'isolation des données en base, indépendamment du
  backend.
- Les droits RGPD (accès, rectification, effacement, portabilité, opposition) sont
  techniquement assurés par l'API et les politiques de cascade.

---

## 8. Contraintes techniques

### 8.1 Performance

| Contrainte | Valeur cible | Mesure |
| --- | --- | --- |
| Fluidité UI | 60 FPS (p90 fil UI ≤ 16,67 ms) | Test Flutter en mode profile — [evidence/performance](../evidence/performance/2026-08-23-rendering/README.md) |
| Capacité streaming | 500 auditeurs simultanés | Benchmark Go hub fan-out — [docs/performance-couts.md](performance-couts.md) |
| Latence API | p95 < 500 ms | Alerte Prometheus `StreamPulseHighLatencyP95` |
| Taux d'erreur | 5xx < 5 % | Alerte Prometheus `StreamPulseHigh5xxRate` |
| Drops audio | < 1 % de chunks abandonnés | Alerte `StreamPulseAudioDrops` |
| Allocation mémoire | 1 clone 32 Kio par chunk (non proportionnel au nombre d'auditeurs) | pprof Go — [docs/performance-couts.md](performance-couts.md) |

### 8.2 Sécurité

- **Authentification** : JWT RS256 (clé privée 4096 bits, jamais commitée).
- **Autorisation** : middleware RBAC Go vérifiant le rôle à chaque handler protégé.
- **Isolation données** : Row Level Security Supabase sur toutes les tables sensibles.
- **Mots de passe** : hachage bcrypt, jamais exposés en clair ni via l'API.
- **Transport** : HTTPS obligatoire en production, géré par Render.
- **Métriques** : endpoint `/metrics` protégé par bearer token.
- **Profiling** : pprof (`/debug/pprof/`) désactivé en production.

### 8.3 Accessibilité

Les widgets interactifs (`StreamCard`, `MiniPlayer`, `AudioControls`) sont annotés avec
`Semantics` Flutter pour la compatibilité avec les lecteurs d'écran (TalkBack, VoiceOver).

---

## 9. Infrastructure et déploiement

### 9.1 Environnements

| Environnement | URL | Backend | Base de données |
| --- | --- | --- | --- |
| Local | `http://localhost:8080` | Docker Compose | Supabase cloud ou local |
| Production | `https://streampulse-api-1jxg.onrender.com` | Render Web Service | Supabase cloud |

### 9.2 Pipeline CI/CD (GitHub Actions)

```
push main
  ├── Tests Go (race detector, couverture)
  ├── Tests Flutter (unit + widget)
  ├── Build Flutter Web
  ├── Vérification contrat OpenAPI
  ├── Publication image Docker Hub (tags SHA + latest) avec SBOM et provenance
  └── Deploy hook Render (après image publiée avec succès)
```

La configuration complète est dans `.github/workflows/deploy.yml`.
Le guide de déploiement initial est dans [docs/deploiement-https.md](deploiement-https.md).

### 9.3 Variables et secrets

Jamais commitées. Configurées dans GitHub (Secrets) et Render (Environment) :

- `JWT_PRIVATE_KEY_FILE` / `JWT_PUBLIC_KEY_FILE` — clés RS256
- `SUPABASE_DB_URL` — chaîne de connexion PostgreSQL
- `CORS_ALLOWED_ORIGINS` — origines autorisées
- `METRICS_BEARER_TOKEN` — protection de `/metrics`
- `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` — publication image
- `RENDER_DEPLOY_HOOK_URL` — déclenchement du déploiement

---

## 10. Observabilité

Le dashboard Grafana `StreamPulse - RNCP Observability` expose :

- Utilisateurs connectés et streams actifs en temps réel
- Débit audio entrant (diffuseur) et sortant (auditeurs) par stream
- Chunks audio livrés et abandonnés
- Latence HTTP (p50 / p95 / p99), taux d'erreur 5xx, requêtes par seconde
- Durée des sessions d'écoute et de diffusion

Alertes Prometheus définies dans `go/infra/prometheus/alerts.yml`.
Documentation complète dans [docs/observability-rncp.md](observability-rncp.md).

---

## 11. Plan de recette

Le plan de recette complet est dans [docs/plan-de-recette.xlsx](plan-de-recette.xlsx).

Scénarios couverts :

| Scénario | Critère de succès |
| --- | --- |
| Inscription et connexion | JWT retourné, redirection vers l'accueil |
| Diffusion live (web) | Auditeur reçoit le flux audio en moins de 2 s |
| Diffusion live (mobile iOS / Android) | Même résultat depuis l'app native |
| Écoute simultanée (500 auditeurs simulés) | 0 % de drops, p95 HTTP < 500 ms |
| Ajout aux favoris / playlist | Persistence après rechargement de la page |
| Recommandations | Streams suggérés différents après enrichissement de l'historique |
| Suspension de compte | Connexion impossible après suspension admin |
| Suppression de compte | Cascade vérifiée : historique, favoris, playlists effacés |
| Fluidité interface | p90 fil UI ≤ 16,67 ms mesuré en mode profile |
| Rollback déploiement | API répond en moins de 30 s après rollback Render |

---

## 12. Documentation associée

| Document | Contenu |
| --- | --- |
| [doc-rgpd.md](doc-rgpd.md) | Données collectées, consentement, droits RGPD (Ce3.1.4) |
| [doc-uml-bpmn.md](doc-uml-bpmn.md) | Diagrammes de classes, composants, séquence, BPMN (Ce3.6.1) |
| [doc-mcd-mpd.md](doc-mcd-mpd.md) | Schéma de base de données MCD / MPD |
| [plan-de-formation-utilisateurs.md](plan-de-formation-utilisateurs.md) | Guide de prise en main par rôle |
| [plan-de-recette.xlsx](plan-de-recette.xlsx) | Scénarios de test et critères d'acceptation |
| [runbook-local.md](runbook-local.md) | Démarrage de l'environnement local |
| [streaming-lifecycle.md](streaming-lifecycle.md) | Cycle de vie détaillé du protocole audio |
| [observability-rncp.md](observability-rncp.md) | Dashboard Grafana, métriques et alertes |
| [performance-couts.md](performance-couts.md) | Benchmarks, capacité et modèle de coûts |
| [deploiement-https.md](deploiement-https.md) | Pipeline CI/CD et configuration Render |
| [go/docs/adr/](../go/docs/adr/) | 8 ADR — décisions architecturales majeures |
