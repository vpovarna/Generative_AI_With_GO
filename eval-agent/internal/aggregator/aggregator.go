package aggregator

import (
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type Weights struct {
	Stage1 float64
	Stage2 float64
}

type Aggregator struct {
	Weights Weights
}

func NewAggregator(weights Weights) *Aggregator {
	return &Aggregator{
		Weights: weights,
	}
}

func (a *Aggregator) Aggregate(id string, stage1 []models.StageResult, stage2 []models.StageResult) models.EvaluationResult {
	result := models.EvaluationResult{
		ID:     id,
		Stages: append(stage1, stage2...),
	}

	stage1Score, stage2Score := 0.0, 0.0

	for _, stage := range stage1 {
		stage1Score += stage.Score
	}

	for _, stage := range stage2 {
		stage2Score += stage.Score
	}

	if len(stage1) == 0 || len(stage2) == 0 {
		result.Verdict = "No stage checked"
		return result
	}

	stage1Avg := stage1Score / float64(len(stage1))
	stage2Avg := stage2Score / float64(len(stage2))

	confidence := (stage1Avg * a.Weights.Stage1) + (stage2Avg * a.Weights.Stage2)

	result.Confidence = confidence
	result.Verdict = a.calculateVerdict(confidence)
	return result
}

func (a *Aggregator) calculateVerdict(confidence float64) models.Verdict {
	if confidence > 0.8 {
		return models.VerdictPass
	}
	if confidence > 0.5 {
		return models.VerdictReview
	}
	return models.VerdictFail
}
