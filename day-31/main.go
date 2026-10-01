package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
)

// ── Data model ───────────────────────────────────────────────────────────────

// Bookmark represents a saved URL with metadata.
type Bookmark struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

// ── Storage ───────────────────────────────────────────────────────────────────

func storePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".bookmarks.json")
}

func load() ([]Bookmark, error) {
	path := storePath()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var bm []Bookmark
	if err := json.Unmarshal(data, &bm); err != nil {
		return nil, err
	}
	return bm, nil
}

func save(bm []Bookmark) error {
	data, err := json.MarshalIndent(bm, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(storePath(), data, 0o600)
}

func nextID(bm []Bookmark) int {
	max := 0
	for _, b := range bm {
		if b.ID > max {
			max = b.ID
		}
	}
	return max + 1
}

// ── Commands ──────────────────────────────────────────────────────────────────

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bookmarks",
		Short: "URL bookmark manager",
	}

	// ── add ──────────────────────────────────────────────────────────────────
	var addTitle string
	addCmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a new bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bm, err := load()
			if err != nil {
				return err
			}
			url := args[0]
			title := addTitle
			if title == "" {
				title = url
			}
			b := Bookmark{
				ID:        nextID(bm),
				URL:       url,
				Title:     title,
				CreatedAt: time.Now(),
			}
			bm = append(bm, b)
			if err := save(bm); err != nil {
				return err
			}
			fmt.Printf("Added bookmark #%d: %s\n", b.ID, b.Title)
			return nil
		},
	}
	addCmd.Flags().StringVar(&addTitle, "title", "", "Human-readable title for the bookmark")

	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all bookmarks",
		RunE: func(cmd *cobra.Command, args []string) error {
			bm, err := load()
			if err != nil {
				return err
			}
			if len(bm) == 0 {
				fmt.Println("No bookmarks yet. Use 'bookmarks add <url>' to add one.")
				return nil
			}
			fmt.Printf("%-4s  %-40s  %s\n", "ID", "Title", "URL")
			fmt.Printf("%-4s  %-40s  %s\n", "----", "----------------------------------------", "----")
			for _, b := range bm {
				title := b.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				fmt.Printf("%-4d  %-40s  %s\n", b.ID, title, b.URL)
			}
			return nil
		},
	}

	// ── delete ───────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a bookmark by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			bm, err := load()
			if err != nil {
				return err
			}
			newBm := bm[:0]
			found := false
			for _, b := range bm {
				if b.ID == id {
					found = true
					fmt.Printf("Deleted bookmark #%d: %s\n", b.ID, b.Title)
					continue
				}
				newBm = append(newBm, b)
			}
			if !found {
				return fmt.Errorf("bookmark #%d not found", id)
			}
			return save(newBm)
		},
	}

	// ── open ─────────────────────────────────────────────────────────────────
	openCmd := &cobra.Command{
		Use:   "open <id>",
		Short: "Print the URL for a bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid id %q: %w", args[0], err)
			}
			bm, err := load()
			if err != nil {
				return err
			}
			for _, b := range bm {
				if b.ID == id {
					fmt.Println(b.URL)
					return nil
				}
			}
			return fmt.Errorf("bookmark #%d not found", id)
		},
	}

	// ── serve ─────────────────────────────────────────────────────────────────
	var serveAddr string
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start a web server that lists bookmarks as HTML",
		RunE: func(cmd *cobra.Command, args []string) error {
			r := chi.NewRouter()
			r.Use(middleware.Logger)
			r.Use(middleware.Recoverer)

			r.Get("/", func(w http.ResponseWriter, req *http.Request) {
				bm, err := load()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				if err := bookmarksTmpl.Execute(w, bm); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			})

			r.Get("/api/bookmarks", func(w http.ResponseWriter, req *http.Request) {
				bm, err := load()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(bm)
			})

			fmt.Printf("Bookmark server listening on http://%s\n", serveAddr)
			return http.ListenAndServe(serveAddr, r)
		},
	}
	serveCmd.Flags().StringVar(&serveAddr, "addr", ":8080", "Address to listen on")

	root.AddCommand(addCmd, listCmd, deleteCmd, openCmd, serveCmd)
	return root
}

var bookmarksTmpl = template.Must(template.New("bookmarks").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>Bookmarks</title>
  <style>
    body { font-family: sans-serif; max-width: 800px; margin: 2rem auto; padding: 0 1rem; }
    table { width: 100%; border-collapse: collapse; }
    th, td { text-align: left; padding: 0.5rem 1rem; border-bottom: 1px solid #ddd; }
    th { background: #f4f4f4; }
    a { color: #0066cc; }
    .id { width: 3rem; color: #888; }
    .date { white-space: nowrap; color: #888; font-size: 0.85em; }
  </style>
</head>
<body>
  <h1>Bookmarks</h1>
  {{if .}}
  <table>
    <thead>
      <tr><th class="id">#</th><th>Title</th><th>URL</th><th>Added</th></tr>
    </thead>
    <tbody>
    {{range .}}
      <tr>
        <td class="id">{{.ID}}</td>
        <td>{{.Title}}</td>
        <td><a href="{{.URL}}" target="_blank">{{.URL}}</a></td>
        <td class="date">{{.CreatedAt.Format "2006-01-02"}}</td>
      </tr>
    {{end}}
    </tbody>
  </table>
  {{else}}
  <p>No bookmarks yet.</p>
  {{end}}
</body>
</html>
`))

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
