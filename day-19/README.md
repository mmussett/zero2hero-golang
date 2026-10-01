# Day 19: JSON and CSV Encoding

## Core Concept: Struct Tags Drive Serialisation

Go's `encoding/json` uses reflection and struct field tags to map between Go types and JSON automatically.

```go
type User struct {
    ID        int    `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email,omitempty"` // omit if empty
    Password  string `json:"-"`               // never serialise
    CreatedAt time.Time `json:"created_at"`
}

u := User{ID: 1, Name: "Alice"}
data, err := json.Marshal(u)
// {"id":1,"name":"Alice","created_at":"0001-01-01T00:00:00Z"}

var parsed User
err = json.Unmarshal(data, &parsed)
```

## Streaming JSON (for large payloads)

```go
// Encode to a writer
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
enc.Encode(users)

// Decode from a reader
dec := json.NewDecoder(resp.Body)
dec.DisallowUnknownFields() // strict mode
var result Response
err := dec.Decode(&result)
```

## Raw JSON and Dynamic Parsing

```go
// Delay parsing with json.RawMessage
type Envelope struct {
    Type    string          `json:"type"`
    Payload json.RawMessage `json:"payload"`
}

// Parse into map for unknown structure
var m map[string]any
json.Unmarshal(data, &m)
```

## Custom Marshaling

```go
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil { return err }
    dur, err := time.ParseDuration(s)
    *d = Duration(dur)
    return err
}
```

## encoding/csv

```go
// Reading
r := csv.NewReader(strings.NewReader(data))
r.FieldsPerRecord = -1  // variable column count
records, err := r.ReadAll()

// Streaming
for {
    record, err := r.Read()
    if err == io.EOF { break }
}

// Writing
w := csv.NewWriter(os.Stdout)
w.Write([]string{"name", "age", "email"})
w.Flush()
```

## Labs

### Lab 1: Basic JSON Marshal/Unmarshal

**What you'll practise:** Encoding a struct to JSON and decoding it back using `json.Marshal` and `json.Unmarshal`.

**Task:**
Define a `Person` struct with `json` tags, marshal it to indented JSON, print it, then unmarshal the bytes back into a new variable and verify the round-trip.

**Steps:**
1. Define `Person` with fields `Name`, `Age`, and `Email` with `json:"..."` tags
2. Marshal a `Person` value using `json.MarshalIndent`
3. Print the JSON string
4. Unmarshal back into a second `Person` and print both values

```go
type Person struct {
    Name  string `json:"name"`
    Age   int    `json:"age"`
    Email string `json:"email"`
}

p := Person{Name: "Alice", Age: 30, Email: "alice@example.com"}
data, err := json.MarshalIndent(p, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))

var p2 Person
if err := json.Unmarshal(data, &p2); err != nil {
    log.Fatal(err)
}
fmt.Printf("%+v\n", p2)
```

**Expected output:**
```
{
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com"
}
{Name:Alice Age:30 Email:alice@example.com}
```

**Checkpoint:** Both the original and decoded struct print identical field values.

---

### Lab 2: JSON Struct Tags — omitempty and Field Renaming

**What you'll practise:** Controlling JSON output with `omitempty`, `-`, and custom field names.

**Task:**
Define a `User` struct that demonstrates three tag behaviours: renamed field, omitted-when-empty field, and never-serialised field. Marshal two instances — one with all fields set, one with empty fields — and compare the output.

**Steps:**
1. Define `User` with `ID int`, `Name string`, `Email string` (omitempty), and `Password string` (never serialise)
2. Marshal a fully populated `User` and an empty-email `User`
3. Observe which fields appear in each output

```go
type User struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email,omitempty"`
    Password string `json:"-"`
}

full  := User{ID: 1, Name: "Bob", Email: "bob@example.com", Password: "secret"}
empty := User{ID: 2, Name: "Carol"}

for _, u := range []User{full, empty} {
    b, _ := json.MarshalIndent(u, "", "  ")
    fmt.Println(string(b))
}
```

**Expected output:**
```
{
  "id": 1,
  "name": "Bob",
  "email": "bob@example.com"
}
{
  "id": 2,
  "name": "Carol"
}
```

**Checkpoint:** `Password` never appears; `email` is absent when empty; no other fields are missing.

---

### Lab 3: Streaming JSON with Decoder/Encoder

**What you'll practise:** Decoding a JSON array token-by-token from a `strings.Reader` without loading everything into memory at once.

**Task:**
Build a function that reads a JSON array of `Product` objects from a `strings.Reader` using `json.NewDecoder` and streams each decoded product to `os.Stdout` via `json.NewEncoder`.

**Steps:**
1. Prepare a multi-item JSON array string
2. Create a `json.Decoder` around a `strings.NewReader`
3. Read the opening `[` token, then loop calling `dec.Decode(&p)` until `]`
4. Encode each decoded product to stdout with `json.NewEncoder`

```go
const input = `[
  {"id":1,"name":"Widget","price":9.99},
  {"id":2,"name":"Gadget","price":24.50},
  {"id":3,"name":"Doohickey","price":4.75}
]`

type Product struct {
    ID    int     `json:"id"`
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

dec := json.NewDecoder(strings.NewReader(input))
dec.Token() // consume '['
enc := json.NewEncoder(os.Stdout)
for dec.More() {
    var p Product
    if err := dec.Decode(&p); err != nil {
        log.Fatal(err)
    }
    enc.Encode(p)
}
```

**Expected output:**
```
{"id":1,"name":"Widget","price":9.99}
{"id":2,"name":"Gadget","price":24.5}
{"id":3,"name":"Doohickey","price":4.75}
```

**Checkpoint:** Each line is a single JSON object; the program handles the array without calling `json.Unmarshal` on the whole string.

---

### Lab 4: Custom MarshalJSON / UnmarshalJSON

**What you'll practise:** Implementing the `json.Marshaler` and `json.Unmarshaler` interfaces to serialise a `Duration` as a human-readable string.

**Task:**
Define a `Duration` type based on `time.Duration`. Implement `MarshalJSON` so it emits `"1h30m"` and `UnmarshalJSON` so it parses that string back. Embed `Duration` in a struct and round-trip it through JSON.

**Steps:**
1. Define `type Duration time.Duration`
2. Implement `MarshalJSON() ([]byte, error)` using `time.Duration.String()`
3. Implement `UnmarshalJSON(b []byte) error` using `time.ParseDuration`
4. Embed in a `Task` struct and marshal/unmarshal a value

```go
type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        return err
    }
    dur, err := time.ParseDuration(s)
    if err != nil {
        return err
    }
    *d = Duration(dur)
    return nil
}

type Task struct {
    Name    string   `json:"name"`
    Timeout Duration `json:"timeout"`
}
```

**Expected output:**
```
{"name":"build","timeout":"1h30m0s"}
Task: build, Timeout: 1h30m0s
```

**Checkpoint:** The JSON shows a string like `"1h30m0s"`, not a raw integer nanosecond count.

---

### Lab 5: CSV Reading into Typed Structs

**What you'll practise:** Using `encoding/csv` to read raw records and map them into typed structs, including handling quoted fields.

**Task:**
Given a CSV string with a header row, use `csv.NewReader` to read all records, skip the header, and convert each row into an `Employee` struct. Handle a field that contains a comma by quoting it in the input.

**Steps:**
1. Write a CSV string with header `name,department,salary`; include a row where the department is `"Engineering, Core"` (quoted)
2. Create a `csv.NewReader` around a `strings.NewReader`
3. Call `r.Read()` to skip the header, then loop with `r.Read()` for data rows
4. Parse each row into an `Employee` struct and print all employees

```go
const csvData = `name,department,salary
Alice,"Engineering, Core",95000
Bob,Marketing,72000
Carol,Sales,68000
`

type Employee struct {
    Name       string
    Department string
    Salary     int
}

r := csv.NewReader(strings.NewReader(csvData))
r.Read() // skip header
for {
    rec, err := r.Read()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    salary, _ := strconv.Atoi(rec[2])
    e := Employee{Name: rec[0], Department: rec[1], Salary: salary}
    fmt.Printf("%+v\n", e)
}
```

**Expected output:**
```
{Name:Alice Department:Engineering, Core Salary:95000}
{Name:Bob Department:Marketing Salary:72000}
{Name:Carol Department:Sales Salary:68000}
```

**Checkpoint:** The quoted comma-containing department parses correctly as a single field.

---

### Lab 6: CSV Writing with a Header Row

**What you'll practise:** Using `csv.NewWriter` to write a header row and data rows, then calling `Flush` and checking for errors.

**Task:**
Take a slice of `Product` structs and write them to a `strings.Builder` as CSV with a header row. Print the result and verify it looks correct.

**Steps:**
1. Define a slice of `Product{Name, Price, Stock}`
2. Create a `csv.NewWriter` around a `strings.Builder`
3. Write the header row `[]string{"name","price","stock"}`
4. Loop over products, writing each as a string slice
5. Call `w.Flush()` and check `w.Error()`

```go
type Product struct {
    Name  string
    Price float64
    Stock int
}

products := []Product{
    {"Widget", 9.99, 100},
    {"Gadget", 24.50, 45},
    {"Doohickey", 4.75, 200},
}

var sb strings.Builder
w := csv.NewWriter(&sb)
w.Write([]string{"name", "price", "stock"})
for _, p := range products {
    w.Write([]string{p.Name, strconv.FormatFloat(p.Price, 'f', 2, 64), strconv.Itoa(p.Stock)})
}
w.Flush()
if err := w.Error(); err != nil {
    log.Fatal(err)
}
fmt.Print(sb.String())
```

**Expected output:**
```
name,price,stock
Widget,9.99,100
Gadget,24.50,45
Doohickey,4.75,200
```

**Checkpoint:** Output matches the expected CSV with the header row first and no trailing comma.

---

### Lab 7: JSON-to-CSV Converter

**What you'll practise:** Reading the same data in JSON format and writing it out as CSV — bridging the two encodings.

**Task:**
Write a function `convertJSONToCSV(jsonInput string, w io.Writer) error` that decodes a JSON array of `Record{Name, Value string}` objects and writes them as CSV (with header) to the provided writer.

**Steps:**
1. Define `Record{Name, Value string}` with JSON tags
2. Unmarshal the input JSON array into `[]Record`
3. Create a `csv.NewWriter` around the provided `io.Writer`
4. Write header `["name","value"]` then one row per record
5. Flush and return any error

```go
const jsonInput = `[
  {"name":"alpha","value":"1"},
  {"name":"beta","value":"2"},
  {"name":"gamma","value":"3"}
]`

func convertJSONToCSV(jsonInput string, w io.Writer) error {
    var records []struct {
        Name  string `json:"name"`
        Value string `json:"value"`
    }
    if err := json.Unmarshal([]byte(jsonInput), &records); err != nil {
        return err
    }
    cw := csv.NewWriter(w)
    cw.Write([]string{"name", "value"})
    for _, r := range records {
        cw.Write([]string{r.Name, r.Value})
    }
    cw.Flush()
    return cw.Error()
}
```

**Expected output:**
```
name,value
alpha,1
beta,2
gamma,3
```

**Checkpoint:** Running `convertJSONToCSV(jsonInput, os.Stdout)` produces valid CSV with the correct header.

---

### Final Lab: Config Manager

**What you'll practise:** Combining JSON encode/decode, custom marshaling, and CSV export in a single program.

**Task:**
Build a config manager that loads from a JSON file, merges environment variable overrides, saves the merged config back to JSON with indentation, and exports it as CSV. Use a `LogLevel` type with custom `MarshalJSON`/`UnmarshalJSON` that maps between `"debug"/"info"/"warn"/"error"` strings and integer constants.

**Steps:**
1. Define `Config{Host string, Port int, DatabaseURL string, LogLevel LogLevel, Features map[string]bool}`
2. Implement `LogLevel.MarshalJSON` (emit string) and `UnmarshalJSON` (parse string to int)
3. Load from `config.json` with `json.NewDecoder`; merge `os.Getenv` overrides
4. Save merged config with `json.MarshalIndent` to `config.out.json`
5. Export all fields as a two-column CSV (`key,value`) using `csv.NewWriter`

```go
type LogLevel int

const (
    LogDebug LogLevel = iota
    LogInfo
    LogWarn
    LogError
)

func (l LogLevel) MarshalJSON() ([]byte, error) {
    names := []string{"debug", "info", "warn", "error"}
    if int(l) >= len(names) {
        return nil, fmt.Errorf("unknown log level: %d", l)
    }
    return json.Marshal(names[l])
}

func (l *LogLevel) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        return err
    }
    names := map[string]LogLevel{"debug": LogDebug, "info": LogInfo, "warn": LogWarn, "error": LogError}
    v, ok := names[s]
    if !ok {
        return fmt.Errorf("unknown log level: %q", s)
    }
    *l = v
    return nil
}
```

**Expected output:**
```
config loaded: {Host:localhost Port:5432 LogLevel:1}
saved to config.out.json
CSV:
key,value
host,localhost
port,5432
log_level,info
```

**Checkpoint:** `config.out.json` contains human-readable indented JSON with `"log_level":"info"` (not an integer); the CSV has the correct header and one row per field.

---

## Day Project: Config Manager

Build a config manager that:
1. Defines a `Config` struct (server host/port, database URL, log level, feature flags)
2. Loads config from JSON file with `json.NewDecoder`
3. Merges environment variable overrides (`os.Getenv`)
4. Saves the merged config back to JSON with indentation
5. Exports the config as CSV with headers

Use custom `MarshalJSON`/`UnmarshalJSON` for a `LogLevel` type (`"debug"` ↔ integer).

**Extension ideas:** support YAML via `gopkg.in/yaml.v3`; add JSON schema validation.

## Official Documentation

- [`encoding/json`](https://pkg.go.dev/encoding/json) — `Marshal`, `Unmarshal`, `NewEncoder`, `NewDecoder`, `RawMessage`, struct tags (`json:"..."`)
- [`encoding/csv`](https://pkg.go.dev/encoding/csv) — `NewReader`, `NewWriter`, `Reader.ReadAll`, `Writer.Flush`
- [`os`](https://pkg.go.dev/os) — `Getenv` for environment variable overrides
- [`io`](https://pkg.go.dev/io) — `EOF`, `Reader` used with streaming decoders
- [`strings`](https://pkg.go.dev/strings) — `NewReader` used in CSV examples
- [`time`](https://pkg.go.dev/time) — `Time`, `Duration` used in struct fields and custom marshaling
- [Go Blog: JSON and Go](https://go.dev/blog/json) — struct tags, streaming, custom marshaling walkthrough
- [Language Spec — Struct tags](https://go.dev/ref/spec#Struct_types)
