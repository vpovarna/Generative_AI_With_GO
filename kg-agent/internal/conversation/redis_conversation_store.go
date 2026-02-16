package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type RedisConversationStore struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisConversationStore(client *redis.Client, ttl time.Duration) *RedisConversationStore {
	return &RedisConversationStore{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisConversationStore) CreateSession(ctx context.Context) (*Session, error) {
	// Generate new session ID
	sessionID := uuid.New().String()

	// Create session with timestamps
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		CreatedAt: now,
		UpdatedAt: now,
		ExpiresAt: now.Add(r.ttl),
	}

	// Create empty conversation
	conversation := Conversation{
		SessionID: sessionID,
		Messages:  []Message{},
		Metadata:  map[string]any{},
	}

	// Marshal conversation to JSON
	data, err := json.Marshal(conversation)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal conversation: %w", err)
	}

	// Store in Redis with TTL
	key := r.generateKey(sessionID)
	err = r.client.Set(ctx, key, data, r.ttl).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store session in redis: %w", err)
	}

	// Add to sessions set for tracking
	err = r.client.SAdd(ctx, "sessions", sessionID).Err()
	if err != nil {
		log.Warn().Err(err).Str("sessionID", sessionID).Msg("Unable to store sessionID to sessions list in redis")
	}

	return session, nil
}

func (r *RedisConversationStore) generateKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
