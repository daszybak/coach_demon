package server

import (
	"coach_demon/internal/app"
	"coach_demon/internal/storage"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type EditorMessage struct {
	ProblemID string `json:"problemId"`
	Code      string `json:"code"`
	Thoughts  string `json:"thoughts"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow any frontend
}

func makeWSHandler(ctx *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			ctx.Logger.Info().Err(err).Msg("could not upgrade WebSocket")
			return
		}
		defer conn.Close()

		ctx.Logger.Info().Msg("WebSocket connection established")

		for {
			_, raw, err := conn.ReadMessage()
			if err != nil {
				var closeErr *websocket.CloseError
				ok := errors.As(err, &closeErr)
				if ok {
					ctx.Logger.Info().Int("code", closeErr.Code).Str("text", closeErr.Text).Msg("WebSocket closed normally")
				} else {
					ctx.Logger.Warn().Err(err).Msg("WebSocket temporary read error")
				}
				time.Sleep(1 * time.Second) // wait a bit before next read
				continue
			}

			var in EditorMessage
			if err := json.Unmarshal(raw, &in); err != nil {
				ctx.Logger.Warn().Err(err).Msg("could not parse incoming JSON")
				continue
			}

			statement, err := ctx.Store.GetStatement(in.ProblemID)
			if err != nil {
				ctx.Logger.Warn().Err(err).Msg("load statement error")
				continue
			}
			if statement == nil {
				ctx.Logger.Info().Msgf("fetching missing statement for %s", in.ProblemID)
				codeforcesStatement, err := ctx.Fetch.Fetch(r.Context(), in.ProblemID)
				if err != nil {
					ctx.Logger.Warn().
						Err(err).
						Str("problemId", in.ProblemID).
						Msg("fetch error")
					continue
				}
				if codeforcesStatement == "" {
					ctx.Logger.Warn().
						Str("problemId", in.ProblemID).
						Msg("fetched empty statement from codeforces")
					continue
				}
				_ = ctx.Store.SaveStatement(storage.StatementEntry{
					Statement: codeforcesStatement,
					ProblemID: in.ProblemID,
				})
			}

			latest, _ := ctx.Store.GetLatestFeedback(in.ProblemID)
			if latest == nil || time.Since(latest.Timestamp) > time.Minute {
				ctx.Logger.Info().Msgf("asking OpenAI for new feedback for %s", in.ProblemID)
				fb, err := ctx.AI.GetFeedback(in.Code, in.Thoughts, statement.Statement)
				if err != nil {
					ctx.Logger.Warn().Err(err).Msg("OpenAI feedback error")
					continue
				}
				err = ctx.Store.SaveFeedback(storage.FeedbackEntry{
					ProblemID:            in.ProblemID,
					Timestamp:            time.Now().UTC(),
					Code:                 in.Code,
					Thoughts:             in.Thoughts,
					Feedback:             fb.Feedback,
					Proof:                fb.Proof,
					OptimalMetaCognition: fb.OptimalMetaCognition,
				})
				if err != nil {
					ctx.Logger.Warn().Err(err).Msg("saving OpenAI feedback failed")
				}
			}
		}
	}
}
