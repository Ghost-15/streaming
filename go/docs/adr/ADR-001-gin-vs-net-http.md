# ADR-001 — Gin vs net/http

**Date :** Sprint 1  
**Statut :** Accepté  
**Auteurs :** Groupe StreamPulse

## Contexte

Le projet nécessite un framework HTTP Go pour l'API backend. Deux options principales ont été évaluées : la bibliothèque standard `net/http` et le framework Gin.

## Décision

**Gin** (`github.com/gin-gonic/gin`) est retenu comme framework HTTP.

## Justification

| Aspect | `net/http` | Gin | Bénéfice |
|--------|-----------|-----|---------|
| Routing | `http.HandleFunc('/path', fn)` | `gin.GET('/path', fn)` + `RouterGroup` | Groupes `/api/v1`, middleware ciblé |
| Middleware | Chainage manuel | `gin.Use(middleware)` | Par groupe ou global, rollback propre |
| Binding/Validation | Décodage JSON manuel | `c.ShouldBindJSON(&req)` + tags `binding:` | Validation auto, erreurs formatées |
| Réponse JSON | `json.NewEncoder(w).Encode()` | `c.JSON(http.StatusOK, gin.H{...})` | Status + body en 1 ligne |
| OTEL | `otelhttp.NewHandler()` | `otelgin.Middleware(serviceName)` | Auto-instrumentation des routes |
| Tests | `httptest.NewRecorder()` | `httptest.NewRecorder()` ✓ | Aucun changement côté test |

**Réduction de boilerplate estimée : ~40%**

## Alternatives rejetées

- **Echo** : aussi performant que Gin, mais moins d'exemples OTEL documentés au moment du choix.
- **Fiber** : basé sur `fasthttp`, incompatible avec `net/http` — casse l'écosystème middleware existant.

## Conséquences

- Dépendance externe supplémentaire (justifiée par la réduction de boilerplate).
- Gin est un wrapper léger sur `net/http` — performances identiques (benchmark httprouter).
- L'intégration `otelgin` fournit l'auto-instrumentation en 1 ligne dans `router.go`.
