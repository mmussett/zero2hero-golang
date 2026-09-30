package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		fmt.Printf("%s %s %v\n", r.Method, r.URL.Path, time.Since(start).Round(time.Microsecond))
	})
}

func setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /echo", func(w http.ResponseWriter, r *http.Request) {
		msg := r.URL.Query().Get("msg")
		if msg == "" {
			msg = "hello"
		}
		writeJSON(w, http.StatusOK, map[string]string{"echo": msg})
	})

	mux.HandleFunc("POST /echo", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON"})
			return
		}
		writeJSON(w, http.StatusOK, payload)
	})

	mux.HandleFunc("GET /headers", func(w http.ResponseWriter, r *http.Request) {
		headers := make(map[string]string)
		for k, v := range r.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}
		writeJSON(w, http.StatusOK, headers)
	})

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return logging(mux)
}

func runServer(addr string) {
	srv := &http.Server{
		Addr:         addr,
		Handler:      setupRoutes(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		fmt.Printf("Server listening on http://%s\n", addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("\nShutting down gracefully...")
	srv.Shutdown(ctx)
}

func runClient(base string) {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(base + "/echo?msg=hello+world")
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("GET /echo:    %s", body)
	}

	resp, err = client.Get(base + "/health")
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("GET /health:  %s", body)
	}

	resp, err = client.Get(base + "/headers")
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("GET /headers: %s", body)
	}
}

func main() {
	addr   := flag.String("addr", "localhost:8080", "server address")
	client := flag.Bool("client", false, "run as HTTP client instead of server")
	flag.Parse()

	if *client {
		runClient("http://" + *addr)
	} else {
		runServer(*addr)
	}
}
