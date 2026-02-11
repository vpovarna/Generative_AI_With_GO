package bedrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

// ClaudeRequest is the message sent to Claude
type ClaudeRequest struct {
	Prompt      string
	MaxTokens   int
	Temperature float64
}

// ClaudeResponse is the Claude's response
type ClaudeResponse struct {
	Content    string
	StopReason string
}

// Claude API request format (what Bedrock expects)
type claudeMessageRequest struct {
	AnthropicVersion string          `json:"anthropic_version"`
	MaxToken         int             `json:"max_tokens"`
	Temperature      float64         `json:"temperature,omitempty"`
	Messages         []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Claude API response format (what Bedrock returns)
type claudeMessageResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

func (c *Client) InvokeModel(ctx context.Context, request ClaudeRequest) (*ClaudeResponse, error) {
	payload := claudeMessageRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxToken:         request.MaxTokens,
		Temperature:      request.Temperature,
		Messages: []claudeMessage{
			{
				Role:    "user",
				Content: request.Prompt,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal request: %w", request)
	}

	// Call Bedrock
	output, err := c.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     &c.modelID,
		Body:        body,
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to invoke mode: %w", err)
	}

	// Parse response
	var response claudeMessageResponse
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bedrock response: %w", err)
	}

	// Extract the response
	var content string
	if len(response.Content) > 0 {
		content = response.Content[0].Text
	}

	return &ClaudeResponse{
		Content:    content,
		StopReason: response.StopReason,
	}, nil
}
