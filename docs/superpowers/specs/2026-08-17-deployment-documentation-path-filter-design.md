# Filtrage documentaire du déploiement de production

## Contexte

Le workflow `.github/workflows/deploy.yml` s'exécute actuellement après chaque
push sur `main`. Un commit qui ajoute uniquement des preuves de production ou
de la documentation reconstruit donc l'image Docker et redéploie inutilement
l'API Render.

## Décision

Ajouter un filtre `paths-ignore` à l'événement `push` du workflow de
déploiement. Les motifs exclus seront :

- `evidence/**` pour les preuves de production archivées ;
- `docs/**` pour la documentation du projet ;
- `**/*.md` pour les fichiers Markdown situés ailleurs dans le dépôt.

Le filtre reste limité au workflow de déploiement. Le workflow CI conserve ses
déclencheurs actuels.

## Comportement attendu

- Un push sur `main` composé uniquement de fichiers correspondant aux trois
  motifs exclus ne déclenche pas le workflow de déploiement.
- Un push qui contient au moins un fichier non exclu déclenche normalement la
  vérification, la publication de l'image et le déploiement de production,
  même si le même commit contient aussi de la documentation.
- Un tag `v*` continue de publier l'image selon le comportement actuel, les
  filtres de chemins ne s'appliquant pas aux pushes de tags.
- Le déclenchement manuel `workflow_dispatch` reste disponible.

## Alternatives écartées

- Une condition au niveau des jobs lancerait tout de même le workflow et
  ajouterait une détection de changements inutilement complexe.
- Un marqueur `[skip ci]` dans le message de commit dépendrait d'une action
  humaine et pourrait ignorer d'autres contrôles que le seul déploiement.

## Validation

La modification sera validée par :

1. une vérification de la syntaxe du workflow ;
2. une vérification automatisée de la présence exacte des trois motifs sous
   `on.push.paths-ignore` ;
3. une vérification que `branches: main`, `tags: v*` et
   `workflow_dispatch` restent configurés ;
4. l'inspection du diff afin de confirmer qu'aucun autre comportement du
   pipeline n'est modifié.

## Risque connu

Si ce workflow est configuré comme contrôle obligatoire d'une pull request,
un workflow ignoré par filtrage de chemins peut rester affiché comme contrôle
en attente. Le workflow actuel ne se déclenche que sur `push` vers `main` ou
sur lancement manuel ; il n'est donc pas utilisé comme contrôle de pull
request dans sa configuration présente.
