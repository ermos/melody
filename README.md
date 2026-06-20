# melody

🎶 Build your SQL queries like a melody — a small, dialect-agnostic query builder for Go.

```bash
go get github.com/ermos/melody/v2
```

## Quick start

```go
import "github.com/ermos/melody/v2" // imported as package "melody"

q, params, err := melody.New("users").
    Dialect(melody.Postgres).
    Select("id", "name").
    Where("active", "=", true).
    Where("age", ">", 18).
    OrderBy("name", melody.Asc).
    Limit(20).
    Get()
// q      => SELECT id, name FROM users WHERE active = $1 AND age > $2 ORDER BY name ASC LIMIT 20
// params => [true, 18]
```

Every value is bound as a parameter — melody never interpolates values into the SQL string.

## Dialects

The builder emits `?` placeholders internally; the dialect rewrites them on render.

| Dialect           | Placeholders | Upsert                            |
|-------------------|--------------|-----------------------------------|
| `melody.Default`  | `?`          | `ON DUPLICATE KEY UPDATE` (MySQL) |
| `melody.Postgres` | `$1, $2, …`  | `ON CONFLICT … DO UPDATE`         |
| `melody.SQLite`   | `?`          | `ON CONFLICT … DO UPDATE`         |

Implement `melody.Dialect` to add your own.

### Set the dialect once

Instead of repeating `.Dialect()` on every query, configure it once with a
factory and keep it around (e.g. a package var):

```go
m := melody.With(melody.Postgres)

m.New("users").Where("id", "=", 1).Get()       // SELECT ... WHERE id = $1
m.NewInsert("users").Set("name", "bob").Get()  // INSERT ... VALUES( $1 )
```

The package-level `New`/`NewInsert`/… still work, and a factory builder can
still override per query with a further `.Dialect()`.

## SELECT

```go
melody.New("users").
    Distinct().
    Select("country", "count(*)").
    Where("active", "=", true).
    GroupBy("country").
    OrderBy("country", melody.Asc).
    Limit(10).Offset(20).
    Get()
```

### WHERE / OR / IN / groups

```go
melody.New("users").
    Where("status", "IN", "active", "pending").
    OrGroupWhere(func(w *melody.WhereContext) {
        w.Where("age", ">", 65).Where("vip", "=", true)
    }).
    Get()
// ... WHERE status IN (?,?) OR ( age > ? AND vip = ? )
```

### JOIN

```go
melody.New("users").
    LeftJoin("orders", func(w *melody.WhereContext) {
        w.On("orders.user_id", "=", "users.id")
    }).
    Get()
```

### COUNT

```go
melody.New("users").Where("active", "=", true).GetCount()        // SELECT count(*) ...
melody.New("users").GetCountWithKey("id")                        // SELECT count(id) ...
```

## INSERT / UPDATE / DELETE

```go
melody.NewInsert("users").Dialect(melody.Postgres).
    Set("name", "bob").Set("age", 30).
    Returning("id").
    Get()
// INSERT INTO users (name, age) VALUES( $1, $2 ) RETURNING id

melody.NewInsert("users").
    Set("name", "bob").Set("age", 30).
    AddRow().
    Set("name", "alice").Set("age", 25).
    Get()
// INSERT INTO users (name, age) VALUES (?, ?), (?, ?)

melody.NewUpdate("users").
    Set("name", "bob").
    Where("id", "=", 1).
    Get()
// UPDATE users SET name = ? WHERE id = ?

melody.NewDelete("users").
    Where("id", "=", 1).
    Get()
// DELETE FROM users WHERE id = ?
```

### Upsert

The conflict clause is rendered by the dialect:

```go
melody.NewInsert("users").Dialect(melody.Postgres).
    Set("id", 1).
    Set("name", "bob").UpdateDuplicateKey().
    OnConflict("id").
    Get()
// INSERT INTO users (id, name) VALUES( $1, $2 )
//   ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name

// default dialect: ... ON DUPLICATE KEY UPDATE name = ?
```

## HTTP query params → WHERE

Map JSON field names (from your model's `json`/`db` tags) to columns and build a
filtered query straight from request query parameters:

```go
type User struct {
    ID   int    `json:"id" db:"id"`
    Name string `json:"name" db:"name"`
}

b, err := melody.New("users").WithQueryParams(User{}, map[string]string{
    "name%5Blike%5D": "bob",  // name[like]=bob
    "per_page":       "20",
    "page":           "2",
})
// ... WHERE name LIKE ? ... LIMIT 20 OFFSET 20
```

Supported conditions: `equal`, `not-equal`, `like`, `not-like`, `in`, `not-in`.
Unknown params or conditions return an error instead of producing broken SQL.

## Pagination response

```go
result := melody.NewResult(users, total, perPage, page)
// result.Meta.Pages computed as ceil(total/perPage)
```

## Execution

melody is a pure query builder — it never touches the database. Pass its output
to your own driver:

```go
q, params, err := melody.New("users").
    Dialect(melody.Postgres).
    Where("id", "=", 1).
    Get()

rows, err := db.Query(q, params...) // database/sql, pgx, ... your call
```

## Status

- Output: `(query string, params []any, err error)` — bring your own driver.
- Identifiers passed to `Where`/`On`/`OrderBy`/`Select` are emitted as-is; treat them
  as developer-controlled (same trust level as hand-written SQL). User input belongs in
  values (bound params) or in `WithQueryParams` (mapped columns).
