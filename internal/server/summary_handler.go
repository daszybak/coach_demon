package server

import (
	"coach_demon/internal/app"
	"coach_demon/internal/storage"
	"encoding/json"
	"net/http"
)

func getSummary(ctx *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		problemID := r.URL.Query().Get("problemId")
		if problemID == "" {
			http.Error(w, "missing problemId in query", http.StatusBadRequest)
			return
		}

		summary, err := ctx.Store.GetSummaryByProblemID(problemID)
		if summary != nil {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(summary); err != nil {
				ctx.Logger.Error().Msgf("failed to encode summary response: %v", err)
				http.Error(w, "internal error during encoding summary", http.StatusInternalServerError)
				return
			}
		}

		// 1️⃣ Fetch problem statement from database
		statement, err := ctx.Store.GetStatement(problemID)
		if err != nil {
			ctx.Logger.Error().Msgf("failed to get statement for %s: %v", problemID, err)
			http.Error(w, "internal error fetching statement", http.StatusInternalServerError)
			return
		}
		if statement == nil {
			http.Error(w, "no statement found for this problem", http.StatusNotFound)
			return
		}

		// 2️⃣ Fetch all feedback entries
		entries, err := ctx.Store.GetAllFeedbacksByProblemID(problemID)
		if err != nil {
			ctx.Logger.Error().Msgf("failed to get all feedbacks for problem ID %s: %v", problemID, err)
			http.Error(w, "internal error fetching feedbacks", http.StatusInternalServerError)
			return
		}
		if len(entries) == 0 {
			http.Error(w, "no feedbacks found for this problem ID", http.StatusNotFound)
			return
		}

		// 3️⃣ Prepare all code + thoughts snapshots
		var feedbacks []string
		var proofs []string
		var optimalMetaCognitions []string
		for _, entry := range entries {
			if entry.Feedback != "" {
				feedbacks = append(feedbacks, entry.Feedback)
			}
			if entry.Feedback != "" {
				proofs = append(proofs, entry.Proof)
			}
			if entry.OptimalMetaCognition != "" {
				optimalMetaCognitions = append(optimalMetaCognitions, entry.OptimalMetaCognition)
			}
		}

		// 4️⃣ Call OpenAI to get a nice summary
		openAISummary, err := ctx.AI.SummarizeFeedback(statement.Statement, feedbacks, proofs, optimalMetaCognitions)
		if err != nil {
			ctx.Logger.Error().Msgf("failed to summarize history for %s: %v", problemID, err)
			http.Error(w, "internal error during summarization", http.StatusInternalServerError)
			return
		}

		err = ctx.Store.SaveSummary(storage.Summary{
			Feedback:             openAISummary.Feedback,
			OptimalMetaCognition: openAISummary.OptimalMetaCognition,
			Proof:                openAISummary.Proof,
		})
		if err != nil {
			ctx.Logger.Error().Msgf("failed to store summary for %s: %v", problemID, err)
		}

		// 5️⃣ Respond to client
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(summary); err != nil {
			ctx.Logger.Error().Msgf("failed to encode summary response: %v", err)
			http.Error(w, "internal error during encoding summary", http.StatusInternalServerError)
		}
	}
}
