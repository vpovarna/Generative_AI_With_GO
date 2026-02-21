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
	Name    string
	Type    string
	Version string
}

type Interaction struct {
	UserQuery string
	Context   string
	Answer    string
}

// Input message
type EvaluationRequest struct {
	EventID     string
	EventType   EventType
	Agent       Agent
	Interaction Interaction
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
	Name     string
	Score    float64
	Reason   string
	Duration time.Duration
}

// Final output emitted to Kafka
type EvaluationResult struct {
	ID         string
	Stages     []StageResult
	Confidence float64
	Verdict    Verdict
}
