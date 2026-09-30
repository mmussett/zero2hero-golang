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

## Day Project: Config Manager

Build a config manager that:
1. Defines a `Config` struct (server host/port, database URL, log level, feature flags)
2. Loads config from JSON file with `json.NewDecoder`
3. Merges environment variable overrides (`os.Getenv`)
4. Saves the merged config back to JSON with indentation
5. Exports the config as CSV with headers

Use custom `MarshalJSON`/`UnmarshalJSON` for a `LogLevel` type (`"debug"` ↔ integer).

**Extension ideas:** support YAML via `gopkg.in/yaml.v3`; add JSON schema validation.
