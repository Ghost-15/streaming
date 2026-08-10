# ADR-004 - Docker Compose et deploiement conteneurise

**Date :** Sprint 1, revise Sprint 4  
**Statut :** Accepte  
**Auteurs :** Groupe StreamPulse

## Contexte

L'infrastructure doit rester simple a exploiter pour le rendu RNCP tout en
montrant une application cloud-native : configuration par variables
d'environnement, image Docker optimisee, observabilite et secrets externalises.

Le perimetre de livraison se concentre sur une stack Docker Compose
reproductible en local et une image API deployable sur un service conteneurise
gere, une VM Docker ou un PaaS.

## Decision

`docker-compose.yml` est retenu comme environnement de developpement, de demo et
de validation locale. La production cible une execution conteneurisee simple :
image Docker de l'API, variables d'environnement injectees par la plateforme,
secrets montes sous forme de fichiers ou variables protegees, et supervision via
Prometheus/Grafana.

## Justification

| Critere | Docker Compose / conteneur gere | Infrastructure avancee hors perimetre |
| --- | --- | --- |
| Reproductibilite locale | Tres bonne | Plus lourde |
| Temps de mise en oeuvre | Faible | Eleve |
| Observabilite RNCP | Couverte par Prometheus, Grafana, OTEL et Loki | Possible mais non necessaire |
| Gestion des secrets | Fichiers Docker secrets / variables protegees | Plus avancee mais hors perimetre |
| Risque de rendu incomplet | Faible | Eleve |

Ce choix evite de promettre une infrastructure qui ne sera pas livree et garde
le projet concentre sur les attendus principaux : API Go, streaming, application
Flutter, observabilite et industrialisation Docker.

## Consequences

- `docker-compose.yml` reste la source de verite pour lancer la stack locale.
- Le Dockerfile multi-stage/distroless reste l'artefact de deploiement principal
  de l'API.
- Les variables d'environnement et secrets restent externalises via `.env`,
  secrets Docker ou mecanismes equivalents de la plateforme cible.
- Les fichiers d'orchestration avances ne font pas partie du projet.
