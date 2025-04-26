package storage

import (
	"context"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoManager struct {
	feedbacks  *mongo.Collection
	statements *mongo.Collection
	logger     *log.Logger
}

func NewMongoManager(uri string, logger *log.Logger) (*MongoManager, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		logger.Printf("failed to connect to MongoDB: %v", err)
		return nil, fmt.Errorf("cannot connect to MongoDB: %w", err)
	}
	if err := client.Ping(context.Background(), nil); err != nil {
		logger.Printf("failed to ping MongoDB: %v", err)
		return nil, fmt.Errorf("cannot ping MongoDB: %w", err)
	}

	db := client.Database("coach_demon")
	logger.Println("connected to MongoDB")

	return &MongoManager{
		feedbacks:  db.Collection("feedbacks"),
		statements: db.Collection("statements"),
		logger:     logger,
	}, nil
}

// SaveFeedback saves a new feedback document
func (m *MongoManager) SaveFeedback(entry FeedbackEntry) error {
	_, err := m.feedbacks.InsertOne(context.Background(), entry)
	if err != nil {
		m.logger.Printf("failed to insert feedback: %v", err)
		return fmt.Errorf("failed to insert feedback: %w", err)
	}
	return nil
}

// GetLatestFeedback returns the latest feedback for a problem
func (m *MongoManager) GetLatestFeedback(problemID string) (*FeedbackEntry, error) {
	filter := bson.M{"problemid": problemID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var entry FeedbackEntry
	err := m.feedbacks.FindOne(context.Background(), filter, opts).Decode(&entry)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		m.logger.Printf("failed to find latest feedback: %v", err)
		return nil, fmt.Errorf("failed to find latest feedback: %w", err)
	}
	return &entry, nil
}

// SummarizeFeedback returns all feedback entries for a problem
func (m *MongoManager) SummarizeFeedback(problemID string) ([]FeedbackEntry, error) {
	filter := bson.M{"problemid": problemID}
	cursor, err := m.feedbacks.Find(context.Background(), filter)
	if err != nil {
		m.logger.Printf("failed to query feedbacks: %v", err)
		return nil, fmt.Errorf("failed to query feedbacks: %w", err)
	}
	defer func() {
		if cerr := cursor.Close(context.Background()); cerr != nil {
			m.logger.Printf("failed to close cursor: %v", cerr)
		}
	}()

	var entries []FeedbackEntry
	if err := cursor.All(context.Background(), &entries); err != nil {
		m.logger.Printf("failed to decode feedbacks: %v", err)
		return nil, fmt.Errorf("failed to decode feedbacks: %w", err)
	}
	return entries, nil
}

// GetStatement fetches a problem statement
func (m *MongoManager) GetStatement(problemID string) (string, error) {
	filter := bson.M{"problemid": problemID}
	var result struct {
		Statement string `bson:"statement"`
	}
	err := m.statements.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		m.logger.Printf("failed to fetch statement: %v", err)
		return "", fmt.Errorf("failed to fetch statement: %w", err)
	}
	return result.Statement, nil
}

// SaveStatement saves a new problem statement
func (m *MongoManager) SaveStatement(problemID, statement string) error {
	filter := bson.M{"problemid": problemID}
	update := bson.M{
		"$setOnInsert": bson.M{
			"problemid": problemID,
			"statement": statement,
		},
	}
	opts := options.Update().SetUpsert(true)

	_, err := m.statements.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		m.logger.Printf("failed to save statement: %v", err)
		return fmt.Errorf("failed to save statement: %w", err)
	}
	return nil
}
