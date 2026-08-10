# ADR-006 — Provider et ChangeNotifier pour l'état Flutter

## Statut

Accepté — Sprint final.

## Contexte

Le client Flutter doit partager l'authentification, les listes de streams, le
lecteur live, le studio diffuseur, les favoris et les playlists entre plusieurs
écrans. L'équipe a comparé Provider/ChangeNotifier, Riverpod et Bloc.

## Décision

Utiliser `provider` avec des `ChangeNotifier` spécialisés par domaine : session,
streams, audio, diffusion, favoris, playlists et administration. Les widgets
observent uniquement le notifier utile avec `watch` et déclenchent les actions
avec `read`.

Le code dépend d'une interface stable exposée par des exports conditionnels :
la version Web implémente `MediaRecorder`, `MediaSource` et Media Session ; le
fallback non-Web conserve la compilation et signale explicitement que le direct
audio est disponible dans le client Web.

## Conséquences

### Positives

- peu de code d'infrastructure pour une application de cette taille ;
- états de chargement, lecture, pause, buffer et erreur faciles à exposer ;
- notifiers testables sans monter toute l'interface ;
- cycle de vie des ressources Web centralisé dans `dispose`.

### Limites

- les effets asynchrones doivent protéger explicitement les courses et l'état
  après destruction ;
- une application plus grande pourrait bénéficier des types immuables et de la
  composition plus stricte de Riverpod ou Bloc ;
- la capture et la lecture natives Android/iOS nécessiteraient une implémentation
  conditionnelle supplémentaire et un service audio de fond dédié.

## Alternatives rejetées

- **Bloc** : robuste mais trop cérémoniel pour le nombre actuel d'états.
- **Riverpod** : bonne sûreté de typage, mais une migration n'apporte pas assez de
  valeur au périmètre final et augmenterait le risque avant la soutenance.

## Vérification

Les tests de notifiers vérifient les états initiaux, transitions, erreurs et
notifications. `flutter analyze`, `flutter test` et `flutter build web --release`
font partie de la validation avant livraison.
