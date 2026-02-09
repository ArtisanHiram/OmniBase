package mcp

import (
	"errors"
	"strings"
)

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]any
	SQLTemplate string
}

func (t Tool) Validate() error {
	if strings.TrimSpace(t.Name) == "" {
		return errors.New("tool name required")
	}
	if t.Parameters == nil {
		return errors.New("tool parameters required")
	}
	if strings.TrimSpace(t.SQLTemplate) == "" {
		return errors.New("sql template required")
	}
	if !isReadOnlySQL(t.SQLTemplate) {
		return errors.New("sql template must be read-only")
	}
	return nil
}

func isReadOnlySQL(sql string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(sql))
	if trimmed == "" {
		return false
	}
	return strings.HasPrefix(trimmed, "select ") && !strings.Contains(trimmed, ";")
}

func DefaultTools() []Tool {
	return []Tool{
		{
			Name:        "query_student_scores",
			Description: "Fetch student scores from MySQL",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"student_id": map[string]any{"type": "integer"},
					"term":       map[string]any{"type": "string"},
				},
				"required": []string{"student_id"},
			},
			SQLTemplate: "select subject, score from student_scores where student_id = :student_id and term = :term",
		},
	}
}
