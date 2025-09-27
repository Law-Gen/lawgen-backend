package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
)

type RedisSessionRepository struct {
	client *redis.Client
	cfg    *config.Config
}

func NewRedisSessionRepository(client *redis.Client, cfg *config.Config) domain.SessionRepository {
	return &RedisSessionRepository{client: client, cfg: cfg}
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// Key for the Redis Set that stores IDs of active user sessions (for sync job)
const activeUserSessionsKey = "active_user_session_ids"

func (r *RedisSessionRepository) GetSessionByID(ctx context.Context, id string) (*domain.Session, error) {
	data, err := r.client.Get(ctx, sessionKey(id)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("session not found in Redis: %w", err)
		}
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	session := &domain.Session{}
	if err := json.Unmarshal(data, session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session from Redis: %w", err)
	}
	return session, nil
}

func (r *RedisSessionRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	if session.MongoID.IsZero() {
		session.MongoID = primitive.NewObjectID()
		session.ID = session.MongoID.Hex() // Ensure ID is set from MongoID
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session for Redis: %w", err)
	}

	ttl := time.Duration(r.cfg.SessionTTLSeconds) * time.Second
	// For account holders, add to the set of active user sessions
	if !session.IsGuest && session.UserID != "" {
		if _, err := r.client.SAdd(ctx, activeUserSessionsKey, session.ID).Result(); err != nil {
			return fmt.Errorf("failed to add user session ID to active set: %w", err)
		}
	}

	_, err = r.client.Set(ctx, sessionKey(session.ID), data, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to create session in Redis: %w", err)
	}
	return nil
}

func (r *RedisSessionRepository) UpdateSession(ctx context.Context, session *domain.Session) error {
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session for Redis update: %w", err)
	}

	ttl := time.Duration(r.cfg.SessionTTLSeconds) * time.Second
	// Refresh TTL on update to keep active sessions alive
	_, err = r.client.Set(ctx, sessionKey(session.ID), data, ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to update session in Redis: %w", err)
	}

	// For account holders, ensure it's in the active set and its TTL is maintained (implicitly by updates)
	if !session.IsGuest && session.UserID != "" {
		if _, err := r.client.SAdd(ctx, activeUserSessionsKey, session.ID).Result(); err != nil {
			return fmt.Errorf("failed to ensure user session ID in active set: %w", err)
		}
		// Also, refresh TTL for the activeUserSessionsKey entry if it were individual,
		// but since it's a Set, individual TTLs are not possible. The session object's TTL is refreshed.
	}
	return nil
}

func (r *RedisSessionRepository) GetSessionsByUserID(ctx context.Context, userID string, page, limit int) ([]*domain.Session, int, error) {
	// Not efficient for Redis. This method is handled by MongoDB repository.
	return nil, 0, fmt.Errorf("GetSessionsByUserID not implemented for Redis repository (use MongoDB for this)")
}

func (r *RedisSessionRepository) SetSessionTTL(ctx context.Context, sessionID string, ttl time.Duration) error {
	_, err := r.client.Expire(ctx, sessionKey(sessionID), ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to set TTL for session %s in Redis: %w", sessionID, err)
	}
	return nil
}

func (r *RedisSessionRepository) GetUserSessionIDs(ctx context.Context) ([]string, error) {
	// Retrieve all active user session IDs from the Redis Set
	ids, err := r.client.SMembers(ctx, activeUserSessionsKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active user session IDs from Redis: %w", err)
	}
	return ids, nil
}

func (r *RedisSessionRepository) AddUserSessionID(ctx context.Context, sessionID string) error {
	_, err := r.client.SAdd(ctx, activeUserSessionsKey, sessionID).Result()
	return err
}

func (r *RedisSessionRepository) RemoveUserSessionID(ctx context.Context, sessionID string) error {
	_, err := r.client.SRem(ctx, activeUserSessionsKey, sessionID).Result()
	return err
}

// ------------------------- CHAT REPOSITORY --------------------------------

type RedisChatRepository struct {
	client *redis.Client
	cfg    *config.Config
}

func NewRedisChatRepository(client *redis.Client, cfg *config.Config) domain.ChatRepository {
	return &RedisChatRepository{client: client, cfg: cfg}
}

func chatHistoryKey(sessionID string) string {
	return fmt.Sprintf("chat_history:%s", sessionID)
}

func (r *RedisChatRepository) GetChatHistory(ctx context.Context, sessionID string, limit int) ([]domain.ChatEntry, error) {
	var cmds []string
	var err error

	// `limit` here refers to the number of *pairs* (user query + LLM response).
	// So we fetch `limit * 2` elements. If limit is 0, fetch all.
	if limit > 0 {
		cmds, err = r.client.LRange(ctx, chatHistoryKey(sessionID), -int64(limit*2), -1).Result()
	} else {
		cmds, err = r.client.LRange(ctx, chatHistoryKey(sessionID), 0, -1).Result()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get chat history from Redis: %w", err)
	}

	var history []domain.ChatEntry
	for _, cmd := range cmds {
		entry := domain.ChatEntry{}
		if err := json.Unmarshal([]byte(cmd), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat entry from Redis: %w", err)
		}
		// Ensure MongoID and ID are populated from unmarshalled values
		if !entry.MongoID.IsZero() {
			entry.ID = entry.MongoID.Hex()
		}
		history = append(history, entry)
	}
	return history, nil
}

func (r *RedisChatRepository) SaveChatEntry(ctx context.Context, entry *domain.ChatEntry) error {
	// For new entries, generate a MongoDB ObjectID now. This ID will be used when saving to MongoDB.
	if entry.MongoID.IsZero() {
		entry.MongoID = primitive.NewObjectID()
		entry.ID = entry.MongoID.Hex() // Ensure public ID is also set
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}

	entry.SyncedToDB = false // Mark as unsynced initially

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal chat entry for Redis: %w", err)
	}

	// Append to list
	_, err = r.client.RPush(ctx, chatHistoryKey(entry.SessionID), data).Result()
	if err != nil {
		// If context was canceled, return a more specific error
		if err == context.Canceled {
			return fmt.Errorf("failed to save chat entry to Redis: context canceled (client disconnected or request timed out)")
		}
		return fmt.Errorf("failed to save chat entry to Redis: %w", err)
	}

	// Set/refresh TTL for the chat history list, linked to the session's TTL
	_, err = r.client.Expire(ctx, chatHistoryKey(entry.SessionID), time.Duration(r.cfg.SessionTTLSeconds)*time.Second).Result()
	if err != nil {
		return fmt.Errorf("failed to set TTL for chat history in Redis: %w", err)
	}

	return nil
}

func (r *RedisChatRepository) BulkSaveChatEntries(ctx context.Context, entries []domain.ChatEntry) error {
	return fmt.Errorf("BulkSaveChatEntries not directly applicable for Redis write, use SaveChatEntry")
}

func (r *RedisChatRepository) GetUnsyncedChatEntries(ctx context.Context, sessionID string) ([]domain.ChatEntry, error) {
	cmds, err := r.client.LRange(ctx, chatHistoryKey(sessionID), 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get chat history for sync from Redis: %w", err)
	}

	var unsyncedEntries []domain.ChatEntry
	for _, cmd := range cmds {
		entry := domain.ChatEntry{}
		if err := json.Unmarshal([]byte(cmd), &entry); err != nil {
			return nil, fmt.Errorf("failed to unmarshal chat entry for sync: %w", err)
		}
		if !entry.SyncedToDB {
			unsyncedEntries = append(unsyncedEntries, entry)
		}
	}
	return unsyncedEntries, nil
}

func (r *RedisChatRepository) MarkChatEntriesAsSynced(ctx context.Context, sessionID string, entryMongoIDs []string) error {
	// This is an inefficient operation for a Redis List as it requires fetching, modifying, and re-writing.
	// For production, consider a different Redis data structure (e.g., Hashes for entries + a List of IDs).

	pipe := r.client.Pipeline()
	defer pipe.Close()

	// 1. Get all elements
	cmds, err := pipe.LRange(ctx, chatHistoryKey(sessionID), 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get chat history for marking sync: %w", err)
	}

	updatedList := make([]interface{}, 0, len(cmds))
	for _, cmd := range cmds {
		entry := domain.ChatEntry{}
		if err := json.Unmarshal([]byte(cmd), &entry); err != nil {
			// Log error, but try to continue
			log.Printf("Warning: Failed to unmarshal chat entry during sync marking: %v", err)
			updatedList = append(updatedList, cmd) // Keep original if unmarshal fails
			continue
		}

		// Check if this entry's MongoID is in the list of IDs to mark as synced
		for _, mongoIDToSync := range entryMongoIDs {
			if entry.MongoID.Hex() == mongoIDToSync {
				entry.SyncedToDB = true
				break
			}
		}
		updatedData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("Warning: Failed to marshal chat entry during sync marking: %v", err)
			updatedList = append(updatedList, cmd) // Keep original if marshal fails
			continue
		}
		updatedList = append(updatedList, updatedData)
	}

	// 2. Delete the old list
	pipe.Del(ctx, chatHistoryKey(sessionID))
	// 3. Push the updated elements
	if len(updatedList) > 0 {
		pipe.RPush(ctx, chatHistoryKey(sessionID), updatedList...)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark chat entries as synced in Redis: %w", err)
	}
	return nil
}
