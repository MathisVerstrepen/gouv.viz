package store

import (
	"context"
	"database/sql"
	"strconv"
)

type filterOption struct {
	Value string
	Label string
}

func (s *Store) intFilterOptions(ctx context.Context, query string) ([]filterOption, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []filterOption
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		text := strconv.Itoa(value)
		options = append(options, filterOption{Value: text, Label: text})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return options, nil
}

func (s *Store) textFilterOptions(ctx context.Context, query string) ([]filterOption, error) {
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var options []filterOption
	for rows.Next() {
		var value string
		var label sql.NullString
		if err := rows.Scan(&value, &label); err != nil {
			return nil, err
		}
		if value == "" {
			continue
		}
		optionLabel := value
		if label.Valid && label.String != "" {
			optionLabel = label.String
		}
		options = append(options, filterOption{Value: value, Label: optionLabel})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return options, nil
}

func mapFilterOptions[T ~struct {
	Value string
	Label string
}](options []filterOption) []T {
	mapped := make([]T, 0, len(options))
	for _, option := range options {
		mapped = append(mapped, T{Value: option.Value, Label: option.Label})
	}
	return mapped
}
