package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

// CountingReader wraps io.Reader and counts bytes read.
type CountingReader struct {
	r     io.Reader
	count int64
}

func NewCountingReader(r io.Reader) *CountingReader { return &CountingReader{r: r} }

func (c *CountingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.count += int64(n)
	return n, err
}

func (c *CountingReader) BytesRead() int64 { return c.count }

// TeeWriter writes every Write call to two writers simultaneously.
type TeeWriter struct {
	a, b io.Writer
}

func NewTeeWriter(a, b io.Writer) *TeeWriter { return &TeeWriter{a, b} }

func (t *TeeWriter) Write(p []byte) (int, error) {
	n, err := t.a.Write(p)
	if err != nil {
		return n, err
	}
	_, err = t.b.Write(p)
	return n, err
}

// Temperature implements fmt.Stringer.
type Temperature float64

func (t Temperature) String() string { return fmt.Sprintf("%.1f°C", float64(t)) }

// Temperatures implements sort.Interface.
type Temperatures []Temperature

func (t Temperatures) Len() int           { return len(t) }
func (t Temperatures) Less(i, j int) bool { return t[i] < t[j] }
func (t Temperatures) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

// Config implements encoding.TextMarshaler / TextUnmarshaler.
type Config struct {
	Host  string
	Port  int
	Debug bool
}

func (c Config) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%s:%d:debug=%v", c.Host, c.Port, c.Debug)), nil
}

func (c *Config) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%s", &c.Host)
	return err
}

func main() {
	fmt.Println("=== CountingReader ===")
	src := strings.NewReader("Hello, Go interfaces!")
	cr := NewCountingReader(src)
	io.Copy(os.Stdout, cr)
	fmt.Printf("\nBytes read: %d\n\n", cr.BytesRead())

	fmt.Println("=== TeeWriter ===")
	var buf bytes.Buffer
	tw := NewTeeWriter(os.Stdout, &buf)
	fmt.Fprintln(tw, "written to stdout AND buffer")
	fmt.Printf("Buffer contains: %q\n\n", buf.String())

	fmt.Println("=== Temperature Sort ===")
	temps := Temperatures{23.5, 17.2, 31.0, 8.4, 25.1}
	fmt.Printf("Before:   %v\n", temps)
	sort.Sort(temps)
	fmt.Printf("Asc:      %v\n", temps)
	sort.Sort(sort.Reverse(temps))
	fmt.Printf("Desc:     %v\n\n", temps)

	fmt.Println("=== Config TextMarshaler ===")
	cfg := Config{Host: "localhost", Port: 8080, Debug: true}
	b, _ := cfg.MarshalText()
	fmt.Printf("Marshalled: %s\n", b)
}
