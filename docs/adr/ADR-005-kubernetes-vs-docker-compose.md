# ADR-005 — Kubernetes vs docker-compose en production

**Date :** Sprint 1  
**Statut :** Accepté  
**Auteurs :** Groupe StreamPulse

## Contexte

L'infrastructure de déploiement doit être choisie pour héberger l'API StreamPulse en production. Le projet inclut également un bonus K8s/Terraform (US-024).

## Décision

**Kubernetes + Terraform** est retenu pour la production. `docker-compose` est conservé pour le développement local uniquement.

## Justification

| Critère | docker-compose | Kubernetes |
|---------|---------------|-----------|
| Déploiement zero-downtime | ✗ | ✓ Rolling update |
| Autoscaling | ✗ | ✓ HPA (CPU/RPS) |
| Gestion des secrets | Variables ENV | K8s Secrets (chiffrés) |
| Health checks | Basique | Liveness + Readiness probes |
| Observabilité | Manuelle | DaemonSet Promtail/OTEL |
| Complexité opérationnelle | Faible | Élevée |

**La complexité accrue de K8s est justifiée par :**
- Préparation à l'échelle (objectif pédagogique RNCP C3.4).
- HPA permet de gérer les pics de streaming sans intervention manuelle.
- Rolling update garantit le zero-downtime lors des déploiements CI/CD.
- Gestion native des secrets (JWT private key stockée en K8s Secret, jamais dans git).

## Terraform

Les modules Terraform provisionnent :
- Le cluster K8s (GKE / EKS selon le provider).
- Le namespace `streampulse`, les ingress, les secrets.
- Infrastructure as Code versionné dans `infra/terraform/`.

## Conséquences

- `docker-compose.yml` reste pour le développement local (`make docker-up`).
- Les manifests K8s sont dans `infra/k8s/` (Sprint 4 — US-024).
- Les modules Terraform sont dans `infra/terraform/` (Sprint 4 — US-024).
- La clé privée JWT est injectée via un K8s Secret monté en volume — jamais dans l'image Docker.
