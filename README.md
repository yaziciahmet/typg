# typg

**Typed Postgres for Go** — introspect your database and generate type-safe table refs, column methods, and row models. Build SQL with compile-time checks, no stringly-typed queries.

typg embeds a lightweight query builder into your generated package, so your app has **zero runtime dependency** on typg itself. Generated code requires Go 1.27.

## Install

```bash
go get github.com/yaziciahmet/typg
```

## Generate

```go
import (
    "context"
    "os"

    "github.com/yaziciahmet/typg/generate"
)

err := generate.Generate(ctx, generate.Options{
    DSN:           os.Getenv("DATABASE_URL"),
    OutDir:        "./gen/db",
    Schema:        "public",
    ExcludeTables: []string{"schema_migrations"},
})
```

**Output:**

```
gen/
├── db/           # typed columns, joins, query builder facade
│   └── query/    # embedded SQL builder (copied at codegen time)
└── models/       # row structs with db tags + enum types
```

## Query

typg builds the SQL and its arguments; execution and scanning are up to your driver. Pass `sql` and `args` to pgx and scan into the generated models with `pgx.RowToStructByName` (or `RowToStructByNameLax` for partial selects).

```go
sql, args := db.Select(db.Users.Name, db.Posts.Title).
    From(db.Users).
    Join(db.Posts, db.Users.ID.EqCol(db.Posts.UserID)).
    Where(db.Users.Status.Eq(models.UserStatusActive)).
    Where(db.Users.Email.ILike("%@example.com")).
    OrderByDesc(db.Posts.CreatedAt).
    Limit(20).
    Build()
```

Predicates take the column's Go type: `Eq`, `Ne`, `Gt`, `Gte`, `Lt`, `Lte`, `Between`, `In`, `NotIn`, `EqAny`, `Like`, `ILike`, `IsNull`, `NotNull`, `Contains`, plus `EqCol` and friends for column-to-column comparison. Combine with `db.And`, `db.Or`, `db.Not`, `db.Exists` and `db.InSelect`.

### Columns

```go
db.Select(db.Users.All()...)                                  // every column
db.Select(db.Users.AllExcept(db.Users.Meta, db.Users.Tags)...) // all but some
```

### Aliases and expressions

```go
author := db.Users.As("author")

sql, args := db.Select(
    db.Posts.ID.As("post_id"),
    author.Name.As("author_name"),
    db.Count().As("n"),
).
    From(db.Posts).
    Join(author, author.ID.EqCol(db.Posts.UserID)).
    GroupBy(db.Posts.ID, author.Name).
    Having(db.Gt(db.Count(), 1)).
    OrderByDesc(db.Count().As("n")).
    Build()
```

Aggregates and helpers: `Count`, `CountOf`, `Sum`, `Max`, `Min`, `Coalesce`. For anything else use `db.NewRawCol("...")` and `db.RawExpr("...")`; `?` marks an argument and `??` a literal question mark.

### Insert, update, upsert

Every column has typed setters: `Set(v)`, `SetNull()` on nullable columns, `SetExpr(sql, args...)`, and `SetExcluded()`.

```go
sql, args := db.Insert(db.Users).
    Set(db.Users.Email.Set("a@b.com"), db.Users.Name.Set("alice")).
    OnConflict(db.Users.Email).
    DoUpdate(db.Users.Name.SetExcluded(), db.Users.UpdatedAt.SetExpr("now()")).
    Returning(db.Users.ID).
    Build()

sql, args = db.Update(db.Users).
    Set(db.Users.Name.Set("bob"), db.Users.LastLoginAt.SetNull()).
    Where(db.Users.ID.Eq(1)).
    Build()

sql, args = db.Delete(db.Users).Where(db.Users.ID.Eq(1)).Build()
```

### CTEs

```go
recent := db.Select(db.Posts.UserID).From(db.Posts).Where(db.Posts.Published.Eq(true))

sql, args := db.With("recent").As(recent).
    Select(db.Users.ID).
    From(db.Users).
    Join(db.Named("recent"), db.EqCol(db.Users.ID, db.Col("recent", "user_id"))).
    Build()
```

## Example

See [`examples/basic`](examples/basic) for a complete setup with schema, generate tool, and tests.

## License

MIT — see [LICENSE](LICENSE).
