package conversation

import "time"

type Session struct {
	ID        string //UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	ExpiresAt time.Time
}

type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

type Conversation struct {
	SessionID string
	Messages  []Message
	Metadata  map[string]any
}
