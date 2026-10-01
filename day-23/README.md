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

## Labs

### Lab 1: Open and Ping

**What you'll practise:** Opening a SQLite database, pinging it, and closing it correctly with `defer`.

**Task:**
Open a SQLite database file, verify the connection is live with `db.Ping()`, print the driver name, then close the connection. Confirm that the driver registered itself via the blank import.

**Steps:**
1. `go get modernc.org/sqlite`
2. Import `_ "modernc.org/sqlite"` for the side-effect registration
3. Call `sql.Open("sqlite", "lab.db")` — this does not actually connect yet
4. Call `db.Ping()` to force a real connection; handle the error

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", "lab.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.Ping(); err != nil {
        log.Fatal("ping failed:", err)
    }
    fmt.Println("connected to SQLite")
    fmt.Println("driver:", db.Driver())
}
```

**Expected output:**
```
connected to SQLite
driver: &sqlite.Driver{}
```

**Checkpoint:** The program exits cleanly, `lab.db` is created on disk, and the ping succeeds.

---

### Lab 2: Create Table and Insert

**What you'll practise:** Running DDL with `db.Exec`, preparing an INSERT statement, and executing it multiple times.

**Task:**
Create a `books` table with `id`, `title`, and `author` columns. Use a prepared statement to insert three rows in a loop.

**Steps:**
1. Execute `CREATE TABLE IF NOT EXISTS books (id INTEGER PRIMARY KEY AUTOINCREMENT, title TEXT NOT NULL, author TEXT NOT NULL)`
2. Prepare `INSERT INTO books (title, author) VALUES (?, ?)`
3. Loop over a slice of `{title, author}` structs and execute the statement
4. Print the last inserted ID for each row

```go
_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    title  TEXT NOT NULL,
    author TEXT NOT NULL
)`)
if err != nil {
    log.Fatal(err)
}

stmt, err := db.Prepare("INSERT INTO books (title, author) VALUES (?, ?)")
if err != nil {
    log.Fatal(err)
}
defer stmt.Close()

books := []struct{ Title, Author string }{
    {"The Go Programming Language", "Donovan & Kernighan"},
    {"Clean Code", "Robert C. Martin"},
    {"Designing Data-Intensive Applications", "Martin Kleppmann"},
}
for _, b := range books {
    res, err := stmt.Exec(b.Title, b.Author)
    if err != nil {
        log.Fatal(err)
    }
    id, _ := res.LastInsertId()
    fmt.Printf("inserted id=%d title=%q\n", id, b.Title)
}
```

**Expected output:**
```
inserted id=1 title="The Go Programming Language"
inserted id=2 title="Clean Code"
inserted id=3 title="Designing Data-Intensive Applications"
```

**Checkpoint:** Running the program a second time inserts three more rows (IDs 4-6) because `IF NOT EXISTS` keeps the existing table.

---

### Lab 3: Query Rows into a Struct Slice

**What you'll practise:** Using `db.QueryContext`, iterating with `rows.Next()`/`rows.Scan()`, and checking `rows.Err()` after the loop.

**Task:**
Query all books from the table, scan each row into a `Book` struct, and print the results. Handle the mandatory `rows.Close()` and `rows.Err()` checks.

**Steps:**
1. Call `db.QueryContext(ctx, "SELECT id, title, author FROM books ORDER BY id")`
2. `defer rows.Close()`
3. Loop with `rows.Next()`, call `rows.Scan(&b.ID, &b.Title, &b.Author)`
4. After the loop, check `rows.Err()`

```go
type Book struct {
    ID     int
    Title  string
    Author string
}

ctx := context.Background()
rows, err := db.QueryContext(ctx, "SELECT id, title, author FROM books ORDER BY id")
if err != nil {
    log.Fatal(err)
}
defer rows.Close()

var books []Book
for rows.Next() {
    var b Book
    if err := rows.Scan(&b.ID, &b.Title, &b.Author); err != nil {
        log.Fatal(err)
    }
    books = append(books, b)
}
if err := rows.Err(); err != nil {
    log.Fatal(err)
}
for _, b := range books {
    fmt.Printf("%d: %s by %s\n", b.ID, b.Title, b.Author)
}
```

**Expected output:**
```
1: The Go Programming Language by Donovan & Kernighan
2: Clean Code by Robert C. Martin
3: Designing Data-Intensive Applications by Martin Kleppmann
```

**Checkpoint:** All rows appear; removing `rows.Err()` check and introducing a bug should surface an error you would otherwise miss.

---

### Lab 4: QueryRow — Single-Row Lookup

**What you'll practise:** Using `db.QueryRowContext` for single-row queries and handling `sql.ErrNoRows` explicitly.

**Task:**
Write a `getBook(ctx, db, id) (Book, error)` function. Return a meaningful error message when the ID does not exist rather than letting the caller see a raw `sql.ErrNoRows`.

**Steps:**
1. Call `db.QueryRowContext(ctx, "SELECT id, title, author FROM books WHERE id = ?", id)`
2. Call `.Scan(...)` and check the error
3. If `errors.Is(err, sql.ErrNoRows)`, return a user-friendly `fmt.Errorf("book %d not found", id)`
4. Test with a valid ID and an ID that does not exist

```go
func getBook(ctx context.Context, db *sql.DB, id int) (Book, error) {
    var b Book
    err := db.QueryRowContext(ctx, "SELECT id, title, author FROM books WHERE id = ?", id).
        Scan(&b.ID, &b.Title, &b.Author)
    if errors.Is(err, sql.ErrNoRows) {
        return Book{}, fmt.Errorf("book %d not found", id)
    }
    if err != nil {
        return Book{}, err
    }
    return b, nil
}
```

**Expected output:**
```
book: {1 The Go Programming Language Donovan & Kernighan}
error: book 999 not found
```

**Checkpoint:** The function returns a typed error for missing rows, not the raw `sql.ErrNoRows`.

---

### Lab 5: Transactions

**What you'll practise:** Starting a transaction, running multiple statements atomically, and rolling back on error with a deferred `tx.Rollback`.

**Task:**
Write a `transferTag(ctx, db, fromID, toID int, tag string) error` function that atomically removes a tag from one book and adds it to another using two UPDATE statements in a single transaction.

**Steps:**
1. `tx, err := db.BeginTx(ctx, nil)`; `defer tx.Rollback()`
2. Execute first UPDATE; return if error
3. Execute second UPDATE; return if error
4. Call `tx.Commit()`; the deferred `Rollback` is a no-op after a successful commit

```go
func transferTag(ctx context.Context, db *sql.DB, fromID, toID int, tag string) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback() // no-op after Commit

    _, err = tx.ExecContext(ctx, "UPDATE books SET author = REPLACE(author, ?, '') WHERE id = ?", tag, fromID)
    if err != nil {
        return err
    }
    _, err = tx.ExecContext(ctx, "UPDATE books SET author = author || ? WHERE id = ?", tag, toID)
    if err != nil {
        return err
    }
    return tx.Commit()
}
```

**Expected output:**
```
transfer complete
```

**Checkpoint:** If either UPDATE returns an error the transaction rolls back; both updates are invisible until `Commit` succeeds.

---

### Lab 6: Prepared Statements Performance

**What you'll practise:** Preparing a statement once and executing it many times, observing the benefit of reuse.

**Task:**
Insert 1 000 rows using a single prepared statement and time it. Then repeat the insertion using bare `db.ExecContext` calls (no prepare). Print both durations.

**Steps:**
1. Truncate or recreate the table
2. Time the `stmt.Exec` loop (prepare once, execute 1 000 times)
3. Time the `db.ExecContext` loop (no explicit prepare)
4. Print both durations side by side

```go
// Prepared
start := time.Now()
stmt, _ := db.Prepare("INSERT INTO books (title, author) VALUES (?, ?)")
for i := 0; i < 1000; i++ {
    stmt.Exec(fmt.Sprintf("Book %d", i), "Author")
}
stmt.Close()
preparedDur := time.Since(start)

// Unprepared
start = time.Now()
for i := 0; i < 1000; i++ {
    db.ExecContext(ctx, "INSERT INTO books (title, author) VALUES (?, ?)",
        fmt.Sprintf("Book %d", i), "Author")
}
unpreparedDur := time.Since(start)

fmt.Printf("prepared:   %v\nunprepared: %v\n", preparedDur, unpreparedDur)
```

**Expected output:**
```
prepared:   18ms
unprepared: 22ms
```

**Checkpoint:** The prepared version is consistently faster; both loops insert exactly 1 000 rows.

---

### Lab 7: Schema Migrations

**What you'll practise:** Implementing a simple migration runner that tracks applied versions in a `schema_migrations` table and skips already-applied migrations.

**Task:**
Define a `[]Migration{Version int, SQL string}` slice. Write a `runMigrations(db)` function that creates the tracking table if needed, then applies only the migrations not yet recorded.

**Steps:**
1. Create `schema_migrations(version INTEGER PRIMARY KEY, applied_at DATETIME)` if it does not exist
2. For each migration, query whether its version is already in the table
3. If not, execute the migration SQL and insert the version into the tracking table — all in one transaction
4. Run the program twice; the second run should skip all migrations

```go
type Migration struct {
    Version int
    SQL     string
}

var migrations = []Migration{
    {1, `CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, text TEXT NOT NULL)`},
    {2, `ALTER TABLE notes ADD COLUMN created_at DATETIME DEFAULT CURRENT_TIMESTAMP`},
}

func runMigrations(db *sql.DB) error {
    _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version    INTEGER PRIMARY KEY,
        applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        return err
    }
    for _, m := range migrations {
        var exists int
        db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", m.Version).Scan(&exists)
        if exists > 0 {
            fmt.Printf("migration %d already applied, skipping\n", m.Version)
            continue
        }
        tx, _ := db.Begin()
        tx.Exec(m.SQL)
        tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.Version)
        if err := tx.Commit(); err != nil {
            tx.Rollback()
            return err
        }
        fmt.Printf("applied migration %d\n", m.Version)
    }
    return nil
}
```

**Expected output (first run):**
```
applied migration 1
applied migration 2
```

**Expected output (second run):**
```
migration 1 already applied, skipping
migration 2 already applied, skipping
```

**Checkpoint:** Migrations are idempotent; a new migration appended to the slice runs only once.

---

### Final Lab: Notes API + SQLite

**What you'll practise:** Replacing an in-memory store with a real SQLite database, using prepared statements, transactions, and an in-memory test DB.

**Task:**
Extend the Day 22 Notes API to persist data in SQLite. Replace `NoteStore` with a `DB` struct wrapping `*sql.DB`. Apply a migration on startup. Use transactions for mutating operations. Test with `:memory:`.

**Steps:**
1. Define `DB{db *sql.DB}` with the same method signatures as `NoteStore`
2. Run `runMigrations` on startup to create the `notes` table
3. Implement `Create`, `List`, `Get`, `Update`, `Delete` using prepared statements
4. Wrap `Create`, `Update`, `Delete` in transactions
5. In tests, open `sql.Open("sqlite", ":memory:")` for a clean database on each test run

```go
type DB struct{ db *sql.DB }

func (s *DB) Create(ctx context.Context, text string) (Note, error) {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return Note{}, err
    }
    defer tx.Rollback()
    res, err := tx.ExecContext(ctx, "INSERT INTO notes (text) VALUES (?)", text)
    if err != nil {
        return Note{}, err
    }
    id, _ := res.LastInsertId()
    if err := tx.Commit(); err != nil {
        return Note{}, err
    }
    return s.Get(ctx, strconv.FormatInt(id, 10))
}
```

**Expected output:**
```
$ go run .
migrations applied
listening on :8080
$ go test -v ./...
--- PASS: TestCreate (0.00s)
--- PASS: TestGetNotFound (0.00s)
--- PASS: TestCRUDCycle (0.00s)
PASS
```

**Checkpoint:** Data survives a server restart (it is on disk); tests use `:memory:` so they are stateless.

---

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
