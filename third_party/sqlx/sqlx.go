package sqlx

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"unicode"
)

type DB struct {
	DB *sql.DB
}

func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

func (db *DB) Close() error {
	if db == nil || db.DB == nil {
		return nil
	}
	return db.DB.Close()
}

func (db *DB) NamedQueryContext(ctx context.Context, query string, arg map[string]any) (*sql.Rows, error) {
	if db == nil || db.DB == nil {
		return nil, errors.New("db is nil")
	}
	parsedQuery, args, err := bindNamed(query, arg)
	if err != nil {
		return nil, err
	}
	return db.DB.QueryContext(ctx, parsedQuery, args...)
}

func bindNamed(query string, args map[string]any) (string, []any, error) {
	if args == nil {
		return query, nil, nil
	}
	var builder strings.Builder
	ordered := make([]any, 0)
	inName := false
	var name strings.Builder
	for i, r := range query {
		if inName {
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
				name.WriteRune(r)
				continue
			}
			value, ok := args[name.String()]
			if !ok {
				return "", nil, errors.New("missing named argument")
			}
			ordered = append(ordered, value)
			builder.WriteRune('?')
			builder.WriteRune(r)
			name.Reset()
			inName = false
			continue
		}
		if r == ':' {
			inName = true
			continue
		}
		builder.WriteRune(r)
		if i == len(query)-1 {
			break
		}
	}
	if inName {
		value, ok := args[name.String()]
		if !ok {
			return "", nil, errors.New("missing named argument")
		}
		ordered = append(ordered, value)
		builder.WriteRune('?')
	}
	return builder.String(), ordered, nil
}
