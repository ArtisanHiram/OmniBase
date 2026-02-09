package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL string
	model   string
	client  *http.Client
}

func NewClient(baseURL, model string) *Client {
	return &Client{baseURL: baseURL, model: model, client: &http.Client{}}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ChatCompletionRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

func (c *Client) ChatCompletion(ctx context.Context, messages []Message, tools []Tool) (string, error) {
	endpoint := fmt.Sprintf("%s/v1/chat/completions", c.baseURL)
	payload := ChatCompletionRequest{Model: c.model, Messages: messages, Tools: tools}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode chat completion request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat completion request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send chat completion request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat completion failed: status %d", resp.StatusCode)
	}

	var decoded ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode chat completion response: %w", err)
	}
	if len(decoded.Choices) == 0 {
		return "", fmt.Errorf("chat completion response missing choices")
	}
	return decoded.Choices[0].Message.Content, nil
}

func (c *Client) Embed(ctx context.Context, input string) ([]float32, error) {
	endpoint := fmt.Sprintf("%s/v1/embeddings", c.baseURL)
	payload := EmbeddingRequest{Model: c.model, Input: input}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embedding request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send embedding request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("embedding failed: status %d", resp.StatusCode)
	}

	var decoded EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}
	if len(decoded.Data) == 0 {
		return nil, fmt.Errorf("embedding response missing data")
	}
	return decoded.Data[0].Embedding, nil
}
