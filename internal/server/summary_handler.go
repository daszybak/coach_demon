package server

import (
	"coach_demon/internal/app"
	"encoding/json"
	"net/http"
)

func makeSummaryHandler(ctx *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		problemID := r.URL.Query().Get("problemId")
		if problemID == "" {
			http.Error(w, "missing problemId in query", http.StatusBadRequest)
			return
		}

		// 1️⃣ Fetch problem statement from database
		statement, err := ctx.Store.GetStatement(problemID)
		if err != nil {
			ctx.Logger.Printf("failed to get statement for %s: %v", problemID, err)
			http.Error(w, "internal error fetching statement", http.StatusInternalServerError)
			return
		}
		if statement == "" {
			http.Error(w, "no statement found for this problem", http.StatusNotFound)
			return
		}

		// 2️⃣ Fetch all feedback entries
		entries, err := ctx.Store.SummarizeFeedback(problemID)
		if err != nil {
			ctx.Logger.Printf("failed to summarize feedbacks for %s: %v", problemID, err)
			http.Error(w, "internal error fetching feedbacks", http.StatusInternalServerError)
			return
		}
		if len(entries) == 0 {
			http.Error(w, "no feedbacks found for this problem", http.StatusNotFound)
			return
		}

		// 3️⃣ Prepare all code + thoughts snapshots
		var codeHistory []string
		var thoughtsHistory []string
		for _, entry := range entries {
			if entry.Code != "" {
				codeHistory = append(codeHistory, entry.Code)
			}
			if entry.Thoughts != "" {
				thoughtsHistory = append(thoughtsHistory, entry.Thoughts)
			}
		}

		// 4️⃣ Call OpenAI to get a nice summary
		summary, err := ctx.AI.SummarizeHistory(statement, codeHistory, thoughtsHistory)
		if err != nil {
			ctx.Logger.Printf("failed to summarize history for %s: %v", problemID, err)
			http.Error(w, "internal error during summarization", http.StatusInternalServerError)
			return
		}

		// 5️⃣ Respond to client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summary); err != nil {
			ctx.Logger.Printf("failed to encode summary response: %v", err)
		}
	}
}
