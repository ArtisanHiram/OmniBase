package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SQLExecutor struct {
	db *sqlx.DB
}

func NewSQLExecutor(driver, dsn string) (*SQLExecutor, error) {
	if driver == "" || dsn == "" {
		return nil, errors.New("driver and dsn are required")
	}
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlx: %w", err)
	}
	return &SQLExecutor{db: db}, nil
}

func (e *SQLExecutor) Close() error {
	if e == nil || e.db == nil {
		return nil
	}
	return e.db.Close()
}

func (e *SQLExecutor) QueryStudentScores(ctx context.Context, studentID int, term string) (map[string]any, error) {
	if e == nil || e.db == nil {
		return nil, errors.New("sql executor not configured")
	}
	query := "select subject, score from student_scores where student_id = :student_id and term = :term"
	rows, err := e.db.NamedQueryContext(ctx, query, map[string]any{"student_id": studentID, "term": term})
	if err != nil {
		return nil, fmt.Errorf("query scores: %w", err)
	}
	defer rows.Close()

	scores := make([]map[string]any, 0)
	for rows.Next() {
		var subject string
		var score int
		if err := rows.Scan(&subject, &score); err != nil {
			return nil, fmt.Errorf("scan scores: %w", err)
		}
		scores = append(scores, map[string]any{"subject": subject, "score": score})
	}
	return map[string]any{"student_id": studentID, "scores": scores}, nil
}
