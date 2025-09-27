package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
)

type SessionRepository struct {
	collection *mongo.Collection
}

func NewSessionRepository(db *mongo.Database) domain.SessionRepository {
	return &SessionRepository{collection: db.Collection("sessions")}
}

func (r *SessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	session.MongoID = primitive.NewObjectID()
	session.ID = session.MongoID.Hex()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}
	session.LastActiveAt = time.Now()

	_, err := r.collection.InsertOne(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to create session in MongoDB: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID format: %w", err)
	}

	var session domain.Session
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("session not found in MongoDB: %w", err)
		}
		return nil, fmt.Errorf("failed to get session from MongoDB: %w", err)
	}
	session.ID = session.MongoID.Hex()
	return &session, nil
}

func (r *SessionRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	objID, err := primitive.ObjectIDFromHex(session.ID)
	if err != nil {
		return fmt.Errorf("invalid session ID format: %w", err)
	}

	session.LastActiveAt = time.Now() // Update last active time

	update := bson.M{"$set": bson.M{
		"userId":       session.UserID,
		"language":     session.Language,
		"lastActiveAt": session.LastActiveAt,
		"isGuest":      session.IsGuest,
		"title":        session.Title,
	}}
	_, err = r.collection.UpdateByID(ctx, objID, update)
	if err != nil {
		return fmt.Errorf("failed to update session in MongoDB: %w", err)
	}
	return nil
}

func (r *SessionRepository) GetSessionsByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Session, int, error) {
	var sessions []*domain.Session
	filter := bson.M{"userId": userID}
	opts := options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "lastActiveAt", Value: -1}}) // Use lastActiveAt for sorting
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find sessions by user ID: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, 0, fmt.Errorf("failed to decode sessions: %w", err)
	}

	for _, s := range sessions {
		s.ID = s.MongoID.Hex() // Ensure string ID is populated for external use
	}

	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sessions by user ID: %w", err)
	}

	return sessions, int(total), nil
}

func (r *SessionRepository) SetSessionTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	// MongoDB manages TTL via indexes if configured, not per-document explicit calls
	// This method is primarily for Redis. No-op for MongoRepo.
	return nil
}

func (r *SessionRepository) GetUserSessionIDs(ctx context.Context) ([]string, error) {
	// Not applicable for MongoDB for retrieving active session IDs. This is a Redis specific method.
	return nil, fmt.Errorf("GetUserSessionIDs not implemented for MongoDB repository")
}

func (r *SessionRepository) AddUserSessionID(ctx context.Context, sessionID string) error {
	// Not applicable for MongoDB. This is a Redis specific method.
	return nil
}

func (r *SessionRepository) RemoveUserSessionID(ctx context.Context, sessionID string) error {
	// Not applicable for MongoDB. This is a Redis specific method.
	return nil
}

// ---------------- CHAT REPOSITORY --------------------------------

type MongoChatRepository struct {
	collection *mongo.Collection
}

func NewChatRepository(db *mongo.Database) domain.ChatRepository {
	return &MongoChatRepository{collection: db.Collection("chats")}
}

func (r *MongoChatRepository) SaveChatEntry(ctx context.Context, entry *domain.ChatEntry) error {
	if entry.MongoID.IsZero() {
		entry.MongoID = primitive.NewObjectID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	_, err := r.collection.InsertOne(ctx, entry)
	if err != nil {
		return fmt.Errorf("failed to save chat entry to MongoDB: %w", err)
	}
	return nil
}

func (r *MongoChatRepository) GetChatHistory(ctx context.Context, sessionID string, limit int) ([]domain.ChatEntry, error) {
	var entries []domain.ChatEntry
	filter := bson.M{"sessionId": sessionID}
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: 1}}) // Chronological order
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history from MongoDB: %w", err)
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &entries); err != nil {
		return nil, fmt.Errorf("failed to decode chat entries: %w", err)
	}

	for i := range entries {
		entries[i].ID = entries[i].MongoID.Hex() // Ensure string ID is populated
	}
	return entries, nil
}

func (r *MongoChatRepository) BulkSaveChatEntries(ctx context.Context, entries []domain.ChatEntry) error {
	if len(entries) == 0 {
		return nil
	}

	var models []mongo.WriteModel
	for _, entry := range entries {
		// Use ReplaceOneModel with Upsert: true.
		// If _id exists, replace the document. If not, insert it.
		// This handles both initial sync and potential re-syncs.
		filter := bson.M{"_id": entry.MongoID}
		model := mongo.NewReplaceOneModel().
			SetFilter(filter).
			SetReplacement(entry).
			SetUpsert(true)
		models = append(models, model)
	}

	bulkWriteOptions := options.BulkWrite().SetOrdered(false) // Allow parallel writes
	_, err := r.collection.BulkWrite(ctx, models, bulkWriteOptions)
	if err != nil {
		return fmt.Errorf("failed to bulk save chat entries to MongoDB: %w", err)
	}
	return nil
}

func (r *MongoChatRepository) GetUnsyncedChatEntries(ctx context.Context, sessionID string) ([]domain.ChatEntry, error) {
	return nil, fmt.Errorf("GetUnsyncedChatEntries is for Redis repository, not MongoDB")
}

func (r *MongoChatRepository) MarkChatEntriesAsSynced(ctx context.Context, sessionID string, entryIDs []string) error {
	return nil
	// fmt.Errorf("MarkChatEntriesAsSynced is for Redis repository, not MongoDB")
}

