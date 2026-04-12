# ADR-004 — Riverpod vs Bloc (Flutter state management)

**Date :** Sprint 1  
**Statut :** Accepté  
**Auteurs :** Groupe StreamPulse

## Contexte

L'application Flutter nécessite une solution de gestion d'état compatible avec l'architecture Clean Code et adaptée à une équipe de 3 développeurs en rotation.

## Décision

**Riverpod** (`flutter_riverpod`) est retenu comme solution de state management.

## Justification

- **Moins de boilerplate** : pas de `BlocEvent` / `BlocState` à créer pour chaque fonctionnalité.
- **Adapté à la rotation** : la syntaxe déclarative de Riverpod est plus accessible pour des développeurs qui découvrent le code d'un autre membre.
- **Clean Code compatible** : les providers Riverpod sont strictement cantonnés à la couche `presentation/providers/` — ils n'accèdent jamais directement aux repositories.
- **Testable** : `ProviderContainer` permet de tester les providers en isolation.

## Alternatives rejetées

- **Bloc** : plus répandu en entreprise, meilleure séparation événement/état pour les grandes équipes, mais boilerplate significatif (`BlocEvent`, `BlocState`, `BlocBuilder`) peu adapté à la rotation rapide.
- **Provider** : déprécié au profit de Riverpod par le même auteur.
- **GetX** : couplage fort qui violerait la règle Clean Code.

## Conséquences

- Riverpod est **limité à la couche présentation** (`presentation/providers/`).
- Les use cases du domaine sont appelés via les providers, jamais directement depuis les widgets.
- `ProviderScope` est requis à la racine de l'application (`main.dart`).
