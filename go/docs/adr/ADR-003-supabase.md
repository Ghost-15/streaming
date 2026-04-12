# ADR-003 — Supabase vs PostgreSQL managé

**Date :** Sprint 1  
**Statut :** Accepté  
**Auteurs :** Groupe StreamPulse

## Contexte

Le projet nécessite une base de données relationnelle hébergée avec gestion des droits d'accès au niveau des lignes (Row Level Security).

## Décision

**Supabase** (PostgreSQL + RLS policies) est retenu comme solution de base de données.

## Justification

- **RLS intégré** : les politiques de sécurité au niveau des lignes garantissent l'isolation des données indépendamment du backend.
- **Auth optionnelle** : on gère le JWT côté Go pour garder le contrôle complet sur les claims RBAC.
- **Storage inclus** : disponible pour les assets audio si nécessaire.
- **SDK Go compatible** : on utilise `database/sql` standard avec `pgx/v5` — pas de vendor lock-in sur le client.

## Connexion Go

```go
// Pool de connexions pgx
cfg.MaxConns = 25
cfg.MinConns = 5
```

## Alternatives rejetées

- **PostgreSQL managé (RDS, Cloud SQL)** : RLS disponible mais configuration manuelle, pas d'interface d'administration intégrée.
- **PlanetScale / Neon** : pas de support RLS natif comparable.

## Conséquences

- Vendor lock-in **partiel** : les RLS policies sont du SQL PostgreSQL standard. Migration vers pg pur possible car on utilise `database/sql` standard.
- Les migrations sont versionnées dans `migrations/` et appliquées via `psql` ou `supabase db push`.
- **Supabase Auth non utilisé** : le JWT est géré côté Go pour les claims RBAC personnalisés.
