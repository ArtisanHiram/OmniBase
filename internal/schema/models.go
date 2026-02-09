package schema

import (
	"errors"
	"fmt"
	"strings"
)

type UserRequest struct {
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
	Mode      string `json:"mode"`
	Message   string `json:"message"`
	StudentID int    `json:"student_id"`
	Term      string `json:"term"`
}

func (req UserRequest) Validate() error {
	if strings.TrimSpace(req.RequestID) == "" {
		return errors.New("request_id is required")
	}
	if strings.TrimSpace(req.TraceID) == "" {
		return errors.New("trace_id is required")
	}
	if strings.TrimSpace(req.Mode) == "" {
		return errors.New("mode is required")
	}
	if strings.TrimSpace(req.Message) == "" {
		return errors.New("message is required")
	}
	return nil
}

type NormalizedRequest struct {
	RequestID string
	TraceID   string
	Mode      string
	Message   string
	StudentID int
	Term      string
}

func (req NormalizedRequest) Validate() error {
	if strings.TrimSpace(req.RequestID) == "" || strings.TrimSpace(req.TraceID) == "" {
		return errors.New("request_id and trace_id are required")
	}
	if strings.TrimSpace(req.Mode) == "" {
		return errors.New("mode is required")
	}
	if strings.TrimSpace(req.Message) == "" {
		return errors.New("message is required")
	}
	return nil
}

type RAGResult struct {
	Query     string
	Passages  []string
	Embedding []float32
}

func (res RAGResult) Validate() error {
	if strings.TrimSpace(res.Query) == "" {
		return errors.New("query is required")
	}
	return nil
}

type RAGContext struct {
	Request   NormalizedRequest
	Passages  []string
	Embedding []float32
}

func (ctx RAGContext) Validate() error {
	if err := ctx.Request.Validate(); err != nil {
		return err
	}
	return nil
}

type MCPResult struct {
	ToolName string
	Payload  map[string]any
}

func (res MCPResult) Validate() error {
	if strings.TrimSpace(res.ToolName) == "" {
		return errors.New("tool_name is required")
	}
	if res.Payload == nil {
		return errors.New("payload is required")
	}
	return nil
}

type MCPContext struct {
	Request  NormalizedRequest
	Passages []string
	Tool     MCPResult
}

func (ctx MCPContext) Validate() error {
	if err := ctx.Request.Validate(); err != nil {
		return err
	}
	if err := ctx.Tool.Validate(); err != nil {
		return err
	}
	return nil
}

type LLMRequest struct {
	RequestID string
	TraceID   string
	System    string
	User      string
	Tools     []ToolDefinition
	Context   map[string]any
}

func (req LLMRequest) Validate() error {
	if strings.TrimSpace(req.RequestID) == "" || strings.TrimSpace(req.TraceID) == "" {
		return errors.New("request_id and trace_id are required")
	}
	if strings.TrimSpace(req.System) == "" {
		return errors.New("system prompt is required")
	}
	if strings.TrimSpace(req.User) == "" {
		return errors.New("user prompt is required")
	}
	return nil
}

type LLMResponse struct {
	Content string
}

func (res LLMResponse) Validate() error {
	if strings.TrimSpace(res.Content) == "" {
		return errors.New("content is required")
	}
	return nil
}

type ToolDefinition struct {
	Name        string
	Description string
	Parameters  map[string]any
}

func (tool ToolDefinition) Validate() error {
	if strings.TrimSpace(tool.Name) == "" {
		return errors.New("tool name is required")
	}
	if tool.Parameters == nil {
		return errors.New("tool parameters are required")
	}
	return nil
}

type StudentAnalysis struct {
	Summary         string                `json:"summary"`
	Analysis        StudentAnalysisDetail `json:"analysis"`
	Recommendations []Recommendation      `json:"recommendations"`
	DataSnapshot    DataSnapshot          `json:"data_snapshot"`
}

type StudentAnalysisDetail struct {
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
	Trend      string   `json:"trend"`
}

type Recommendation struct {
	Action  string `json:"action"`
	Example string `json:"example"`
}

type DataSnapshot struct {
	Math    int `json:"math"`
	English int `json:"english"`
	Physics int `json:"physics"`
}

func (resp StudentAnalysis) Validate() error {
	if strings.TrimSpace(resp.Summary) == "" {
		return errors.New("summary is required")
	}
	if len(resp.Analysis.Strengths) == 0 {
		return errors.New("analysis.strengths is required")
	}
	if len(resp.Analysis.Weaknesses) == 0 {
		return errors.New("analysis.weaknesses is required")
	}
	if strings.TrimSpace(resp.Analysis.Trend) == "" {
		return errors.New("analysis.trend is required")
	}
	if len(resp.Recommendations) == 0 {
		return errors.New("recommendations is required")
	}
	for i, rec := range resp.Recommendations {
		if strings.TrimSpace(rec.Action) == "" || strings.TrimSpace(rec.Example) == "" {
			return fmt.Errorf("recommendations[%d] must include action and example", i)
		}
	}
	if resp.DataSnapshot.Math == 0 || resp.DataSnapshot.English == 0 || resp.DataSnapshot.Physics == 0 {
		return errors.New("data_snapshot requires math, english, physics")
	}
	return nil
}
