# Plan de formation utilisateurs - StreamPulse

## Objectif du document

Ce plan de formation explique comment accompagner les utilisateurs de StreamPulse selon leur rôle : auditeur, diffuseur et administrateur. Il sert de guide d'usage pour la prise en main de l'application et de support RNCP pour montrer que les utilisateurs peuvent exploiter les fonctionnalités livrées.

## Publics concernés

| Rôle | Besoin principal | Fonctionnalités utilisées |
| --- | --- | --- |
| Auditeur | Écouter des streams en direct et gérer sa bibliothèque | Accueil, lecteur audio, inscription, connexion, playlists, favoris |
| Diffuseur | Créer et animer des lives audio | Studio, création de stream, démarrage, arrêt, reprise, suppression |
| Administrateur | Gérer les comptes et suivre les statistiques | Liste utilisateurs, rôles, suspension, réactivation, statistiques |

## Pré-requis

- Avoir accès à l'application StreamPulse web ou mobile.
- Disposer d'un compte de test par rôle : auditeur, diffuseur, administrateur.
- Pour le rôle diffuseur, utiliser un navigateur autorisant l'accès au micro.
- Pour l'administrateur, utiliser un compte possédant déjà le rôle `admin`.
- Savoir se connecter à l'environnement de recette ou de démonstration.

## Organisation de la formation

La formation est prévue sur une session courte de 1 h 30. Elle peut être réalisée en présentiel ou en visioconférence avec partage d'écran.

| Séquence | Durée | Participants | Objectif |
| --- | ---: | --- | --- |
| Présentation générale | 10 min | Tous | Comprendre le rôle de StreamPulse et les règles de sécurité |
| Parcours auditeur | 25 min | Auditeurs, diffuseurs, admin | S'inscrire, se connecter, écouter, gérer playlists et favoris |
| Parcours diffuseur | 25 min | Diffuseurs, admin | Créer, démarrer, diffuser, arrêter et reprendre un live |
| Parcours admin | 20 min | Admin | Gérer les utilisateurs, rôles, suspensions et statistiques |
| Questions et validation | 10 min | Tous | Vérifier l'autonomie des utilisateurs |

## Guide auditeur

### Se créer un compte

1. Ouvrir StreamPulse.
2. Cliquer sur l'icône de compte ou sur le lien d'inscription.
3. Choisir le rôle `Auditeur`.
4. Renseigner prénom, nom, email et mot de passe.
5. Valider avec `Créer mon compte`.

Résultat attendu : le compte est créé et l'utilisateur arrive sur l'accueil.

### Se connecter

1. Ouvrir la page de connexion.
2. Saisir son email et son mot de passe.
3. Cliquer sur `Se connecter`.

Résultat attendu : l'accueil affiche les streams en direct, les recommandations et les éléments personnels si disponibles.

### Écouter un stream

1. Depuis l'accueil, repérer un stream marqué `LIVE`.
2. Cliquer sur le bouton de lecture.
3. Utiliser le lecteur pour mettre en pause, reprendre ou arrêter l'écoute.
4. Quitter le lecteur quand l'écoute est terminée.

Bonnes pratiques : vérifier le volume du terminal, garder une connexion réseau stable et rafraîchir l'accueil si aucun live n'apparaît.

### Gérer ses playlists

1. Aller dans `Ma Bibliothèque`.
2. Ouvrir l'onglet `Playlists`.
3. Cliquer sur `Playlist` pour créer une playlist.
4. Utiliser le menu d'une playlist pour la renommer, ajouter un titre, passer au titre suivant ou supprimer la playlist.

Résultat attendu : les changements sont visibles immédiatement dans la bibliothèque.

### Gérer ses favoris

1. Aller dans `Ma Bibliothèque`.
2. Ouvrir l'onglet `Favoris`.
3. Ajouter un favori depuis l'action prévue.
4. Retirer un favori avec l'icône dédiée.

Résultat attendu : la liste des favoris reflète l'ajout ou le retrait.

## Guide diffuseur

### Accéder au studio

1. Se connecter avec un compte `Diffuseur`.
2. Ouvrir l'onglet ou le lien `Studio`.
3. Vérifier que l'écran `Mon Studio` est visible.

Résultat attendu : le diffuseur voit l'état de diffusion et la liste de ses lives.

### Créer et démarrer un live

1. Saisir un titre de stream clair.
2. Cliquer sur `Démarrer le live`.
3. Autoriser l'accès au micro si le navigateur le demande.
4. Vérifier que l'état passe à `EN DIRECT`.

Résultat attendu : le live devient visible sur l'accueil public et les auditeurs peuvent l'écouter.

### Diffuser correctement

- Parler près du micro sans saturer le son.
- Garder l'onglet StreamPulse ouvert pendant la diffusion.
- Ne pas lancer une écoute du live sur le même appareil sans casque, afin d'éviter un retour audio.
- Prévenir les auditeurs avant d'arrêter le stream.

### Arrêter un live

1. Depuis `Mon Studio`, cliquer sur `Arrêter le stream`.
2. Vérifier que l'état passe à `HORS LIGNE`.
3. Contrôler que le live n'apparaît plus dans la liste publique des lives actifs.

Résultat attendu : l'ingestion audio est arrêtée et les auditeurs ne reçoivent plus de flux.

### Reprendre ou supprimer un live

1. Dans `Mes lives`, choisir un live existant.
2. Cliquer sur `Continuer ce live` pour relancer une session.
3. Cliquer sur `Supprimer ce live` si le live ne doit plus être conservé.
4. Confirmer uniquement si la suppression est volontaire.

Point d'attention : la suppression est définitive.

## Guide administrateur

### Accéder à l'administration

1. Se connecter avec un compte `Administrateur`.
2. Ouvrir la page `Administration`.
3. Vérifier les onglets `Utilisateurs` et `Statistiques`.

Résultat attendu : l'administrateur accède aux fonctions de gestion.

### Lister et contrôler les utilisateurs

1. Ouvrir l'onglet `Utilisateurs`.
2. Vérifier l'email, le nom, le rôle et le statut de chaque compte.
3. Rafraîchir la liste après une modification importante.

Résultat attendu : la liste permet d'identifier les comptes auditeurs, diffuseurs et admins.

### Changer un rôle

1. Ouvrir le menu d'un utilisateur.
2. Cliquer sur `Changer le rôle`.
3. Choisir `Auditeur`, `Diffuseur` ou `Administrateur`.
4. Confirmer.

Règle de sécurité : attribuer le rôle administrateur uniquement à une personne autorisée.

### Suspendre ou réactiver un compte

1. Ouvrir le menu d'un utilisateur.
2. Cliquer sur `Suspendre` pour bloquer l'accès.
3. Vérifier que le badge `Suspendu` apparaît.
4. Utiliser `Réactiver` pour rendre l'accès au compte.

Résultat attendu : un compte suspendu ne peut plus se connecter tant qu'il n'est pas réactivé.

### Consulter les statistiques

1. Ouvrir l'onglet `Statistiques`.
2. Lire le nombre total d'utilisateurs.
3. Vérifier la répartition par rôle.

Usage attendu : suivre l'activité générale des comptes et préparer les démonstrations ou bilans.

## Sécurité et bonnes pratiques

- Utiliser un mot de passe d'au moins 8 caractères.
- Ne jamais partager ses identifiants.
- Se déconnecter après utilisation sur un poste partagé.
- Ne pas attribuer de rôle élevé sans validation.
- Ne pas supprimer ou suspendre un compte sans raison métier claire.
- Signaler toute erreur 401, 403, 404 ou tout problème audio pendant la recette.

## Exercices de validation

| Rôle | Exercice | Critère de réussite |
| --- | --- | --- |
| Auditeur | Créer un compte, se connecter, écouter un live, créer une playlist | L'utilisateur réalise le parcours sans aide |
| Diffuseur | Créer un live, autoriser le micro, arrêter puis reprendre le live | Le live passe bien de `HORS LIGNE` à `EN DIRECT`, puis revient hors ligne |
| Administrateur | Changer un rôle, suspendre puis réactiver un compte de test | Les badges et droits changent conformément à l'action |

## Support et dépannage

| Problème | Cause probable | Action recommandée |
| --- | --- | --- |
| Connexion refusée | Email ou mot de passe incorrect | Vérifier les identifiants, puis réessayer |
| Accès interdit | Rôle insuffisant | Demander le bon rôle à un administrateur |
| Aucun stream visible | Aucun live actif ou liste non rafraîchie | Rafraîchir l'accueil ou attendre un diffuseur |
| Micro non disponible | Permission navigateur refusée | Autoriser le micro dans les paramètres du navigateur |
| Coupure audio | Réseau instable ou live arrêté | Vérifier la connexion, puis relancer la lecture |

## Validation de fin de formation

La formation est considérée comme validée lorsque chaque participant sait :

- se connecter et se déconnecter ;
- identifier son rôle dans l'application ;
- réaliser les actions principales de son rôle ;
- reconnaître un message d'erreur courant ;
- demander de l'aide avec une description claire du problème rencontré.
