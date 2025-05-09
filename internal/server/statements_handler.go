package server

import (
	"coach_demon/internal/app"
	"encoding/json"
	"net/http"
)

func getStatements(ctx *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statements, err := ctx.Store.GetAllStatements()
		if err != nil {
			ctx.Logger.Error().Msgf("failed to get all statements: %v", err)
			http.Error(w, "internal error fetching statements", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(statements); err != nil {
			ctx.Logger.Error().Msgf("failed to encode statements: %v", err)
			http.Error(w, "internal error encoding statements", http.StatusInternalServerError)
			return
		}
	}
}
