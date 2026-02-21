package stages

import (
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type Checker interface {
	Check(evaluationContext models.EvaluationContext) models.StageResult
}
