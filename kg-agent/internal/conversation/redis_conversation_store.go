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

func (r *RedisConversationStore) GetConversation(ctx context.Context, sessionID string) (*Conversation, error) {
	key := r.generateKey(sessionID)

	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("session id: %s not found", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var conversation Conversation
	if err := json.Unmarshal([]byte(data), &conversation); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}
	return &conversation, nil
}

func (r *RedisConversationStore) AddMessage(ctx context.Context, sessionID string, message Message) error {
	conversation, err := r.GetConversation(ctx, sessionID)
	if err != nil {
		return err
	}

	conversation.Messages = append(conversation.Messages, message)
	data, err := json.Marshal(conversation)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation. Error: %w", err)
	}

	key := r.generateKey(sessionID)
	err = r.client.Set(ctx, key, data, r.ttl).Err()

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

func (r *RedisConversationStore) CleanExpiredSessions(ctx context.Context) error {
	sessions, err := r.client.SMembers(ctx, "sessions").Result()
	if err != nil {
		return fmt.Errorf("failed to get session IDs: %w", err)
	}

	for _, session := range sessions {
		key := r.generateKey(session)

		exists, err := r.client.Exists(ctx, key).Result()
		if err != nil {
			log.Warn().Err(err).Str("sessionID", session).Msg("failed to check if session exists")
			continue
		}

		// If session is expired, exists will be 0
		if exists == 0 {
			err := r.client.SRem(ctx, "sessions", session).Err()
			if err != nil {
				log.Warn().Err(err).Str("sessionID", session).Msg("failed to remove expired session from set")
			}
		}
	}

	return nil
}

func (r *RedisConversationStore) generateKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}
