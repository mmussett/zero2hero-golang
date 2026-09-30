# Day 23: database/sql and SQLite

## Core Concept: One Interface, Many Drivers

`database/sql` is a generic interface — drivers plug in for each database. Import a driver for its side effect (registering itself):

```go
import (
    "database/sql"
    _ "modernc.org/sqlite" // registers "sqlite" driver
)

db, err := sql.Open("sqlite", "notes.db")
db.SetMaxOpenConns(1) // SQLite is single-writer
defer db.Close()
```

## Schema Migrations (Simple Pattern)

```go
func migrate(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS notes (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        text       TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    return err
}
```

## CRUD Operations

```go
// Insert with prepared statement
stmt, err := db.Prepare("INSERT INTO notes (text) VALUES (?)")
defer stmt.Close()
result, err := stmt.Exec(text)
id, err := result.LastInsertId()

// Query single row
var n Note
err := db.QueryRow("SELECT id, text, created_at FROM notes WHERE id = ?", id).
    Scan(&n.ID, &n.Text, &n.CreatedAt)
if errors.Is(err, sql.ErrNoRows) { /* not found */ }

// Query multiple rows
rows, err := db.Query("SELECT id, text FROM notes ORDER BY created_at DESC")
defer rows.Close()
for rows.Next() {
    var n Note
    rows.Scan(&n.ID, &n.Text)
    notes = append(notes, n)
}
rows.Err() // always check after loop
```

## Transactions

```go
tx, err := db.Begin()
if err != nil { return err }
defer tx.Rollback() // no-op if committed

_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", amount, from)
if err != nil { return err }
_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", amount, to)
if err != nil { return err }

return tx.Commit()
```

## Context-Aware Queries

All `database/sql` methods have a `Context` variant — prefer them:

```go
db.QueryRowContext(ctx, "SELECT ...", args...)
db.ExecContext(ctx, "INSERT ...", args...)
```

## Day Project: Persist Notes to SQLite

Extend the Day 22 notes API to use SQLite instead of the in-memory store:
1. Replace `Store` with a `DB` struct wrapping `*sql.DB`
2. Implement the same CRUD interface using prepared statements
3. Wrap all mutations in transactions
4. Write integration tests using an in-memory SQLite DB (`:memory:`)

Run with: `go run .`

**Extension ideas:** add full-text search with SQLite FTS5; implement cursor-based pagination.

## Official Documentation

- [`database/sql`](https://pkg.go.dev/database/sql) — `Open`, `DB`, `Stmt`, `Row`, `Rows`, `Tx`, `ErrNoRows`, `QueryRow`, `Query`, `Exec`, `Begin`, `Prepare`, context variants (`QueryRowContext`, `ExecContext`)
- [`errors`](https://pkg.go.dev/errors) — `Is` for checking `sql.ErrNoRows`
- [`context`](https://pkg.go.dev/context) — `Context`-aware query methods
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure-Go SQLite driver (CGo-free)
- [Go Blog: Accessing a relational database](https://go.dev/doc/tutorial/database-access) — official database/sql tutorial
- [Go Blog: Organizing a Go module](https://go.dev/blog/organizing-go-code)
