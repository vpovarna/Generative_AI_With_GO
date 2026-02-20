package embedding

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/rs/zerolog/log"
)

var titanEmbeddingModelID string = "amazon.titan-embed-text-v2:0"

type BedrockEmbedder struct {
	client  *bedrockruntime.Client
	modelID string
}

type BedrockEmbedderRequest struct {
	InputText string `json:"inputText"`
}

type BedrockEmbedderResponse struct {
	Embedding           []float32 `json:"embedding"`
	InputTextTokenCount int       `json:"inputTextTokenCount"`
}

func NewBedrockEmbedder(bedrockClient *bedrockruntime.Client) *BedrockEmbedder {
	return &BedrockEmbedder{
		client:  bedrockClient,
		modelID: titanEmbeddingModelID,
	}
}

func (e *BedrockEmbedder) GenerateEmbeddings(ctx context.Context, query string) ([]float32, error) {

	payload := BedrockEmbedderRequest{
		InputText: query,
	}

	bytes, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Unable to serialize user query")
		return nil, err
	}

	response, err := e.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		Body:        bytes,
		ModelId:     aws.String(e.modelID),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
	})

	if err != nil {
		log.Error().Err(err).Msg("Unable to query embedding model")
		return nil, err
	}

	var bedrockEmbedderResponse BedrockEmbedderResponse

	err = json.Unmarshal(response.Body, &bedrockEmbedderResponse)
	if err != nil {
		log.Error().Err(err).Msg("Failed embedder response")
		return nil, err
	}

	return bedrockEmbedderResponse.Embedding, nil
}

func (e *BedrockEmbedder) GenerateBatchEmbeddings(ctx context.Context, queries []string) ([][]float32, error) {
	requestBody := map[string]any{
		"inputText": queries,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize input queries: %w", err)
	}

	output, err := e.client.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId: aws.String(e.modelID),
		Accept:  aws.String("application/json"),
		Body:    bodyBytes,
	})

	var response struct {
		Embeddings [][]float32 `json:"embeddings"`
	}
	if err := json.Unmarshal(output.Body, &response); err != nil {
		return nil, fmt.Errorf("Unable to read bedrock response, Error: %w", err)
	}

	return response.Embeddings, nil
}
