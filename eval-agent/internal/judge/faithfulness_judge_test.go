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

func TestFaithfulnessJudge_Evaluate_HappyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.98, "reason": "Answer is fully grounded in the provided context with no hallucinations"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewFaithfulnessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Context: "AWS KMS is a managed service that makes it easy to create and control encryption keys. It uses hardware security modules to protect keys.",
		Answer:  "AWS KMS is a managed service for creating and controlling encryption keys, using hardware security modules for protection.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "faithfulness-judge" {
		t.Errorf("Expected name='faithfulness-judge', got '%s'", result.Name)
	}

	if result.Score != 0.98 {
		t.Errorf("Expected score=0.98, got %f", result.Score)
	}

	if result.Reason != "Answer is fully grounded in the provided context with no hallucinations" {
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

func TestFaithfulnessJudge_Evaluate_Hallucination(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.3, "reason": "Answer introduces facts not present in context: mentions pricing details not in source"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewFaithfulnessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Context: "Redis is an in-memory data store.",
		Answer:  "Redis is an in-memory data store that costs $0.10 per hour.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.3 {
		t.Errorf("Expected score=0.3, got %f", result.Score)
	}

	if result.Reason != "Answer introduces facts not present in context: mentions pricing details not in source" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}

func TestFaithfulnessJudge_Evaluate_LlmApiError(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: nil,
		ErrorToReturn:    errors.New("bedrock throttling error"),
	}

	judge := NewFaithfulnessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Context: "Some context",
		Answer:  "Some answer",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "faithfulness-judge" {
		t.Errorf("Expected name='faithfulness-judge', got '%s'", result.Name)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 on error, got %f", result.Score)
	}

	if result.Reason != "Failed to call LLM" {
		t.Errorf("Expected error reason, got '%s'", result.Reason)
	}
}

func TestFaithfulnessJudge_Evaluate_InvalidJsonFormat(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "not JSON",
			content: "The answer is faithful to the context",
		},
		{
			name:    "broken JSON",
			content: `{"score": 0.9, "reason": "faithful"`,
		},
		{
			name:    "JSON array instead of object",
			content: `[0.9, "faithful"]`,
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

			judge := NewFaithfulnessJudge(mockClient, &logger)

			result := judge.Evaluate(context.Background(), models.EvaluationContext{
				Context: "Test context",
				Answer:  "Test answer",
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

func TestFaithfulnessJudge_Evaluate_MissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"other_data": "value"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewFaithfulnessJudge(mockClient, &logger)

	result := judge.Evaluate(context.Background(), models.EvaluationContext{
		Context: "Test context",
		Answer:  "Test answer",
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

func TestFaithfulnessJudge_Evaluate_EmptyContext(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.0, "reason": "No context provided to evaluate faithfulness"}`,
			StopReason: "end_turn",
		},
	}

	judge := NewFaithfulnessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Context: "",
		Answer:  "Some answer about encryption",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	// The judge should still run even with empty context
	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 for empty context, got %f", result.Score)
	}

	if result.Name != "faithfulness-judge" {
		t.Errorf("Expected name='faithfulness-judge', got '%s'", result.Name)
	}
}
