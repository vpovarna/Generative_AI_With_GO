package stages

import (
	"sync"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type StageRunner struct {
	Checkers []Checker
}

func NewStageRunner(checkers []Checker) *StageRunner {
	return &StageRunner{
		Checkers: checkers,
	}
}

func (r *StageRunner) Run(evaluationContext models.EvaluationContext) []models.StageResult {
	result := make(chan models.StageResult, len(r.Checkers))
	var wg sync.WaitGroup

	for _, checker := range r.Checkers {
		wg.Add(1)
		go func(c Checker) {
			defer wg.Done()
			result <- c.Check(evaluationContext)
		}(checker)
	}

	wg.Wait()
	close(result)

	var stageResults []models.StageResult
	for res := range result {
		stageResults = append(stageResults, res)
	}

	return stageResults
}
