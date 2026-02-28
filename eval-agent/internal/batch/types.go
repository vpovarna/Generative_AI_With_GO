package batch

import "github.com/povarna/generative-ai-agents/eval-agent/internal/models"

type InputRecord struct {
	LineNumber int
	Request    models.EvaluationRequest
	Error      error
}
