package api

import (
	"context"
	"net/http"

	"github.com/emicklei/go-restful/v3"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/bedrock"
	"github.com/povarna/generative-ai-with-go/kg-agent/internal/middleware"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	bedrockClient *bedrock.Client
	modelID       string
}

func NewHandler(client *bedrock.Client, modelID string) *Handler {
	return &Handler{
		bedrockClient: client,
		modelID:       modelID,
	}
}

// Query handles POST /api/v1/query
func (h *Handler) Query(req *restful.Request, resp *restful.Response) {
	var queryRequest QueryRequest

	if err := req.ReadEntity(&queryRequest); err != nil {
		log.Error().Err(err).Msg("Failed to parse request body")
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	queryRequest.SetDefaults()
	if err := queryRequest.Validate(); err != nil {
		middleware.HandleError(resp, err, http.StatusBadRequest)
		return
	}

	log.Info().
		Str("prompt", queryRequest.Prompt).
		Int("max_tokens", queryRequest.MaxToken).
		Float64("temperature", queryRequest.Temperature).
		Msg("Process Query")

	ctx := context.Background()
	response, err := h.bedrockClient.InvokeModel(ctx, bedrock.ClaudeRequest{
		Prompt:      queryRequest.Prompt,
		MaxTokens:   queryRequest.MaxToken,
		Temperature: queryRequest.Temperature,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to invoke Claude")
		middleware.HandleError(resp, err, http.StatusInternalServerError)
		return
	}

	queryResponse := QueryResponse{
		Content:    response.Content,
		StopReason: response.StopReason,
		Model:      h.modelID,
	}

	resp.WriteHeaderAndEntity(http.StatusOK, queryResponse)
}

// Health handler GET API /api/v1/health
func (h *Handler) Health(req *restful.Request, resp *restful.Response) {
	healthResponse := HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	}

	resp.WriteHeaderAndEntity(http.StatusOK, healthResponse)
}
