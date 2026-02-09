package qdrant

import (
	"context"
	"fmt"

	qdrantclient "github.com/qdrant/go-client/qdrant"
)

type Client struct {
	client     *qdrantclient.Client
	collection string
}

func NewClient(baseURL, apiKey, collection string) *Client {
	return &Client{
		client:     qdrantclient.NewClient(baseURL, apiKey),
		collection: collection,
	}
}

func (c *Client) Search(ctx context.Context, vector []float32, limit int) ([]string, error) {
	if c.client == nil {
		return nil, fmt.Errorf("qdrant client not configured")
	}
	if limit <= 0 {
		limit = 5
	}
	resp, err := c.client.SearchPoints(ctx, qdrantclient.SearchPointsRequest{
		Collection:  c.collection,
		Vector:      vector,
		Limit:       uint32(limit),
		WithPayload: true,
	})
	if err != nil {
		return nil, err
	}

	passages := make([]string, 0, len(resp.Result))
	for _, item := range resp.Result {
		if text, ok := item.Payload["text"].(string); ok {
			passages = append(passages, text)
		}
	}
	return passages, nil
}
