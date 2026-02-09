package flow

import (
	"context"

	adkflow "github.com/google/adk-go/flow"

	"omnibase/internal/schema"
)

type Flow struct {
	Normalizer RequestNormalizerNode
	RAG        RAGRetrievalNode
	MCP        MCPToolDispatchNode
	LLM        LLMCompletionNode
	Formatter  ResponseFormatterNode
}

func (f Flow) Execute(ctx context.Context, input schema.UserRequest) (schema.StudentAnalysis, error) {
	pipeline := adkflow.Pipeline[schema.UserRequest, schema.NormalizedRequest, schema.RAGContext, schema.MCPContext, schema.LLMResponse, schema.StudentAnalysis]{
		First:  f.Normalizer,
		Second: f.RAG,
		Third:  f.MCP,
		Fourth: f.LLM,
		Fifth:  f.Formatter,
	}
	return pipeline.Execute(ctx, input)
}
