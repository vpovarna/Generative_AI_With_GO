package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type SearchClientConfig struct {
	BaseURL             string
	Timeout             time.Duration
	MaxIdleConns        int
	MaxIdleConnsPerHost int
}

type SearchClient struct {
	baseURL    string
	httpClient *http.Client
}

type SearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type SearchResult struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
}

type SearchResponse struct {
	Query  string         `json:"query"`
	Result []SearchResult `json:"result"`
	Count  int            `json:"count"`
	Method string         `json:"method"`
}

// NewSearchClient creates a new search API client with configurable timeout
func NewSearchClient(searchClientConfig SearchClientConfig) *SearchClient {
	transport := &http.Transport{
		MaxIdleConns:        searchClientConfig.MaxIdleConns,
		MaxIdleConnsPerHost: searchClientConfig.MaxIdleConnsPerHost,
		IdleConnTimeout:     30 * time.Second, // Keep fixed for idle connections
		DisableCompression:  false,
		DisableKeepAlives:   false,
	}

	httpClient := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return fmt.Errorf("http client stopped after 10 redirects")
			}
			return nil
		},
		Timeout: searchClientConfig.Timeout,
	}

	return &SearchClient{
		baseURL:    searchClientConfig.BaseURL,
		httpClient: httpClient,
	}
}

func (c *SearchClient) HybridSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	searchURL := c.baseURL + "/search/v1/hybrid"

	log.Info().
		Str("url", searchURL).
		Str("query", query).
		Int("limit", limit).
		Msg("Calling search API")

	payload := SearchRequest{
		Query: query,
		Limit: limit,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal search request")
		return nil, fmt.Errorf("failed to marshal search request: %w", err)
	}

	// Create HTTP request with context with ctx for cancellation
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, searchURL, bytes.NewBuffer(body))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Search API request failed")
		return nil, fmt.Errorf("search API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().
			Int("status_code", resp.StatusCode).
			Msg("Search API returned non-OK status")
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	var searchResp SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		log.Error().Err(err).Msg("Failed to decode search response")
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	log.Info().
		Int("results_count", searchResp.Count).
		Str("method", searchResp.Method).
		Msg("Search API call successful")

	// Return just the results array (not the whole response)
	return searchResp.Result, nil
}
