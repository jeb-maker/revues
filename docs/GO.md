# Bonnes pratiques Go — Revues

Guide normatif pour agents et contributeurs. Complète [CONVENTIONS.md](./CONVENTIONS.md).

Références officielles : [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md) (sélectif).

---

## Style et formatage

| Règle | Détail |
|-------|--------|
| `gofmt` | Obligatoire — vérifié par `check.sh` |
| `go vet` | Obligatoire — zéro warning |
| `golangci-lint` | Obligatoire si installé — config `.golangci.yml` |
| Imports | Groupe std / externes / internes ; `goimports` implicite via fmt |
| Longueur ligne | ~100 caractères (souple) |
| Commentaires | Sur exports publics uniquement ; pas de commentaires évidents |

---

## Structure des packages

```
internal/<domain>/     # logique métier par domaine
internal/store/        # seul package avec SQL
internal/web/          # handlers HTTP fins + middleware
```

| Règle | Exemple |
|-------|---------|
| Handlers fins | handler parse → appelle service → render |
| Pas de SQL hors `store` | `internal/projects/service.go` appelle `store.ProjectByID` |
| Pas de logique métier dans `main` | `main` = wiring, config, démarrage |
| `internal/` | API non exportée hors module |
| Pas de cycles | `store` ne importe pas `web` |

---

## Erreurs

```go
// ✅ Wrapper avec contexte
if err != nil {
    return fmt.Errorf("load project %d: %w", id, err)
}

// ✅ Erreurs sentinelles pour branchement
var ErrNotFound = errors.New("not found")

// ❌ Ignorer silencieusement
_ = rows.Close()

// ❌ panic hors main
panic("unexpected")
```

| Règle | Détail |
|-------|--------|
| Toujours traiter `error` | `errcheck` enforced |
| `%w` pour wrapping | permet `errors.Is` / `errors.As` |
| Pas de `panic` | sauf `main` ou init irrécupérable |
| HTTP | mapper erreurs → status (404, 403, 400, 500) — pas de fuite interne |

---

## context.Context

```go
func (s *Service) GetProject(ctx context.Context, id int64) (*Project, error) {
    return s.store.ProjectByID(ctx, id)
}

func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
    project, err := h.svc.GetProject(r.Context(), id)
    // ...
}
```

- **Premier argument** `ctx context.Context` sur toutes les fonctions I/O (DB, HTTP, OAuth)
- Propager `r.Context()` depuis les handlers
- Pas de `context.Background()` dans le chemin requête

---

## base de données (`database/sql`)

```go
// ✅ Transaction atomique
tx, err := db.BeginTx(ctx, nil)
if err != nil { return err }
defer tx.Rollback()

if err := updateItem(ctx, tx, item); err != nil { return err }
if err := insertEvent(ctx, tx, event); err != nil { return err }
return tx.Commit()
```

| Règle | Détail |
|-------|--------|
| Requêtes paramétrées | `$1` / `?` — **jamais** concat SQL |
| `defer rows.Close()` | systématique après `Query` |
| Transactions | update + audit dans la même tx |
| Connexion | pool injecté ; `REVUES_DB_MAX_OPEN_CONNS` (défaut 10), `busy_timeout` 5 s |
| PRAGMA | `foreign_keys=ON`, `journal_mode=WAL` au démarrage |
| Timeouts | `context` sur requêtes longues |

---

## HTTP (chi)

```go
r.With(middleware.RequireAuth, middleware.RequireRole("editor")).
    Post("/projects/{id}/runs", h.CreateRun)
```

| Règle | Détail |
|-------|--------|
| Méthodes explicites | `Get`, `Post`, `Patch` — pas de catch-all |
| Chi URL params | `chi.URLParam(r, "id")` + validation |
| Status codes | `http.Status*` constants |
| Body limit | `http.MaxBytesReader` sur uploads |
| Timeouts serveur | `ReadHeaderTimeout`, `WriteTimeout` sur `http.Server` |

---

## Interfaces

```go
// ✅ Petite interface côté consommateur
type ProjectStore interface {
    ProjectByID(ctx context.Context, id int64) (*Project, error)
}

// ❌ Interface large « au cas où »
type Store interface { /* 40 méthodes */ }
```

- **Accept interfaces, return structs**
- Interfaces définies **là où elles sont utilisées** (pas dans `store`)
- Mocks générés ou manuels pour tests

---

## Tests

```go
func TestSnapshotRunItems(t *testing.T) {
    tests := []struct {
        name    string
        items   []TemplateItem
        wantLen int
    }{
        {"empty template", nil, 0},
        {"three items", items, 3},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := snapshot(tt.items)
            if len(got) != tt.wantLen {
                t.Errorf("len = %d, want %d", len(got), tt.wantLen)
            }
        })
    }
}
```

| Règle | Détail |
|-------|--------|
| Table-driven | préféré pour logique métier et RBAC |
| `t.Helper()` | dans helpers de test |
| SQLite mémoire | `:memory:` pour tests `store` |
| `httptest` | tests handlers sans réseau |
| `-race` | exécuté par `check.sh` |
| Nommage | `Test<Function>_<scenario>` |
| Pas de sleep | utiliser channels ou mocks |

Fichiers : `*_test.go` colocalisés ou `package xxx_test` pour tests d'intégration.

---

## Concurrence

| Règle | Détail |
|-------|--------|
| Goroutines emails/webhooks | avec `context` et log erreur |
| Pas de goroutine leak | lifecycle lié à la requête ou shutdown graceful |
| Mutex | seulement si état partagé in-process (éviter si possible) |
| SQLite | un writer — transactions courtes ; lectures parallèles via pool + WAL ; retry `SQLITE_BUSY` sur écritures critiques |

---

## Logging

```go
slog.Error("webhook delivery failed",
    "event_id", eventID,
    "url", redactURL(url),
    "err", err,
)
```

- **`log/slog`** (stdlib) — structuré
- Niveaux : `Debug` dev, `Info` métier, `Error` échecs
- **Jamais** logger secrets, tokens, mots de passe, corps SMTP

---

## Sécurité Go spécifique

| Règle | Détail |
|-------|--------|
| `crypto/rand` | tokens session, CSRF |
| Pas `math/rand` | pour secrets |
| Comparaison constant-time | `subtle.ConstantTimeCompare` pour tokens |
| `html/template` | auto-escape ; pas `template.HTML` sur input user |
| `json.Decoder` | `DisallowUnknownFields()` si API externe |

---

## Dépendances

| Règle | Détail |
|-------|--------|
| Minimalisme | stdlib d'abord ; chi, goose, modernc sqlite OK |
| `go mod tidy` | avant chaque PR touchant imports |
| Pas de wrapper inutile | pas de lib qui réexporte stdlib |
| Versions | tag sémantique ; pas de `latest` sans raison |

---

## Anti-patterns interdits

| ❌ | ✅ |
|----|-----|
| `init()` avec logique métier | wiring explicite dans `main` |
| Variables globales mutables | injection de dépendances |
| `interface{}` / `any` partout | types concrets ou interfaces ciblées |
| Cast sans `ok` | `v, ok := x.(T)` |
| SQL dans handlers | `internal/store` |
| ORM (GORM, etc.) | SQL explicite |
| `ioutil` (déprécié) | `io` / `os` |
| `goto` | boucles / fonctions |

---

## Checklist agent (avant PR)

- [ ] `gofmt` / `go vet` / `go test -race ./...` verts
- [ ] `golangci-lint run` vert (si installé)
- [ ] `context.Context` propagé sur I/O
- [ ] Erreurs wrappées, aucune ignorée
- [ ] SQL paramétré dans `store` uniquement
- [ ] Tests table-driven sur logique touchée
- [ ] Pas de `panic`, pas de secret loggé
- [ ] `./scripts/check.sh` vert
