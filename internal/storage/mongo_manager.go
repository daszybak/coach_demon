package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoManager struct {
	feedbacks  *mongo.Collection
	summaries  *mongo.Collection
	statements *mongo.Collection
	logger     *zerolog.Logger
}

func NewMongoManager(uri string, logger *zerolog.Logger) (*MongoManager, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("cannot ping MongoDB: %w", err)
	}

	db := client.Database("coach_demon")
	logger.Info().Msg("connected to MongoDB")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "problemid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = db.Collection("summaries").Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create unique index on summaries: %w", err)
	}

	_, err = db.Collection("statements").Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create unique index on statements: %w", err)
	}

	return &MongoManager{
		feedbacks:  db.Collection("feedbacks"),
		statements: db.Collection("statements"),
		summaries:  db.Collection("summaries"),
		logger:     logger,
	}, nil
}

func (m *MongoManager) SaveFeedback(entry FeedbackEntry) error {
	_, err := m.feedbacks.InsertOne(context.Background(), entry)
	if err != nil {
		return fmt.Errorf("failed to insert feedback: %w", err)
	}
	return nil
}

func (m *MongoManager) GetLatestFeedback(problemID string) (*FeedbackEntry, error) {
	filter := bson.M{"problemid": problemID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var entry FeedbackEntry
	err := m.feedbacks.FindOne(context.Background(), filter, opts).Decode(&entry)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find latest feedback: %w", err)
	}
	return &entry, nil
}

func (m *MongoManager) GetAllFeedbacksByProblemID(problemID string) ([]FeedbackEntry, error) {
	filter := bson.M{"problemid": problemID}
	cursor, err := m.feedbacks.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query feedbacks: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(context.Background()); cerr != nil {
			m.logger.Error().Msgf("failed to close cursor: %v", cerr)
		}
	}()

	var entries []FeedbackEntry
	if err := cursor.All(context.Background(), &entries); err != nil {
		return nil, fmt.Errorf("failed to decode feedbacks: %w", err)
	}
	return entries, nil
}

func (m *MongoManager) GetStatement(problemID string) (*StatementEntry, error) {
	filter := bson.M{"problemid": problemID}
	var result StatementEntry
	err := m.statements.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch statement: %w", err)
	}
	return &result, nil
}

func (m *MongoManager) SaveStatement(entry StatementEntry) error {
	_, err := m.statements.InsertOne(context.Background(), entry)
	if mongo.IsDuplicateKeyError(err) {
		m.logger.Warn().Msgf("statement already exists for problemID %s", entry.ProblemID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert statement: %w", err)
	}
	return nil
}

func (m *MongoManager) SaveSummary(entry Summary) error {
	_, err := m.summaries.InsertOne(context.Background(), entry)
	if mongo.IsDuplicateKeyError(err) {
		m.logger.Warn().Msgf("summary already exists for problemID %s", entry.ProblemID)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to insert summary: %w", err)
	}
	return nil
}

func (m *MongoManager) GetSummaryByProblemID(problemID string) (*Summary, error) {
	filter := bson.M{"problemid": problemID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var entry Summary
	err := m.summaries.FindOne(context.Background(), filter, opts).Decode(&entry)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find summary: %w", err)
	}
	return &entry, nil
}

func (m *MongoManager) GetAllStatements() ([]StatementEntry, error) {
	cursor, err := m.statements.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var statements []StatementEntry
	for cursor.Next(context.Background()) {
		var doc StatementEntry
		if err := cursor.Decode(&doc); err != nil {
			m.logger.Err(err).Msg("failed to decode statement entry")
			continue
		}
		statements = append(statements, doc)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return statements, nil
}
