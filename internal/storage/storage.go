package storage

import "time"

// FeedbackEntry is your per-minute snapshot.
type FeedbackEntry struct {
	ProblemID            string    `bson:"problemid"`
	Timestamp            time.Time `bson:"timestamp"`
	Code                 string    `bson:"code,omitempty"`
	Thoughts             string    `bson:"thoughts,omitempty"`
	Feedback             string    `bson:"feedback,omitempty"`
	Statement            string    `bson:"statement,omitempty"`
	Suggestions          []string  `bson:"suggestions,omitempty"`
	Proofs               []string  `bson:"proofs,omitempty"`
	OptimalMetaCognition []string  `bson:"optimalmetacognition,omitempty"`
}

type Storage interface {
	SaveFeedback(entry FeedbackEntry) error
	SummarizeFeedback(problemID string) ([]FeedbackEntry, error)
	GetLatestFeedback(problemID string) (*FeedbackEntry, error)

	GetStatement(problemID string) (string, error)
	SaveStatement(problemID, statement string) error
}
