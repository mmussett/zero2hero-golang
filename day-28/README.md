# Day 28: Embedding, Templates & Build Tags

Master three production-grade Go features in one day: baking files into your binary with `//go:embed`, generating text and HTML with `text/template` and `html/template`, and controlling what gets compiled with build tags. Every concept is introduced alongside a hands-on lab so you build and verify as you learn.

---

## Table of Contents

- [Section 1: //go:embed](#section-1-goembed)
- [Section 2: text/template](#section-2-texttemplate)
- [Section 3: html/template — Safe HTML Rendering](#section-3-htmltemplate--safe-html-rendering)
- [Section 4: Build Tags](#section-4-build-tags)
- [Section 5: go:generate](#section-5-gogenerate)
- [Final Project: Go Book Store](#final-project-go-book-store)
- [Official Documentation](#official-documentation)

---

## Section 1: //go:embed

`//go:embed` is a compile-time directive that instructs the Go toolchain to read one or more files or directories from the filesystem and bake their contents directly into the binary. The result is a fully self-contained executable — no config files, template directories, or static assets need to be shipped alongside it.

### How it works

```
source tree at build time:
  assets/
    banner.txt   ← read and encoded into the binary
    index.html
    style.css
```

After `go build`, the binary contains the file bytes. The originals are not needed at runtime.

### Three embedding targets

| Declaration | Type | Best for |
|---|---|---|
| `//go:embed file.txt` `var f []byte` | `[]byte` | single binary file |
| `//go:embed file.txt` `var s string` | `string` | single text file |
| `//go:embed dir` `var e embed.FS` | `embed.FS` | trees of files |

### Rules and restrictions

- The path must be relative and must not contain `..`
- Files whose names begin with `.` or `_` are skipped by default — to include them, use the `all:` prefix: `//go:embed all:dir`
- The `embed` package must be imported (even if only the blank identifier is used) when you use `//go:embed`
- The directive must appear immediately before the `var` declaration with no blank lines between them

```go
import _ "embed"   // blank import is enough when using []byte / string forms

//go:embed assets/banner.txt
var banner []byte

// With embed.FS you need the named import
import "embed"

//go:embed assets
var assets embed.FS
```

### Reading embedded files

```go
// []byte / string — use directly
fmt.Println(string(banner))

// embed.FS — use the io/fs helpers
data, err := fs.ReadFile(assets, "assets/index.html")
entries, err := fs.ReadDir(assets, "assets")

// fs.Sub strips a prefix, returning a plain fs.FS rooted at that directory
sub, err := fs.Sub(assets, "assets")
data, err = fs.ReadFile(sub, "index.html")  // no "assets/" prefix needed
```

---

### Lab 1: Embed a Single File as []byte

**What you'll practise:** Declaring a `//go:embed` variable and verifying the binary is self-contained.

**Task:**
Create `assets/banner.txt` containing ASCII art. Embed it as a `[]byte` variable and print it on startup. Build the binary and confirm it runs without the source tree.

**Steps:**
1. Create `assets/banner.txt` with any ASCII art (several lines of text).
2. Add `import _ "embed"` to your source file.
3. Declare `//go:embed assets/banner.txt` immediately above `var banner []byte`.
4. In `main()`, call `fmt.Println(string(banner))`.
5. Run with `go run .`, then build with `go build -o lab1` and run `./lab1` from a different directory (copy the binary).

```go
package main

import (
    _ "embed"
    "fmt"
)

//go:embed assets/banner.txt
var banner []byte

func main() {
    fmt.Println(string(banner))
}
```

**Expected output:**
```
  ____       ____    ___
 / ___| ___ |___ \  / _ \
| |  _ / _ \  __) || | | |
| |_| | (_) |/ __/ | |_| |
 \____|\___/|_____| \___/

Day 28: Embedding, Templates & Build Tags
```

**Checkpoint:** Run `go build -o lab1 . && ./lab1`. The banner prints even when you delete `assets/banner.txt` after building.

---

### Lab 2: Embed a Directory as embed.FS

**What you'll practise:** Using `embed.FS` with `fs.ReadFile` and `fs.ReadDir` to list and read files from an embedded directory.

**Task:**
Add `assets/script.js` to the existing `assets/` directory (which already has `index.html` and `style.css`). Embed the whole directory as `embed.FS`, then use `fs.ReadDir` to list every file and `fs.ReadFile` to print the first 40 bytes of each.

**Steps:**
1. Create `assets/script.js` — a few lines of JavaScript is fine.
2. Add `import "embed"` and `import "io/fs"`.
3. Declare `//go:embed assets` immediately above `var assets embed.FS`.
4. Use `fs.ReadDir(assets, "assets")` to iterate entries.
5. Use `fs.ReadFile(assets, "assets/"+name)` to read each file.

```go
package main

import (
    "embed"
    "fmt"
    "io/fs"
    "log"
)

//go:embed assets
var assets embed.FS

func main() {
    entries, err := fs.ReadDir(assets, "assets")
    if err != nil {
        log.Fatal(err)
    }
    for _, e := range entries {
        info, _ := e.Info()
        data, _ := fs.ReadFile(assets, "assets/"+e.Name())
        preview := string(data)
        if len(preview) > 40 {
            preview = preview[:40] + "..."
        }
        fmt.Printf("%-20s  %5d bytes  %q\n", e.Name(), info.Size(), preview)
    }
}
```

**Expected output:**
```
banner.txt            96 bytes  "  ____       ____    ___\n / ___| ..."
index.html           427 bytes  "<!DOCTYPE html>\n<html lang=\"en\">\n..."
script.js             58 bytes  "// Day 28: demo script\ndocument.ad..."
style.css            512 bytes  "/* Day 28 stylesheet */\n\n*, *::be..."
```

**Checkpoint:** Run `go run .`. All four files appear. Delete any asset file — `go run .` still works because the files are embedded.

---

### Lab 3: Serve Embedded Files over HTTP

**What you'll practise:** Wiring `embed.FS` into `http.FileServer` to serve a directory with no runtime file reads.

**Task:**
Build a minimal HTTP server that serves everything in `assets/` under the URL path `/assets/` using the embedded filesystem from Lab 2.

**Steps:**
1. Import `"net/http"`, `"embed"`, and `"io/fs"`.
2. Use `fs.Sub(assets, "assets")` to strip the `assets/` prefix from the embedded tree.
3. Wrap the result in `http.FileServer(http.FS(sub))` and mount it at `/assets/` with `http.StripPrefix`.
4. Add a root handler that reads and serves `index.html` from the embedded FS.
5. Listen on `:8080` and test with `curl http://localhost:8080/assets/style.css`.

```go
sub, err := fs.Sub(assets, "assets")
if err != nil {
    log.Fatal(err)
}

mux := http.NewServeMux()
mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(sub))))
mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    content, err := fs.ReadFile(sub, "index.html")
    if err != nil {
        http.NotFound(w, r)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(content)
})

log.Println("listening on :8080")
log.Fatal(http.ListenAndServe(":8080", mux))
```

**Expected output:**
```
$ curl -s http://localhost:8080/assets/style.css | head -3
/* Day 28 stylesheet */

*, *::before, *::after {
```

**Checkpoint:** `go run .` starts the server. `curl http://localhost:8080/` returns the HTML. `curl http://localhost:8080/assets/style.css` returns the stylesheet. `curl http://localhost:8080/missing` returns a 404. Kill the source files and rebuild — the binary still serves everything.

---

## Section 2: text/template

The `text/template` package provides data-driven text generation. A template is a string (or file) containing static text interspersed with *actions* — expressions enclosed in `{{` and `}}` — that are evaluated against a data value at execution time.

### Terminology

| Term | Meaning |
|---|---|
| **Template** | The string containing static text and actions |
| **Data** | The Go value passed to `Execute` — called "dot" (`.`) inside the template |
| **Action** | Code inside `{{ }}` — conditionals, loops, definitions, calls |
| **Pipeline** | A sequence of commands chained with `\|` |

### Core actions

```
{{.}}                  print dot (the current data value)
{{.Field}}             print a struct field or map key
{{.Method arg}}        call a method; method must return (T) or (T, error)
{{if .Cond}} … {{end}} conditional
{{if .Cond}} … {{else}} … {{end}}
{{range .Slice}} … {{end}}   iterate; dot becomes each element
{{range .Map}} … {{end}}     iterate map entries
{{with .Val}} … {{end}}      execute block with dot set to .Val (skipped if zero)
{{define "name"}} … {{end}}  define a named sub-template
{{template "name" .}}        include a named sub-template, passing dot
{{block "name" .}} … {{end}} define + execute (can be overridden)
```

### Pipelines

```
{{.Name | printf "%q"}}          pipe dot.Name into printf
{{.Price | printf "%.2f"}}
{{.Tags | join ", "}}            custom function from FuncMap
```

### Whitespace trimming

A dash inside a delimiter trims adjacent whitespace:

```
{{- .Name}}    trim whitespace BEFORE this action
{{.Name -}}    trim whitespace AFTER this action
{{- .Name -}}  trim both sides
```

### The API

```go
// Parse from a string
t := template.Must(template.New("name").Parse(`Hello, {{.Name}}!`))

// Execute — writes to any io.Writer
t.Execute(os.Stdout, data)

// Execute a named sub-template
t.ExecuteTemplate(os.Stdout, "header", data)

// Parse from files
t, err := template.ParseFiles("tmpl/a.html", "tmpl/b.html")

// Parse from an fs.FS (works with embed.FS)
t, err := template.ParseFS(fsys, "tmpl/*.html")
```

### template.Must

`template.Must` panics if the error is non-nil — useful for package-level initialization where a parse failure is a programming error:

```go
var report = template.Must(template.New("report").Parse(`
{{range .Items}}- {{.}}
{{end}}`))
```

---

### Lab 4: Hello Template

**What you'll practise:** Parsing a template from a string and executing it against a struct with `Execute`.

**Task:**
Parse the string `"Hello, {{.Name}}! You have {{.Count}} messages."`, execute it with a struct, and write the output to `os.Stdout`.

**Steps:**
1. Define a struct with `Name string` and `Count int`.
2. Parse the template with `template.Must(template.New("hello").Parse(...))`.
3. Call `t.Execute(os.Stdout, data)` with an instance of your struct.
4. Try changing the data values and observe how the output changes.
5. Replace `os.Stdout` with a `strings.Builder` — capture the output and print its length.

```go
package main

import (
    "os"
    "text/template"
)

type Inbox struct {
    Name  string
    Count int
}

var tpl = template.Must(template.New("hello").Parse(
    "Hello, {{.Name}}! You have {{.Count}} messages.\n",
))

func main() {
    tpl.Execute(os.Stdout, Inbox{Name: "Gopher", Count: 3})
    tpl.Execute(os.Stdout, Inbox{Name: "Alice", Count: 0})
}
```

**Expected output:**
```
Hello, Gopher! You have 3 messages.
Hello, Alice! You have 0 messages.
```

**Checkpoint:** `go run .` produces exactly the two lines above with no extra whitespace. Swap the struct for a `map[string]any{"Name": "Bob", "Count": 7}` — the template executes identically.

---

### Lab 5: Control Flow — range, if, and Whitespace Trimming

**What you'll practise:** Using `{{range}}`, `{{if}}`, `{{else}}`, and whitespace-trimming dashes to generate a tidy text report.

**Task:**
Build a template that receives a `[]string` of to-do items and renders a numbered plain-text report. Items whose text starts with `"DONE:"` should be marked complete; all others are pending. Show how adding `{{-` and `-}}` cleans up unwanted blank lines.

**Steps:**
1. Declare `type Report struct { Title string; Items []string }`.
2. Write a template that ranges over `Items`, uses `{{if}}` to detect the `"DONE:"` prefix, and prints a status indicator.
3. Use `{{- range}}` and `{{- end}}` to eliminate extra blank lines between iterations.
4. Execute with a mix of done and pending items.
5. Compare the output with and without the `-` trims.

```go
const reportTpl = `=== {{.Title}} ===
{{- range $i, $item := .Items}}
{{- if hasPrefix $item "DONE:"}}
  [x] {{$item}}
{{- else}}
  [ ] {{$item}}
{{- end}}
{{- end}}
Total: {{len .Items}} items
`

funcs := template.FuncMap{
    "hasPrefix": strings.HasPrefix,
}
t := template.Must(template.New("report").Funcs(funcs).Parse(reportTpl))
t.Execute(os.Stdout, Report{
    Title: "Sprint Tasks",
    Items: []string{
        "DONE: Write tests",
        "Add documentation",
        "DONE: Fix CI",
        "Deploy to staging",
    },
})
```

**Expected output:**
```
=== Sprint Tasks ===
  [x] DONE: Write tests
  [ ] Add documentation
  [x] DONE: Fix CI
  [ ] Deploy to staging
Total: 4 items
```

**Checkpoint:** `go run .` prints the report with no extra blank lines between items. Remove the `-` from `{{- range}}` and `{{- end}}` — extra blank lines appear. Restore them.

---

### Lab 6: Named Templates and Composition

**What you'll practise:** Defining reusable sub-templates with `{{define}}` and including them with `{{template}}`.

**Task:**
Build a multi-section text report by defining a `"header"` and a `"footer"` sub-template, then compose them into a `"body"` template using `{{template "header" .}}` and `{{template "footer" .}}`. Parse all definitions together using a single `template.Must(template.New("").Parse(...))` call on a concatenated string.

**Steps:**
1. Write a `{{define "header"}}` block that prints a decorated title.
2. Write a `{{define "footer"}}` block that prints a separator and a timestamp field.
3. Write a `{{define "body"}}` block that calls header, ranges over `.Sections`, then calls footer.
4. Use `t.ExecuteTemplate(os.Stdout, "body", data)` to render the composed output.
5. Try calling `t.ExecuteTemplate(os.Stdout, "header", data)` on its own — observe that sub-templates are individually executable.

```go
const (
    headerTpl = `{{define "header"}}
+--- {{.Title}} ---+
{{end}}`
    footerTpl = `{{define "footer"}}
+--- Generated: {{.Generated}} ---+
{{end}}`
    bodyTpl = `{{define "body"}}{{template "header" .}}
{{- range .Sections}}
  * {{.}}
{{- end}}
{{template "footer" .}}{{end}}`
)

type Doc struct {
    Title     string
    Sections  []string
    Generated string
}

t := template.Must(template.New("").Parse(headerTpl + footerTpl + bodyTpl))
t.ExecuteTemplate(os.Stdout, "body", Doc{
    Title:     "Weekly Report",
    Sections:  []string{"Q3 revenue up 12%", "New hire onboarded", "CI fixed"},
    Generated: "2026-10-01",
})
```

**Expected output:**
```
+--- Weekly Report ---+
  * Q3 revenue up 12%
  * New hire onboarded
  * CI fixed
+--- Generated: 2026-10-01 ---+
```

**Checkpoint:** `go run .` prints the composed output. Rename `"body"` to `"report"` and update the `ExecuteTemplate` call — confirm it still works.

---

### Lab 7: FuncMap — Custom Template Functions

**What you'll practise:** Registering custom Go functions via `template.FuncMap` and calling them inside a template with pipelines.

**Task:**
Register three functions — `upper` (wraps `strings.ToUpper`), `join` (wraps `strings.Join`), and `now` (returns a formatted timestamp string) — then use them in a template that renders an inventory report.

**Steps:**
1. Build a `template.FuncMap` with `"upper"`, `"join"`, and `"now"` keys.
2. Call `t.Funcs(funcMap)` **before** calling `t.Parse(...)` — order matters.
3. In the template, use `{{.Name | upper}}`, `{{.Tags | join ", "}}`, and `{{now}}`.
4. Verify that trying to call an unregistered function causes a parse error.
5. Add a fourth function `"currency"` that formats a `float64` as `"$XX.XX"`.

```go
funcs := template.FuncMap{
    "upper": strings.ToUpper,
    "join":  strings.Join,
    "now":   func() string { return time.Now().Format("2006-01-02") },
}

const tplStr = `Item   : {{.Name | upper}}
Tags   : {{.Tags | join ", "}}
Date   : {{now}}
`

t := template.Must(template.New("item").Funcs(funcs).Parse(tplStr))
t.Execute(os.Stdout, struct {
    Name string
    Tags []string
}{"widget", []string{"hardware", "v2", "sale"}})
```

**Expected output:**
```
Item   : WIDGET
Tags   : hardware, v2, sale
Date   : 2026-10-01
```

**Checkpoint:** `go run .` prints the three lines. Comment out the `"upper"` entry from FuncMap, run again — you get a parse error mentioning `"upper"` is undefined. Restore it.

---

### Lab 8: Parse Templates from an Embedded fs.FS

**What you'll practise:** Using `template.ParseFS` to load a template stored inside an `embed.FS`, combining the embedding knowledge from Labs 1–3 with the template knowledge from Labs 4–7.

**Task:**
Store a text report template in `assets/report.tmpl`. Parse it with `template.ParseFS`, passing the embedded FS from Lab 2. Execute it with a list of structs representing products.

**Steps:**
1. Create `assets/report.tmpl` containing a template that ranges over a `[]Product` slice and prints each product's name and price.
2. Import `"embed"`, `"io/fs"`, and `"text/template"`.
3. Declare `//go:embed assets` and embed the directory.
4. Use `template.ParseFS(assets, "assets/report.tmpl")` to parse.
5. Execute with a slice of product structs.

```go
// assets/report.tmpl
Product Report
==============
{{- range .}}
  {{printf "%-30s $%.2f" .Name .Price}}
{{- end}}

Total: {{len .}} products
```

```go
//go:embed assets
var assets embed.FS

type Product struct {
    Name  string
    Price float64
}

func main() {
    t, err := template.ParseFS(assets, "assets/report.tmpl")
    if err != nil {
        log.Fatal(err)
    }
    products := []Product{
        {"The Go Programming Language", 49.99},
        {"Learning Go, 2nd Edition", 39.99},
        {"Concurrency in Go", 34.99},
    }
    t.Execute(os.Stdout, products)
}
```

**Expected output:**
```
Product Report
==============
  The Go Programming Language    $49.99
  Learning Go, 2nd Edition       $39.99
  Concurrency in Go              $34.99

Total: 3 products
```

**Checkpoint:** `go run .` renders the report. Delete `assets/report.tmpl` from disk and rebuild — `go build` fails (the file must be present at build time). This is the trade-off: file must exist when you build, not when you run.

---

## Section 3: html/template — Safe HTML Rendering

`html/template` is a drop-in replacement for `text/template` with one critical difference: it **automatically escapes** all data values before inserting them into HTML output.

### Why it matters

If a user provides the string `<script>alert('xss')</script>` and you insert it into HTML with `text/template`, it executes as JavaScript in the browser. `html/template` encodes it to `&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;`, making it safe.

### Same API, different package

```go
import "html/template"   // instead of "text/template"

// Everything else is identical: New, Parse, Execute, ParseFS, FuncMap, Must
t := template.Must(template.New("page").Parse(`<p>{{.}}</p>`))
t.Execute(os.Stdout, "<b>bold</b>")
// output: <p>&lt;b&gt;bold&lt;/b&gt;</p>
```

### Context-aware escaping

`html/template` understands HTML context — it applies different escaping rules depending on where data appears:

| Context | Example | Escaping applied |
|---|---|---|
| HTML content | `<p>{{.}}</p>` | HTML entity encoding |
| Attribute value | `<a href="{{.URL}}">` | URL + attribute encoding |
| JavaScript string | `<script>var x={{.}}</script>` | JS string escaping |
| CSS value | `<div style="color:{{.Color}}">` | CSS value escaping |

### Trusted HTML with template.HTML

When you have HTML you've generated yourself and trust completely, wrap it in `template.HTML` to bypass escaping:

```go
safe := template.HTML("<strong>trusted content</strong>")
t.Execute(w, safe)  // rendered as-is, not escaped
```

Never use `template.HTML` with user input.

### Rule of thumb

**Always use `html/template` for any output that ends up in a browser.** Use `text/template` only for plain text, config files, or code generation.

---

### Lab 9: Auto-Escaping and XSS Prevention

**What you'll practise:** Observing the concrete difference between `text/template` and `html/template` when rendering untrusted user input.

**Task:**
Render a user-supplied string containing `<script>alert('xss')</script>` with both packages and compare the output. This is the single most important reason to choose `html/template` over `text/template` for web output.

**Steps:**
1. Write a helper `render(pkg string, tplStr string, data any)` that parses the template using both packages and prints the result.
2. Render `<p>{{.}}</p>` with the XSS payload using `text/template`.
3. Render the same template and data using `html/template`.
4. Observe: `text/template` passes the script tag through; `html/template` escapes it.
5. Try injecting into an attribute: `<a href="{{.URL}}">link</a>` with `URL: "javascript:alert(1)"` — note how `html/template` also neutralises this.

```go
import (
    textTpl "text/template"
    htmlTpl "html/template"
)

payload := `<script>alert('xss')</script>`

// text/template — DANGEROUS for HTML output
t1, _ := textTpl.New("").Parse("<p>User says: {{.}}</p>")
fmt.Print("text/template: ")
t1.Execute(os.Stdout, payload)
fmt.Println()

// html/template — safe
t2, _ := htmlTpl.New("").Parse("<p>User says: {{.}}</p>")
fmt.Print("html/template: ")
t2.Execute(os.Stdout, payload)
fmt.Println()
```

**Expected output:**
```
text/template: <p>User says: <script>alert('xss')</script></p>
html/template: <p>User says: &lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</p>
```

**Checkpoint:** Run `go run .`. The `text/template` line contains a raw `<script>` tag — dangerous in a browser. The `html/template` line contains `&lt;script&gt;` — completely safe. This is why you always use `html/template` for web output.

---

### Lab 10: Full HTML Page with a Product Table

**What you'll practise:** Rendering a complete HTML page with `html/template`, using `{{range}}`, `{{if}}`, and dynamic CSS classes to style table rows based on data.

**Task:**
Render a full `<!DOCTYPE html>` page containing a table of products. Rows where `InStock` is `false` should receive a `"out-stock"` CSS class. Use `{{if .InStock}}` to conditionally assign classes and display status text.

**Steps:**
1. Define `type Product struct { Name string; Price float64; InStock bool }`.
2. Write an `html/template` with a `<table>` that ranges over `[]Product`.
3. Use `{{if .InStock}}in-stock{{else}}out-stock{{end}}` as the `class` attribute on each `<tr>`.
4. Register a `"formatPrice"` FuncMap function that formats `float64` as `"$XX.XX"`.
5. Execute to `os.Stdout` and inspect the source with `go run . | grep class`.

```go
const pageTpl = `<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><title>Products</title></head>
<body>
<table>
  <thead><tr><th>Name</th><th>Price</th><th>Status</th></tr></thead>
  <tbody>
  {{- range .}}
  <tr class="{{if .InStock}}in-stock{{else}}out-stock{{end}}">
    <td>{{.Name}}</td>
    <td>{{.Price | formatPrice}}</td>
    <td>{{if .InStock}}In Stock{{else}}Out of Stock{{end}}</td>
  </tr>
  {{- end}}
  </tbody>
</table>
</body>
</html>`

funcs := template.FuncMap{
    "formatPrice": func(f float64) string { return fmt.Sprintf("$%.2f", f) },
}
t := template.Must(template.New("page").Funcs(funcs).Parse(pageTpl))
products := []Product{
    {"The Go Programming Language", 49.99, true},
    {"Go in Action", 44.99, false},
    {"Learning Go, 2nd Edition", 39.99, true},
}
t.Execute(os.Stdout, products)
```

**Expected output (abbreviated):**
```html
<!DOCTYPE html>
<html lang="en">
...
  <tr class="in-stock">
    <td>The Go Programming Language</td>
    <td>$49.99</td>
    <td>In Stock</td>
  </tr>
  <tr class="out-stock">
    <td>Go in Action</td>
    <td>$44.99</td>
    <td>Out of Stock</td>
  </tr>
...
```

**Checkpoint:** `go run . | grep class` shows `in-stock` and `out-stock` classes on the correct rows. Try injecting `<b>` tags in a product name — `html/template` escapes them safely.

---

### Lab 11: Embedded HTML Templates and an HTTP Server

**What you'll practise:** Combining `//go:embed`, `html/template`, and `net/http` into a working web server that renders a dynamic HTML page from embedded template files.

**Task:**
Store the HTML template from Lab 10 as `templates/index.html`. Embed the `templates/` directory with `//go:embed`. Parse it with `template.ParseFS`. Wire it into an HTTP handler that executes the template on each request.

**Steps:**
1. Create `templates/index.html` with the product table template (adapt from Lab 10).
2. Add `//go:embed templates` and `var tmplFS embed.FS` to the source file.
3. Parse with `template.Must(template.New("").Funcs(funcs).ParseFS(tmplFS, "templates/*.html"))`.
4. In the handler, execute with `tmpl.ExecuteTemplate(w, "index.html", products)`.
5. Serve `/` on `:8080`. Test with `curl -s http://localhost:8080/ | grep "<td>"`.

```go
//go:embed templates
var tmplFS embed.FS

func main() {
    funcs := template.FuncMap{
        "formatPrice": func(f float64) string { return fmt.Sprintf("$%.2f", f) },
    }
    tmpl := template.Must(
        template.New("").Funcs(funcs).ParseFS(tmplFS, "templates/*.html"),
    )

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        tmpl.ExecuteTemplate(w, "index.html", products)
    })

    log.Println("listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

**Expected output:**
```
$ curl -s http://localhost:8080/ | grep "<td>"
    <td>The Go Programming Language</td>
    <td>$49.99</td>
    <td>In Stock</td>
    ...
```

**Checkpoint:** `go run .` starts the server. Open `http://localhost:8080/` in a browser — the product table renders. Delete `templates/index.html` from disk and restart — the server still works because the template is embedded in the binary.

---

## Section 4: Build Tags

Build tags (also called *build constraints*) tell the Go toolchain to include or exclude a source file from compilation. They are declared at the very top of a file, before the `package` clause, using the `//go:build` directive.

### Syntax

```go
//go:build linux
//go:build !windows
//go:build linux && amd64
//go:build linux || darwin
//go:build integration || e2e
//go:build prod
```

- `&&` means "and" (both conditions must be true)
- `||` means "or" (either condition is true)
- `!` means "not"
- Parentheses group expressions: `//go:build (linux || darwin) && amd64`

### Built-in tags

| Tag | Set when |
|---|---|
| `linux`, `darwin`, `windows`, … | `GOOS` matches |
| `amd64`, `arm64`, `386`, … | `GOARCH` matches |
| `go1.21`, `go1.22`, … | toolchain version is at least that version |

### Custom tags

Any identifier not matching a built-in tag is a custom tag:

```bash
go build -tags prod .
go build -tags "integration e2e" .
go test -tags integration ./...
```

### File-naming shortcuts

A file named `foo_linux.go` is automatically constrained to `GOOS=linux` without needing a `//go:build` directive. Similarly `foo_amd64.go` constrains on `GOARCH=amd64`. The convention is `name_GOOS.go`, `name_GOARCH.go`, or `name_GOOS_GOARCH.go`.

### Cross-compilation

```bash
GOOS=linux   GOARCH=amd64  go build -o app-linux   .
GOOS=darwin  GOARCH=arm64  go build -o app-darwin   .
GOOS=windows GOARCH=amd64  go build -o app.exe      .
```

No external toolchain is needed — Go ships support for all major OS/arch combinations.

---

### Lab 12: OS-Specific Code with Build Tags

**What you'll practise:** Writing two implementations of the same function, each guarded by a build tag, and observing that the correct one compiles on each platform.

**Task:**
Define `func osName() string` in two files: `platform_unix.go` (compiles on everything except Windows) and `platform_windows.go` (compiles only on Windows). `main.go` calls `osName()` and prints the result.

**Steps:**
1. Create `platform_unix.go` with `//go:build !windows` at the top. Return `"unix-like"`.
2. Create `platform_windows.go` with `//go:build windows` at the top. Return `"windows"`.
3. In `main.go` (no build tag), call `fmt.Println("Running on:", osName())`.
4. Run `go run .` — the right file compiles automatically based on your OS.
5. Cross-compile: `GOOS=windows go build -o app.exe .` should succeed. Inspect which file was compiled with `go list -f '{{.GoFiles}}' .`.

```go
// platform_unix.go
//go:build !windows

package main

func osName() string { return "unix-like" }
```

```go
// platform_windows.go
//go:build windows

package main

func osName() string { return "windows" }
```

```go
// main.go
package main

import "fmt"

func main() {
    fmt.Println("Running on:", osName())
}
```

**Expected output (on Linux/macOS):**
```
Running on: unix-like
```

**Expected output (on Windows):**
```
Running on: windows
```

**Checkpoint:** `go run .` prints the correct OS name. Run `GOOS=windows go vet .` — it passes. Run `go list -f '{{.GoFiles}}' .` to see exactly which `.go` files are in the current build.

---

### Lab 13: Dev vs Prod Template Loading

**What you'll practise:** Using a custom build tag to switch between two implementations of the same function — one reading templates from disk (dev) and one from an embedded FS (prod).

**Task:**
Create `templates_dev.go` (`//go:build !prod`) that returns `os.DirFS("templates")`, and `templates_prod.go` (`//go:build prod`) that uses `//go:embed templates` to return an embedded FS. Both expose `func openTemplates() fs.FS`. Wire this into the HTTP server from Lab 11.

**Steps:**
1. Create `templates_dev.go` with build tag `//go:build !prod`. Implement `openTemplates()` using `os.DirFS`.
2. Create `templates_prod.go` with build tag `//go:build prod`. Add `//go:embed templates` and return the sub-FS.
3. In `main.go` (no build tag), call `openTemplates()` to get the FS, then parse templates from it.
4. Run `go run .` — dev mode, templates reload on restart.
5. Run `go run -tags prod .` — prod mode, templates are embedded.

```go
// templates_dev.go
//go:build !prod

package main

import (
    "fmt"
    "io/fs"
    "os"
)

func openTemplates() (fs.FS, error) {
    fmt.Println("[dev] loading templates from disk")
    return os.DirFS("templates"), nil
}
```

```go
// templates_prod.go
//go:build prod

package main

import (
    "embed"
    "fmt"
    "io/fs"
)

//go:embed templates
var embeddedTemplates embed.FS

func openTemplates() (fs.FS, error) {
    fmt.Println("[prod] loading templates from embedded binary")
    return fs.Sub(embeddedTemplates, "templates")
}
```

**Expected output (dev mode):**
```
[dev] loading templates from disk
listening on http://localhost:8080
```

**Expected output (prod mode):**
```
[prod] loading templates from embedded binary
listening on http://localhost:8080
```

**Checkpoint:** In dev mode, edit `templates/index.html` and restart — changes appear immediately. In prod mode, `go build -tags prod -o server-prod .` produces a binary that serves templates even when the `templates/` directory is deleted.

---

## Section 5: go:generate

`//go:generate` embeds a shell command in a Go source file. Running `go generate ./...` scans all `.go` files in the package tree and executes each directive. This is used to automate code generation — string enumeration, mocks, protocol buffer stubs, and build metadata.

### Syntax

```go
//go:generate <command> [arguments]
```

`go generate` sets several environment variables that generator scripts can read:

| Variable | Value |
|---|---|
| `$GOFILE` | The source file containing the directive |
| `$GOPACKAGE` | The package name |
| `$GODIR` | Absolute path to the directory |
| `$GOARCH` | Current GOARCH |
| `$GOOS` | Current GOOS |

### Common generator programs

```go
//go:generate stringer -type=Status          // generate String() for an int enum
//go:generate mockgen -destination=... ...   // generate interface mocks
//go:generate go run gen/version.go          // run a local generator script
//go:generate protoc --go_out=. service.proto
```

### Generator scripts with //go:build ignore

A generator is usually a standalone Go program in a subdirectory. Add `//go:build ignore` to prevent it from being compiled as part of the package it lives next to:

```go
//go:build ignore

package main  // its own program, not part of the host package

func main() { /* generate stuff */ }
```

---

### Lab 14: Generate Version Info

**What you'll practise:** Writing a generator script that writes a Go source file containing a build timestamp, then wiring it with `//go:generate` and importing the result.

**Task:**
Create `gen/version.go` (a standalone program) that writes `internal/version/version.go` containing `const BuildTime = "<current UTC time>"`. Add `//go:generate go run gen/version.go` to `main.go`. Run `go generate ./...` then `go run .` to see the timestamp in the output.

**Steps:**
1. Create `gen/version.go` with `//go:build ignore` at the top. In its `main()`, format `time.Now().UTC()` as RFC3339 and write it to `internal/version/version.go`.
2. Create `internal/version/version.go` as a placeholder with `const BuildTime = "1970-01-01T00:00:00Z"`.
3. Add `//go:generate go run gen/version.go` above the `package main` line in `main.go`.
4. Run `go generate ./...` — the placeholder is overwritten with the current time.
5. Run `go run .` — the server prints the build time at startup.

```go
// gen/version.go
//go:build ignore

package main

import (
    "fmt"
    "os"
    "time"
)

func main() {
    const dst = "internal/version/version.go"
    if err := os.MkdirAll("internal/version", 0o755); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    content := fmt.Sprintf(`// Code generated by gen/version.go; DO NOT EDIT.

package version

// BuildTime is the UTC timestamp from the last "go generate" run.
const BuildTime = %q
`, time.Now().UTC().Format(time.RFC3339))

    if err := os.WriteFile(dst, []byte(content), 0o644); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
    fmt.Println("generated", dst)
}
```

```go
// internal/version/version.go (placeholder — overwritten by go generate)
package version

const BuildTime = "1970-01-01T00:00:00Z"
```

```go
// main.go (excerpt)
//go:generate go run gen/version.go

package main

import "github.com/mmussett/zero2hero-golang/day-28/internal/version"

func main() {
    fmt.Println("build time:", version.BuildTime)
    // ...
}
```

**Expected output:**
```
$ go generate ./...
generated internal/version/version.go

$ go run .
build time: 2026-10-01T09:45:12Z
  ____       ____    ___
 ...
listening on http://localhost:8080
```

**Checkpoint:** Run `go generate ./...`, then inspect `internal/version/version.go` — the `BuildTime` constant should show the current time. Run `go run .` and confirm the time appears in the startup log.

---

## Final Project: Go Book Store

Bring everything together: a fully self-contained web application that demonstrates every technique from today's labs.

### What it does

- Serves a product catalogue at `/`
- Accepts a `?q=` query parameter to filter products by name
- Renders an HTML page using `html/template` with a custom FuncMap
- Serves static assets (CSS, JS) from `assets/`
- Prints an ASCII-art banner from `assets/banner.txt` at startup
- Displays the build timestamp from `internal/version`
- Switches between disk-based (dev) and embedded (prod) modes via build tag
- Exposes a `//go:generate` directive that refreshes the build timestamp

### File structure

```
day-27/
├── go.mod
├── main.go               ← no build tag: main(), types, handlers, FuncMap
├── main_prod.go          ← //go:build prod: embed + openAssets/openTemplates
├── templates_dev.go      ← //go:build !prod: openAssets/openTemplates from disk
├── assets/
│   ├── banner.txt        ← ASCII art (embedded and printed at startup)
│   ├── index.html        ← static fallback
│   ├── style.css
│   └── script.js
├── templates/
│   └── index.html        ← html/template source
├── gen/
│   └── version.go        ← generator (//go:build ignore)
└── internal/
    └── version/
        └── version.go    ← generated — const BuildTime
```

### Build and run

```bash
# Dev mode — templates and assets read from disk (fast iteration)
go run .

# Prod mode — everything embedded, self-contained binary
go run -tags prod .

# Build a prod binary
go build -tags prod -o bookstore-prod .

# Regenerate the build timestamp
go generate ./...

# Cross-compile
GOOS=linux GOARCH=amd64 go build -tags prod -o bookstore-linux .
```

### Key design decisions

**Single `main()` in `main.go` (no build tag):** The two build-tag files only define `openAssets()` and `openTemplates()` — the shared entry point and all business logic live in a tag-free file.

**`template.ParseFS` after `Funcs`:** The FuncMap must be registered before parsing because template functions are resolved at parse time, not execution time.

**`html/template` not `text/template`:** The product names and search query come from HTTP parameters — always use `html/template` when output goes to a browser.

**`fs.Sub` to strip directory prefix:** Both dev and prod implementations use `fs.Sub` internally so callers always receive a rooted FS regardless of how the files were loaded.

---

## Official Documentation

### embed package

- [`embed`](https://pkg.go.dev/embed) — The `//go:embed` directive and `FS` type. Start here for the rules on what can and cannot be embedded.
- [Go Blog: Using Go Embed](https://go.dev/blog/embed) — Official walkthrough with examples of all three embedding targets (`[]byte`, `string`, `embed.FS`).

### io/fs

- [`io/fs`](https://pkg.go.dev/io/fs) — The `FS` interface, `ReadFile`, `ReadDir`, `WalkDir`, `Sub`, and `Glob`. `embed.FS` implements this interface, making it interchangeable with `os.DirFS` for local development.

### text/template

- [`text/template`](https://pkg.go.dev/text/template) — Full reference for template syntax, actions, pipelines, FuncMap, and the parse/execute API.
- [Go Blog: The Go Template Language](https://go.dev/blog/template) — Deep dive into template composition, context, and custom functions.

### html/template

- [`html/template`](https://pkg.go.dev/html/template) — Same API as `text/template` with context-aware HTML, URL, JS, and CSS escaping.
- [OWASP XSS Prevention](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html) — Why escaping matters and how `html/template` implements it.

### go:generate

- [`go generate`](https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source) — Official command reference, environment variables, and design rationale.
- [Go Blog: Generating code](https://go.dev/blog/generate) — Explains the philosophy and common patterns for code generation in Go.

### Build constraints

- [Build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints) — Complete reference for `//go:build` syntax, boolean operators, GOOS/GOARCH tags, and the `-tags` flag.
- [Go spec: Build constraints](https://pkg.go.dev/cmd/go#hdr-Build_constraints) — The canonical list of built-in constraint tags.

### Cross-compilation

- [Go downloads](https://go.dev/dl/) — Official Go toolchain. Includes support for all GOOS/GOARCH targets with no extra setup.
- [`go env GOARCH GOOS`](https://pkg.go.dev/cmd/go#hdr-Print_Go_environment_information) — Print the current build target. Override with `GOOS=linux GOARCH=amd64 go build .`
