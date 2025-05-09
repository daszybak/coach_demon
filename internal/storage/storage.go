package storage

import "time"

type FeedbackEntry struct {
	ProblemID            string    `bson:"problemID"`
	Timestamp            time.Time `bson:"timestamp"`
	Code                 string    `bson:"code,omitempty"`
	Thoughts             string    `bson:"thoughts,omitempty"`
	Feedback             string    `bson:"feedback,omitempty"`
	Proof                string    `bson:"proofs,omitempty"`
	OptimalMetaCognition string    `bson:"optimalMetaCognition,omitempty"`
}

type Summary struct {
	ProblemID            string `bson:"problemID"`
	Feedback             string `bson:"feedback"`
	Proof                string `bson:"proof"`
	OptimalMetaCognition string `bson:"optimalMetaCognition"`
}

type StatementEntry struct {
	ProblemID string `bson:"problemID"`
	Statement string `bson:"statement"`
}

type Storage interface {
	SaveFeedback(entry FeedbackEntry) error
	GetAllFeedbacksByProblemID(problemID string) ([]FeedbackEntry, error)
	GetLatestFeedback(problemID string) (*FeedbackEntry, error)

	GetStatement(problemID string) (*StatementEntry, error)
	SaveStatement(entry StatementEntry) error
	GetAllStatements() ([]StatementEntry, error)

	GetSummaryByProblemID(problemID string) (*Summary, error)
	SaveSummary(summary Summary) error
}
