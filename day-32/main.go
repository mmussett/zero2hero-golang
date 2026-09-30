// Day 32 – Idiomatic Go Project Structure
//
// This program demonstrates domain-driven layout by wiring together
// independent packages from internal/:
//
//   internal/notes     – Note domain type, Store interface, MemoryStore
//   internal/notes     – HTTP handlers (handler.go)
//   internal/platform  – HTTP server with graceful shutdown
//
// Run:
//
//	go run . [-addr :8080]
//
// Example requests:
//
//	curl -X POST -d '{"title":"Buy milk","body":"Organic 2%"}' http://localhost:8080/notes
//	curl http://localhost:8080/notes
//	curl http://localhost:8080/notes/1
//	curl -X PUT  -d '{"title":"Buy oat milk","body":""}' http://localhost:8080/notes/1
//	curl -X DELETE http://localhost:8080/notes/1
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/mmussett/zero2hero-golang/day-32/internal/notes"
	"github.com/mmussett/zero2hero-golang/day-32/internal/platform"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	// Domain: in-memory note store
	store := notes.NewMemoryStore()

	// HTTP layer
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Notes sub-routes
	h := notes.NewHandler(store)
	h.Routes(r)

	// Seed a couple of demo notes so the API is not empty on first run
	if _, err := store.Create("Welcome to Day 32", "This is the Notes API restructured with domain-driven layout."); err != nil {
		log.Println("seed note 1:", err)
	}
	if _, err := store.Create("Go project structure", "internal/, cmd/, platform/ — keep concerns separate."); err != nil {
		log.Println("seed note 2:", err)
	}

	fmt.Printf("Notes API\n  addr  : http://localhost%s\n  routes: GET/POST /notes, GET/PUT/DELETE /notes/{id}\n\n", *addr)

	srv := platform.NewServer(*addr, r)
	if err := srv.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
