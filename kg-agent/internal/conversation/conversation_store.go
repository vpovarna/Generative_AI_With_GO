package conversation

import "context"

type ConversationStore interface {
	CreateSession(ctx context.Context) (*Session, error)
	GetConversation(ctx context.Context, sessionID string) (*Conversation, error)
	AddMessage(ctx context.Context, sessionID string, message Message) error
	CleanExpiredSessions(ctx context.Context) error
}
