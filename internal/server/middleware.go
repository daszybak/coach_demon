package server

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"time"
)

// With attaches a slice of middlewares to a router in one call.
func With(r chi.Router, mws ...func(http.Handler) http.Handler) {
	for _, m := range mws {
		r.Use(m)
	}
}

// CORS middleware
func CORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return reqID
	}
	return "-"
}

// NewLogger returns a logger middleware using the provided logger
func NewLogger(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()

			next.ServeHTTP(ww, r)

			dur := time.Since(start)
			reqID := getRequestID(r.Context())

			logger.Printf("[%s] %s %s %d %s", reqID, r.Method, r.URL.Path, ww.Status(), dur)
		})
	}
}
