package stages

import (
	"fmt"
	"strings"
	"time"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
)

type OverlapChecker struct {
	MinOverlapThreshold float64
}

func NewOverlapChecker() *OverlapChecker {
	return &OverlapChecker{}
}

// OverlapChecker scores an answer based on keyword overlap with the query.
// It tokenizes both strings, computes the ratio of shared unique words,
// and returns a low score if the answer doesn't share enough terms with the query.
func (c *OverlapChecker) Check(evaluationContext models.EvaluationContext) models.StageResult {

	if c.MinOverlapThreshold == 0.0 {
		// set default value
		c.MinOverlapThreshold = 0.1
	}

	result := models.StageResult{
		Name:     "overlap-checker",
		Score:    0.0,
		Reason:   "",
		Duration: 0,
	}
	now := time.Now()

	if len(evaluationContext.Query) == 0 {
		result.Reason = "Empty Query"
		result.Duration = time.Since(now)
		return result
	}

	if len(evaluationContext.Answer) == 0 {
		result.Reason = "Empty Answer"
		result.Duration = time.Since(now)
		return result
	}

	queryTokens := c.stringTokenizer(evaluationContext.Query)
	answerTokens := c.stringTokenizer(evaluationContext.Answer)

	uniqueQueryTokens := extractUniqueTokens(queryTokens)
	uniqueAnswerTokens := extractUniqueTokens(answerTokens)

	count := 0
	for token := range uniqueQueryTokens {
		if _, exists := uniqueAnswerTokens[token]; exists {
			count++
		}
	}

	score := float64(count) / float64(len(uniqueQueryTokens))
	if score < c.MinOverlapThreshold {
		result.Reason = fmt.Sprintf("Low keyword overlap: %.0f%% of query terms found in answer", score*100)
		result.Score = score
	} else {
		result.Reason = "There is a good overlap"
		result.Score = score
	}

	result.Duration = time.Since(now)
	return result

}

func extractUniqueTokens(tokens []string) map[string]bool {
	unique := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		unique[t] = true
	}
	return unique
}

func (c *OverlapChecker) stringTokenizer(s string) []string {
	return strings.Split(s, " ")
}
