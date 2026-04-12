# ADR-002 — Clean Code vs Domain-Driven Design (DDD)

**Date :** Sprint 1  
**Statut :** Accepté  
**Auteurs :** Groupe StreamPulse

## Contexte

L'architecture applicative doit être choisie pour structurer le code backend Go et l'application Flutter. Deux approches ont été considérées : Clean Code (Uncle Bob) et Domain-Driven Design (DDD).

## Décision

**Clean Code** avec couches explicites (Entity → UseCase → Repository → Handler) est retenu.

## Règle d'or

Le flux de dépendances va **uniquement de l'extérieur vers l'intérieur** :

```
Handler → UseCase → Entity
Infrastructure → Repository (interface) → Entity
```

Un Handler n'est jamais importé par un UseCase. Une Entity n'importe rien.

## Justification

Clean Code est plus simple à maîtriser pour 3 développeurs en rotation :
- Couches claires avec responsabilités explicites.
- Pas d'Aggregate, ValueObject, DomainEvent complexes à justifier en soutenance.
- Le code est testable à chaque niveau (Entity, UseCase, Handler) de manière indépendante.
- La même structure s'applique côté Flutter (`domain/` → `data/` → `presentation/`).

## Alternatives rejetées

- **DDD** : plus expressif pour les domaines complexes, mais surcharge cognitive pour 3 devs en rotation. Les concepts Aggregate/ValueObject auraient nécessité plus d'abstraction sans bénéfice visible à cette échelle.

## Conséquences

- Toute violation de la règle de dépendances = dette technique documentée dans un ADR.
- Les interfaces `Repository` sont dans `internal/repository/` — jamais d'implémentation dans ce package.
- Injection de dépendances manuelle dans `cmd/server/main.go` (pas de framework DI).
