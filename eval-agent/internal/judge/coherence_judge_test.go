package judge

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/povarna/generative-ai-with-go/eval-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/eval-agent/internal/models"
	"github.com/rs/zerolog"
)

func TestCoherenceJudge_Evaluate_HappyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.95, "reason": "Answer is logically consistent with no contradictions"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCoherenceJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Answer: "Encryption uses mathematical algorithms to transform plaintext into ciphertext. This process ensures data confidentiality by making the data unreadable without the decryption key.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "coherence-judge" {
		t.Errorf("Expected name='coherence-judge', got '%s'", result.Name)
	}

	if result.Score != 0.95 {
		t.Errorf("Expected score=0.95, got %f", result.Score)
	}

	if result.Reason != "Answer is logically consistent with no contradictions" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}

	if result.Duration == 0 {
		t.Error("Expected duration to be measured")
	}

	if mockClient.LastRequest.MaxTokens != 256 {
		t.Errorf("Expected MaxTokens=256, got %d", mockClient.LastRequest.MaxTokens)
	}

	if mockClient.LastRequest.Temperature != 0.0 {
		t.Errorf("Expected Temperature=0.0, got %f", mockClient.LastRequest.Temperature)
	}
}

func TestCoherenceJudge_Evaluate_LowScore(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.2, "reason": "Answer contains contradictory statements"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCoherenceJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Answer: "Encryption is reversible. However, encryption cannot be reversed.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.2 {
		t.Errorf("Expected score=0.2, got %f", result.Score)
	}

	if result.Reason != "Answer contains contradictory statements" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}

func TestCoherenceJudge_Evaluate_LlmApiError(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: nil,
		ErrorToReturn:    errors.New("network timeout"),
	}

	judge := NewCoherenceJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Answer: "Some answer",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "coherence-judge" {
		t.Errorf("Expected name='coherence-judge', got '%s'", result.Name)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 on error, got %f", result.Score)
	}

	if result.Reason != "Failed to call LLM" {
		t.Errorf("Expected error reason, got '%s'", result.Reason)
	}
}

func TestCoherenceJudge_Evaluate_InvalidJsonFormat(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "plain text response",
			content: "The answer is coherent",
		},
		{
			name:    "malformed JSON - incomplete",
			content: `{"score": 0.8`,
		},
		{
			name:    "invalid JSON structure",
			content: `{score: 0.7, reason: "test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockLLMClient{
				ResponseToReturn: &bedrock.ClaudeResponse{
					Content:    tt.content,
					StopReason: "end_turn",
				},
			}

			judge := NewCoherenceJudge(mockClient, &logger)

			result := judge.Evaluate(context.Background(), models.EvaluationContext{
				Answer: "Test answer",
			})

			if result.Reason != "Failed to deserialize LLM response" {
				t.Errorf("Expected deserialization error, got '%s'", result.Reason)
			}

			if result.Score != 0.0 {
				t.Errorf("Expected score=0.0 on JSON error, got %f", result.Score)
			}
		})
	}
}

func TestCoherenceJudge_Evaluate_MissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"irrelevant": "data"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCoherenceJudge(mockClient, &logger)

	result := judge.Evaluate(context.Background(), models.EvaluationContext{
		Answer: "Test answer",
	})

	if result.Reason == "Failed to deserialize LLM response" {
		t.Error("Should not fail on valid JSON with missing fields")
	}

	if result.Reason != "Invalid LLM response: missing score and reason" {
		t.Errorf("Expected validation error, got '%s'", result.Reason)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0, got %f", result.Score)
	}
}
