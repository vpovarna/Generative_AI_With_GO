package executor

import (
	"context"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/aggregator"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/judge"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/stages"
)

type Executor struct {
	stageRunner        *stages.StageRunner
	judgeRunner        *judge.JudgeRunner
	aggregator         *aggregator.Aggregator
	earlyExitThreshold float64
}

func NewExecutor(
	stageRunner *stages.StageRunner,
	judgeRunner *judge.JudgeRunner,
	aggregator *aggregator.Aggregator,
	earlyExitThreshold float64,
) *Executor {
	return &Executor{
		stageRunner:        stageRunner,
		judgeRunner:        judgeRunner,
		aggregator:         aggregator,
		earlyExitThreshold: earlyExitThreshold,
	}
}

func (e *Executor) Execute(ctx context.Context, evalCtx models.EvaluationContext) models.EvaluationResult {
	id := evalCtx.RequestID
	result := models.EvaluationResult{
		ID:         id,
		Stages:     []models.StageResult{},
		Confidence: 0,
		Verdict:    "",
	}

	stageEvalResults := e.stageRunner.Run(evalCtx)

	if len(stageEvalResults) == 0 {
		result.Verdict = models.VerdictFail
		return result
	}

	stageEvalScore := 0.0
	for _, stageEval := range stageEvalResults {
		stageEvalScore += stageEval.Score
	}

	stageEvalAvgScore := stageEvalScore / float64(len(stageEvalResults))

	if stageEvalAvgScore < e.earlyExitThreshold {
		result.Stages = append(result.Stages, stageEvalResults...)
		result.Verdict = models.VerdictFail

		return result
	}

	judgeEvaResults := e.judgeRunner.Run(ctx, evalCtx)

	return e.aggregator.Aggregate(id, stageEvalResults, judgeEvaResults)
}
