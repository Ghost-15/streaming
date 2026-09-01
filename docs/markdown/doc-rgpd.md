# Protection des données personnelles — RGPD (Ce3.1.4)

StreamPulse collecte uniquement les données strictement nécessaires à son fonctionnement,
conformément au Règlement Général sur la Protection des Données (RGPD — UE 2016/679).

## Données collectées, finalités et base légale

| Donnée | Finalité | Base légale | Conservation |
| --- | --- | --- | --- |
| Email | Identification et authentification du compte | Consentement (inscription) | Jusqu'à suppression du compte |
| Prénom, nom | Affichage du nom du diffuseur sur les streams | Consentement (inscription) | Jusqu'à suppression du compte |
| Mot de passe (hash bcrypt) | Authentification sécurisée — jamais stocké en clair | Exécution du contrat | Jusqu'à suppression du compte |
| Rôle utilisateur | Contrôle d'accès aux fonctionnalités (user / diffuseur / admin) | Intérêt légitime | Jusqu'à suppression du compte |
| Statut de suspension | Gestion des abus et sécurité de la plateforme | Intérêt légitime | Jusqu'à levée de la suspension |
| Historique d'écoute | Moteur de recommandation personnalisé | Consentement (action volontaire de rejoindre un stream) | Supprimé en cascade à la suppression du compte |
| Favoris et playlists | Bibliothèque personnelle de l'utilisateur | Consentement | Supprimé en cascade à la suppression du compte |
| Métadonnées de diffusion | Titre du stream, nombre d'auditeurs, horodatage | Exécution du contrat | Archivé après la fin du stream |

## Recueil du consentement

**Inscription** — L'utilisateur fournit explicitement son email, prénom, nom et mot de passe.
La validation du formulaire constitue un consentement libre, spécifique et éclairé au
traitement des données d'identification.

**Historique d'écoute** — La collecte est déclenchée par l'action volontaire de rejoindre
un stream (event `join`). L'utilisateur peut quitter à tout moment (event `leave`).
Aucun tracking passif n'est effectué.

## Mesures de sécurité technique

- Hachage bcrypt des mots de passe
- Authentification par JWT (tokens signés)
- HTTPS sur toutes les communications
- Row Level Security (RLS) Supabase : chaque utilisateur n'accède qu'à ses propres données
- Suppression en cascade (`ON DELETE CASCADE`) : la suppression d'un compte efface
  automatiquement l'historique, les favoris, les playlists et les streams associés
- Accès admin via rôle vérifié en base de données (policy RLS dédiée)

**Row Level Security (RLS)** — Chaque politique RLS garantit qu'un utilisateur authentifié
ne peut lire et modifier que ses propres données, indépendamment des requêtes applicatives.
Les administrateurs disposent de policies séparées avec vérification du rôle en base.

## Droits des utilisateurs

| Article | Droit | Application dans StreamPulse |
| --- | --- | --- |
| Art. 15 | Accès | Données consultables via les écrans Profil et Bibliothèque de l'application |
| Art. 16 | Rectification | Prénom, nom et mot de passe modifiables depuis les paramètres du compte |
| Art. 17 | Effacement (droit à l'oubli) | Suppression du compte → suppression en cascade de toutes les données associées |
| Art. 18 | Limitation | Aucun traitement automatique en arrière-plan sans action explicite de l'utilisateur |
| Art. 20 | Portabilité | L'API expose les données au format JSON standardisé, exportable à la demande |
| Art. 21 | Opposition | L'utilisateur peut ne pas rejoindre les streams pour éviter toute collecte d'historique |
