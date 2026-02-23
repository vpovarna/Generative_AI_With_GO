package models

import "time"

type Verdict string

const (
	VerdictPass   Verdict = "pass"
	VerdictFail   Verdict = "fail"
	VerdictReview Verdict = "review"
)

type EventType string

const (
	EventTypeAgentResponse EventType = "agent_response"
	EventTypeAgentError    EventType = "agent_error"
)

type Agent struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

type Interaction struct {
	UserQuery string `json:"user_query"`
	Context   string `json:"context"`
	Answer    string `json:"answer"`
}

// Input message

type EvaluationRequest struct {
	EventID     string      `json:"event_id"`
	EventType   EventType   `json:"event_type"`
	Agent       Agent       `json:"agent"`
	Interaction Interaction `json:"interaction"`
}

// Normalized internal object
type EvaluationContext struct {
	RequestID string
	Query     string
	Context   string
	Answer    string
	CreatedAt time.Time
}

// One evaluator's output
type StageResult struct {
	Name     string        `json:"name"`
	Score    float64       `json:"score"`
	Reason   string        `json:"reason"`
	Duration time.Duration `json:"duration_ns"`
}

// Final output emitted to Kafka
type EvaluationResult struct {
	ID         string        `json:"id"`
	Stages     []StageResult `json:"stages"`
	Confidence float64       `json:"confidence"`
	Verdict    Verdict       `json:"verdict"`
}
