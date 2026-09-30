// Day 34 – Concurrent Download Manager
//
// Demonstrates:
//   - Semaphore pattern via a buffered channel to cap concurrency
//   - sync.WaitGroup for fan-out coordination
//   - sync.Mutex to protect a shared results map
//   - sync/atomic for a lock-free total-bytes counter
//   - net/http/httptest.Server so the demo runs without real network access
package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// ── Result type ───────────────────────────────────────────────────────────────

// DownloadResult records the outcome of one download attempt.
type DownloadResult struct {
	URL      string
	FilePath string
	Bytes    int64
	Err      error
	Duration time.Duration
}

// ── Semaphore ─────────────────────────────────────────────────────────────────

// semaphore is a counting semaphore implemented with a buffered channel of
// empty structs. Acquire blocks when the limit is reached; Release frees a
// slot.
type semaphore chan struct{}

func newSemaphore(n int) semaphore { return make(chan struct{}, n) }
func (s semaphore) Acquire()       { s <- struct{}{} }
func (s semaphore) Release()       { <-s }

// ── Downloader ────────────────────────────────────────────────────────────────

// Manager orchestrates concurrent downloads.
type Manager struct {
	concurrency int
	outDir      string
	client      *http.Client
}

// NewManager creates a Manager that downloads at most concurrency files in
// parallel, saving them under outDir.
func NewManager(concurrency int, outDir string) *Manager {
	return &Manager{
		concurrency: concurrency,
		outDir:      outDir,
		client:      &http.Client{Timeout: 30 * time.Second},
	}
}

// Download fetches all URLs concurrently (up to m.concurrency at a time).
// It returns a map of URL → DownloadResult plus aggregate statistics.
func (m *Manager) Download(urls []string) (map[string]DownloadResult, int64, int, int) {
	sem := newSemaphore(m.concurrency)
	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		results    = make(map[string]DownloadResult, len(urls))
		totalBytes int64 // atomic
		successful int32 // atomic
		failed     int32 // atomic
	)

	for _, rawURL := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			sem.Acquire()
			defer sem.Release()

			start := time.Now()
			res := m.downloadOne(u)
			res.Duration = time.Since(start)

			// Update shared state
			atomic.AddInt64(&totalBytes, res.Bytes)
			if res.Err != nil {
				atomic.AddInt32(&failed, 1)
			} else {
				atomic.AddInt32(&successful, 1)
			}

			mu.Lock()
			results[u] = res
			mu.Unlock()
		}(rawURL)
	}

	wg.Wait()
	return results, atomic.LoadInt64(&totalBytes),
		int(atomic.LoadInt32(&successful)),
		int(atomic.LoadInt32(&failed))
}

// downloadOne performs a single HTTP GET and writes the body to a temp file.
func (m *Manager) downloadOne(url string) DownloadResult {
	res := DownloadResult{URL: url}

	resp, err := m.client.Get(url)
	if err != nil {
		res.Err = fmt.Errorf("GET %s: %w", url, err)
		return res
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		res.Err = fmt.Errorf("GET %s: unexpected status %d", url, resp.StatusCode)
		return res
	}

	// Write to a temp file in outDir
	tmp, err := os.CreateTemp(m.outDir, "download-*.tmp")
	if err != nil {
		res.Err = fmt.Errorf("create temp: %w", err)
		return res
	}
	defer tmp.Close()

	n, err := io.Copy(tmp, resp.Body)
	if err != nil {
		res.Err = fmt.Errorf("copy body: %w", err)
		return res
	}

	res.FilePath = tmp.Name()
	res.Bytes = n
	return res
}

// ── Demo test server ──────────────────────────────────────────────────────────

// startTestServer creates a local HTTP server that serves files of varying
// sizes so the demo works without a real network connection.
func startTestServer() *httptest.Server {
	mux := http.NewServeMux()

	// Register 10 endpoints with different payloads
	sizes := []int{512, 1024, 2048, 4096, 8192, 512, 16384, 1024, 2048, 4096}
	for i, size := range sizes {
		i, size := i, size // capture
		path := fmt.Sprintf("/file%d", i+1)
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/octet-stream")
			payload := make([]byte, size)
			for j := range payload {
				payload[j] = byte('A' + (i % 26))
			}
			w.Write(payload)
		})
	}
	return httptest.NewServer(mux)
}

// ── main ──────────────────────────────────────────────────────────────────────

func main() {
	const concurrency = 3 // at most 3 simultaneous downloads

	// Start local test server
	ts := startTestServer()
	defer ts.Close()

	// Build list of 10 URLs against the test server
	urls := make([]string, 10)
	for i := range urls {
		urls[i] = fmt.Sprintf("%s/file%d", ts.URL, i+1)
	}

	// Output directory for downloaded files
	outDir, err := os.MkdirTemp("", "downloads-*")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(outDir) // clean up after the demo

	fmt.Printf("=== Day 34: Concurrent Download Manager ===\n\n")
	fmt.Printf("Concurrency limit : %d\n", concurrency)
	fmt.Printf("URLs to download  : %d\n", len(urls))
	fmt.Printf("Output directory  : %s\n\n", filepath.Base(outDir))

	start := time.Now()
	mgr := NewManager(concurrency, outDir)
	results, totalBytes, successful, failed := mgr.Download(urls)
	elapsed := time.Since(start)

	// Print per-file results
	fmt.Printf("%-45s  %-8s  %-10s  %s\n", "URL", "Bytes", "Duration", "Status")
	fmt.Printf("%-45s  %-8s  %-10s  %s\n",
		"---------------------------------------------", "--------", "----------", "------")
	for _, u := range urls {
		r := results[u]
		status := "OK"
		if r.Err != nil {
			status = "FAIL: " + r.Err.Error()
		}
		fmt.Printf("%-45s  %-8d  %-10s  %s\n",
			u[len(ts.URL):], r.Bytes, r.Duration.Round(time.Millisecond), status)
	}

	// Summary
	fmt.Printf("\n── Summary ──────────────────────────────────────────────────\n")
	fmt.Printf("  Total bytes   : %d\n", totalBytes)
	fmt.Printf("  Successful    : %d\n", successful)
	fmt.Printf("  Failed        : %d\n", failed)
	fmt.Printf("  Time elapsed  : %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("─────────────────────────────────────────────────────────────\n")
}
