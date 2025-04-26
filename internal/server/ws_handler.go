package server

import (
	"coach_demon/internal/app"
	"coach_demon/internal/storage"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type EditorMessage struct {
	ProblemID string `json:"problemId"`
	Code      string `json:"code"`
	Thoughts  string `json:"thoughts"`
}

var upgrader = websocket.Upgrader{}

func makeWSHandler(ctx *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.Logger.Printf("couldn't upgrade WebSocket: %v", err)
			return
		}
		defer conn.Close()

		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				ctx.Logger.Printf("WebSocket read error: %v", err)
				return
			}

			var in EditorMessage
			if err := json.Unmarshal(raw, &in); err != nil {
				ctx.Logger.Printf("couldn't parse JSON: %v", err)
				continue
			}

			now := time.Now().UTC()

			// 🧠 Load or fetch problem statement
			statement, err := ctx.Store.GetStatement(in.ProblemID)
			if err != nil {
				ctx.Logger.Printf("load statement error: %v", err)
				continue
			}
			if statement == "" {
				ctx.Logger.Printf("statement not found in DB, fetching online for %s", in.ProblemID)
				statement, err = ctx.Fetch.Fetch(r.Context(), in.ProblemID)
				if err != nil {
					ctx.Logger.Printf("fetch statement error %s: %v", in.ProblemID, err)
					continue
				}
				ctx.Logger.Printf("fetched statement for %s: %s", in.ProblemID, trimForLog(statement))
				if err := ctx.Store.SaveStatement(in.ProblemID, statement); err != nil {
					ctx.Logger.Printf("save statement error: %v", err)
				}
			} else {
				ctx.Logger.Printf("loaded cached statement for %s: %s", in.ProblemID, trimForLog(statement))
			}

			// 📦 Always snapshot user code and thoughts
			err = ctx.Store.SaveFeedback(storage.FeedbackEntry{
				ProblemID: in.ProblemID,
				Timestamp: now,
				Code:      in.Code,
				Thoughts:  in.Thoughts,
			})
			if err != nil {
				ctx.Logger.Printf("store code snapshot: %v", err)
			}

			// 🧠 OpenAI feedback every minute
			latest, err := ctx.Store.GetLatestFeedback(in.ProblemID)
			if err != nil {
				ctx.Logger.Printf("fetch latest feedback: %v", err)
				continue
			}
			if latest == nil || time.Since(latest.Timestamp) > time.Minute {
				ctx.Logger.Printf("requesting OpenAI feedback for %s", in.ProblemID)
				fb, err := ctx.AI.GetFeedback(in.Code, in.Thoughts, statement)
				if err != nil {
					ctx.Logger.Printf("openai error: %v", err)
					continue
				}
				now = time.Now().UTC()
				err = ctx.Store.SaveFeedback(storage.FeedbackEntry{
					ProblemID:            in.ProblemID,
					Timestamp:            now,
					Statement:            statement,
					Code:                 in.Code,
					Thoughts:             in.Thoughts,
					Feedback:             fb.Feedback,
					Suggestions:          fb.Suggestions,
					Proofs:               fb.Proofs,
					OptimalMetaCognition: fb.OptimalMetaCognition,
				})
				if err != nil {
					ctx.Logger.Printf("store OpenAI feedback: %v", err)
				}
			}
		}
	}
}

func trimForLog(s string) string {
	const MAX = 200
	if len(s) > MAX {
		return s[:MAX] + "..."
	}
	return s
}
