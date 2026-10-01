package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// --- Typed context keys ---

type contextKey int

const (
	keyRequestID contextKey = iota
	keyUser
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, keyRequestID, id)
}

func requestID(ctx context.Context) string {
	id, _ := ctx.Value(keyRequestID).(string)
	return id
}

func withUser(ctx context.Context, user string) context.Context {
	return context.WithValue(ctx, keyUser, user)
}

func userFromCtx(ctx context.Context) string {
	u, _ := ctx.Value(keyUser).(string)
	return u
}

// --- Middleware ---

func generateID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generateID()
		ctx := withRequestID(r.Context(), id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func loggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestID(r.Context()),
			)
		})
	}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if !strings.HasPrefix(token, "Bearer ") || len(token) <= 7 {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		user := strings.TrimPrefix(token, "Bearer ")
		ctx := withUser(r.Context(), user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// --- Pipeline ---

type Stage func(ctx context.Context, input string) (string, error)

func runPipeline(ctx context.Context, stages []Stage, input string) (string, error) {
	result := input
	for _, stage := range stages {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
		var err error
		result, err = stage(ctx, result)
		if err != nil {
			return "", err
		}
	}
	return result, nil
}

func validateStage(ctx context.Context, input string) (string, error) {
	logger := slog.Default()
	logger.Info("validate stage", "request_id", requestID(ctx), "len", len(input))
	if strings.TrimSpace(input) == "" {
		return "", errors.New("input cannot be empty")
	}
	time.Sleep(200 * time.Millisecond) // simulate work
	return strings.TrimSpace(input), nil
}

func enrichStage(ctx context.Context, input string) (string, error) {
	slog.Default().Info("enrich stage", "request_id", requestID(ctx))
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(300 * time.Millisecond):
	}
	user := userFromCtx(ctx)
	return fmt.Sprintf("[enriched by %s] %s", user, input), nil
}

func storeStage(ctx context.Context, input string) (string, error) {
	slog.Default().Info("store stage", "request_id", requestID(ctx))
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(200 * time.Millisecond):
	}
	return input, nil
}

// --- Handlers ---

type reportRequest struct {
	Data string `json:"data"`
}

type reportResponse struct {
	Result    string `json:"result"`
	RequestID string `json:"request_id"`
	User      string `json:"user"`
}

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id"`
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// enforce 5-second processing timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error(), RequestID: requestID(ctx)})
		return
	}

	stages := []Stage{validateStage, enrichStage, storeStage}
	result, err := runPipeline(ctx, stages, req.Data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		status := http.StatusInternalServerError
		if errors.Is(err, context.DeadlineExceeded) {
			status = http.StatusGatewayTimeout
		} else if errors.Is(err, context.Canceled) {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(errorResponse{Error: err.Error(), RequestID: requestID(ctx)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reportResponse{
		Result:    result,
		RequestID: requestID(ctx),
		User:      userFromCtx(ctx),
	})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/reports", authMiddleware(http.HandlerFunc(reportHandler)))

	handler := requestIDMiddleware(loggingMiddleware(logger)(mux))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "err", err)
	}
	logger.Info("server stopped")
}
