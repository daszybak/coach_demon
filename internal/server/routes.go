package server

import (
	"coach_demon/internal/app"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func New(ctx *app.App) http.Handler {
	r := chi.NewRouter()
	With(r,
		middleware.Recoverer, // optional
		RequestID,
		NewLogger(ctx.Logger), // 👈 using your App logger!
		CORS,
	)

	r.Get("/summary/{problemId}", makeSummaryHandler(ctx))
	r.Handle("/ws", makeWSHandler(ctx))
	return r
}
