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

func TestCompletenessJudge_Evaluate_HappyPath(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 1.0, "reason": "All parts of the query are fully addressed"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "What is encryption and how does it work?",
		Answer: "Encryption is the process of encoding data to protect it. It works by using algorithms and keys to transform plaintext into ciphertext.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "completeness-judge" {
		t.Errorf("Expected name='completeness-judge', got '%s'", result.Name)
	}

	if result.Score != 1.0 {
		t.Errorf("Expected score=1.0, got %f", result.Score)
	}

	if result.Reason != "All parts of the query are fully addressed" {
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

func TestCompletenessJudge_Evaluate_PartialAnswer(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.5, "reason": "Only first part addressed: explained 'what' but not 'how' or 'why'"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "What is encryption, how does it work, and why is it important?",
		Answer: "Encryption is the process of encoding data.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.5 {
		t.Errorf("Expected score=0.5, got %f", result.Score)
	}

	if result.Reason != "Only first part addressed: explained 'what' but not 'how' or 'why'" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}

func TestCompletenessJudge_Evaluate_IncompleteAnswer(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.0, "reason": "Major parts ignored: did not address any of the three requested examples"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "Give me 3 examples of encryption algorithms",
		Answer: "Encryption is important.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0, got %f", result.Score)
	}

	if result.Reason != "Major parts ignored: did not address any of the three requested examples" {
		t.Errorf("Expected specific reason, got '%s'", result.Reason)
	}
}

func TestCompletenessJudge_Evaluate_LlmApiError(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: nil,
		ErrorToReturn:    errors.New("connection timeout"),
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "Some query",
		Answer: "Some answer",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if !mockClient.WasCalled {
		t.Error("Expected the mock LLM client to be called, but it wasn't")
	}

	if result.Name != "completeness-judge" {
		t.Errorf("Expected name='completeness-judge', got '%s'", result.Name)
	}

	if result.Score != 0.0 {
		t.Errorf("Expected score=0.0 on error, got %f", result.Score)
	}

	if result.Reason != "Failed to call LLM" {
		t.Errorf("Expected error reason, got '%s'", result.Reason)
	}
}

func TestCompletenessJudge_Evaluate_InvalidJsonFormat(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "plain text",
			content: "The answer is complete",
		},
		{
			name:    "incomplete JSON",
			content: `{"score": 1.0, "reason":`,
		},
		{
			name:    "wrong JSON type",
			content: `"score: 1.0, reason: complete"`,
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

			judge := NewCompletenessJudge(mockClient, &logger)

			result := judge.Evaluate(context.Background(), models.EvaluationContext{
				Query:  "Test query",
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

func TestCompletenessJudge_Evaluate_MissingFields(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"unrelated": "field"}`,
			StopReason: "end_turn",
		},
		ErrorToReturn: nil,
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	result := judge.Evaluate(context.Background(), models.EvaluationContext{
		Query:  "Test query",
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

func TestCompletenessJudge_Evaluate_MultiPartQuery(t *testing.T) {
	logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)

	mockClient := &MockLLMClient{
		ResponseToReturn: &bedrock.ClaudeResponse{
			Content:    `{"score": 0.67, "reason": "Addressed 2 out of 3 parts: answered benefits and types, but missed use cases"}`,
			StopReason: "end_turn",
		},
	}

	judge := NewCompletenessJudge(mockClient, &logger)

	evalContext := models.EvaluationContext{
		Query:  "What are the benefits of encryption, what types exist, and what are common use cases?",
		Answer: "Encryption provides confidentiality and integrity. Common types include symmetric and asymmetric encryption.",
	}

	result := judge.Evaluate(context.Background(), evalContext)

	if result.Score != 0.67 {
		t.Errorf("Expected score=0.67, got %f", result.Score)
	}

	if result.Name != "completeness-judge" {
		t.Errorf("Expected name='completeness-judge', got '%s'", result.Name)
	}
}
