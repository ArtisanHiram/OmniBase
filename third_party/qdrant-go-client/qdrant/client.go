package qdrant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
	APIKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	return &Client{BaseURL: baseURL, APIKey: apiKey, client: &http.Client{}}
}

type SearchPointsRequest struct {
	Collection  string
	Vector      []float32
	Limit       uint32
	WithPayload bool
}

type SearchPointsResponse struct {
	Result []ScoredPoint `json:"result"`
}

type ScoredPoint struct {
	Payload map[string]any `json:"payload"`
}

func (c *Client) SearchPoints(ctx context.Context, req SearchPointsRequest) (SearchPointsResponse, error) {
	endpoint := fmt.Sprintf("%s/collections/%s/points/search", c.BaseURL, req.Collection)
	payload := map[string]any{
		"vector":       req.Vector,
		"limit":        req.Limit,
		"with_payload": req.WithPayload,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SearchPointsResponse{}, fmt.Errorf("encode search request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return SearchPointsResponse{}, fmt.Errorf("create search request: %w", err)
	}
	httpReq.Header.Set("content-type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("api-key", c.APIKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return SearchPointsResponse{}, fmt.Errorf("send search request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return SearchPointsResponse{}, fmt.Errorf("search failed: status %d", resp.StatusCode)
	}

	var decoded SearchPointsResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return SearchPointsResponse{}, fmt.Errorf("decode search response: %w", err)
	}
	return decoded, nil
}
