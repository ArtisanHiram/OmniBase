package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	adkflow "github.com/google/adk-go/flow"

	"omnibase/internal/llm"
	"omnibase/internal/logging"
	"omnibase/internal/mcp"
	"omnibase/internal/qdrant"
	"omnibase/internal/schema"
)

type RequestNormalizerNode struct{}

func (n RequestNormalizerNode) Name() string { return "request_normalizer" }

func (n RequestNormalizerNode) Run(ctx context.Context, input schema.UserRequest) (schema.NormalizedRequest, error) {
	logger := logging.FromContext(ctx, nil)
	if err := input.Validate(); err != nil {
		return schema.NormalizedRequest{}, err
	}
	mode := strings.ToLower(strings.TrimSpace(input.Mode))
	switch mode {
	case "customer_support", "student_analysis":
	default:
		return schema.NormalizedRequest{}, fmt.Errorf("unsupported mode: %s", input.Mode)
	}
	output := schema.NormalizedRequest{
		RequestID: strings.TrimSpace(input.RequestID),
		TraceID:   strings.TrimSpace(input.TraceID),
		Mode:      mode,
		Message:   strings.TrimSpace(input.Message),
		StudentID: input.StudentID,
		Term:      strings.TrimSpace(input.Term),
	}
	if err := output.Validate(); err != nil {
		return schema.NormalizedRequest{}, err
	}
	if logger != nil {
		logger.Info("normalized request", "mode", output.Mode)
	}
	return output, nil
}

var _ adkflow.Node[schema.UserRequest, schema.NormalizedRequest] = (*RequestNormalizerNode)(nil)

type RAGRetrievalNode struct {
	LLM    *llm.Client
	Qdrant *qdrant.Client
	TopK   int
}

func (n RAGRetrievalNode) Name() string { return "rag_retrieval" }

func (n RAGRetrievalNode) Run(ctx context.Context, input schema.NormalizedRequest) (schema.RAGContext, error) {
	logger := logging.FromContext(ctx, nil)
	if err := input.Validate(); err != nil {
		return schema.RAGContext{}, err
	}
	if n.TopK <= 0 {
		n.TopK = 5
	}
	vector, err := n.LLM.Embed(ctx, input.Message)
	if err != nil {
		return schema.RAGContext{}, err
	}
	passages, err := n.Qdrant.Search(ctx, vector, n.TopK)
	if err != nil {
		return schema.RAGContext{}, err
	}
	if logger != nil {
		logger.Info("rag retrieved", "passage_count", len(passages))
	}
	return schema.RAGContext{Request: input, Passages: passages, Embedding: vector}, nil
}

var _ adkflow.Node[schema.NormalizedRequest, schema.RAGContext] = (*RAGRetrievalNode)(nil)

type MCPToolDispatchNode struct {
	Client *mcp.Client
}

func (n MCPToolDispatchNode) Name() string { return "mcp_tool_dispatch" }

func (n MCPToolDispatchNode) Run(ctx context.Context, input schema.RAGContext) (schema.MCPContext, error) {
	logger := logging.FromContext(ctx, nil)
	if err := input.Validate(); err != nil {
		return schema.MCPContext{}, err
	}
	toolName := "query_student_scores"
	args := map[string]any{
		"student_id": input.Request.StudentID,
		"term":       input.Request.Term,
	}
	if input.Request.Mode != "student_analysis" {
		return schema.MCPContext{Request: input.Request, Passages: input.Passages, Tool: schema.MCPResult{ToolName: "noop", Payload: map[string]any{}}}, nil
	}
	if input.Request.StudentID == 0 {
		return schema.MCPContext{}, errors.New("student_id is required for student_analysis")
	}
	payload, err := n.Client.Dispatch(ctx, toolName, args)
	if err != nil {
		return schema.MCPContext{}, err
	}
	if logger != nil {
		logger.Info("mcp tool dispatched", "tool", toolName)
	}
	return schema.MCPContext{Request: input.Request, Passages: input.Passages, Tool: schema.MCPResult{ToolName: toolName, Payload: payload}}, nil
}

var _ adkflow.Node[schema.RAGContext, schema.MCPContext] = (*MCPToolDispatchNode)(nil)

type LLMCompletionNode struct {
	Client *llm.Client
	Tools  []mcp.Tool
}

func (n LLMCompletionNode) Name() string { return "llm_completion" }

func (n LLMCompletionNode) Run(ctx context.Context, input schema.MCPContext) (schema.LLMResponse, error) {
	logger := logging.FromContext(ctx, nil)
	if err := input.Validate(); err != nil {
		return schema.LLMResponse{}, err
	}
	tools := make([]llm.Tool, 0, len(n.Tools))
	for _, tool := range n.Tools {
		tools = append(tools, llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		})
	}
	payload, _ := json.Marshal(input.Tool.Payload)
	systemPrompt := "You are OmniBase AI. Always respond with valid JSON matching the required schema."
	contextPrompt := ""
	if len(input.Passages) > 0 {
		contextPrompt = "\nRetrieved passages:\n" + strings.Join(input.Passages, "\n")
	}
	userPrompt := fmt.Sprintf("%s\nTool data: %s", input.Request.Message, string(payload))
	messages := []llm.Message{{Role: "system", Content: systemPrompt + contextPrompt}, {Role: "user", Content: userPrompt}}

	content, err := n.Client.ChatCompletion(ctx, messages, tools)
	if err != nil {
		return schema.LLMResponse{}, err
	}
	if logger != nil {
		logger.Info("llm completion received")
	}
	return schema.LLMResponse{Content: content}, nil
}

var _ adkflow.Node[schema.MCPContext, schema.LLMResponse] = (*LLMCompletionNode)(nil)

type ResponseFormatterNode struct{}

func (n ResponseFormatterNode) Name() string { return "response_formatter" }

func (n ResponseFormatterNode) Run(ctx context.Context, input schema.LLMResponse) (schema.StudentAnalysis, error) {
	logger := logging.FromContext(ctx, nil)
	if err := input.Validate(); err != nil {
		return schema.StudentAnalysis{}, err
	}
	var response schema.StudentAnalysis
	if err := json.Unmarshal([]byte(input.Content), &response); err != nil {
		return schema.StudentAnalysis{}, fmt.Errorf("invalid LLM JSON: %w", err)
	}
	if err := response.Validate(); err != nil {
		return schema.StudentAnalysis{}, err
	}
	if logger != nil {
		logger.Info("response formatted")
	}
	return response, nil
}

var _ adkflow.Node[schema.LLMResponse, schema.StudentAnalysis] = (*ResponseFormatterNode)(nil)
