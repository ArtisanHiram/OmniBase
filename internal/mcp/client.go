package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	baseURL  string
	tools    map[string]Tool
	client   *http.Client
	executor *SQLExecutor
}

func NewClient(baseURL string, tools []Tool, executor *SQLExecutor) (*Client, error) {
	toolMap := make(map[string]Tool, len(tools))
	for _, tool := range tools {
		if err := tool.Validate(); err != nil {
			return nil, fmt.Errorf("invalid tool %s: %w", tool.Name, err)
		}
		toolMap[tool.Name] = tool
	}
	return &Client{baseURL: baseURL, tools: toolMap, client: &http.Client{}, executor: executor}, nil
}

type ToolRequest struct {
	ToolName  string         `json:"tool_name"`
	Arguments map[string]any `json:"arguments"`
	SQL       string         `json:"sql"`
}

type ToolResponse struct {
	Data map[string]any `json:"data"`
}

func (c *Client) Dispatch(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	tool, ok := c.tools[toolName]
	if !ok {
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
	if c.baseURL == "" && c.executor != nil {
		return c.dispatchLocal(ctx, toolName, args)
	}
	payload := ToolRequest{ToolName: toolName, Arguments: args, SQL: tool.SQLTemplate}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode tool request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/v1/tools/%s", c.baseURL, toolName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create tool request: %w", err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dispatch tool request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("tool dispatch failed: status %d", resp.StatusCode)
	}

	var decoded ToolResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode tool response: %w", err)
	}
	if decoded.Data == nil {
		return nil, fmt.Errorf("tool response missing data")
	}
	return decoded.Data, nil
}

func (c *Client) dispatchLocal(ctx context.Context, toolName string, args map[string]any) (map[string]any, error) {
	switch toolName {
	case "query_student_scores":
		studentID, _ := args["student_id"].(int)
		term, _ := args["term"].(string)
		if studentID == 0 {
			return nil, fmt.Errorf("student_id is required")
		}
		return c.executor.QueryStudentScores(ctx, studentID, term)
	default:
		return nil, fmt.Errorf("unsupported local tool: %s", toolName)
	}
}

func (c *Client) Tools() []Tool {
	tools := make([]Tool, 0, len(c.tools))
	for _, tool := range c.tools {
		tools = append(tools, tool)
	}
	return tools
}
