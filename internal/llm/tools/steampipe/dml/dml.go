package dml

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
	"github.com/jackc/pgx/v5"
)

const (
	parameterQuery = "query"
)

type Tool struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) tools.Function {
	return &Tool{
		db: db,
	}
}

func (t Tool) Name() string {
	return "execute_aws_query"
}

func (t Tool) Description() string {
	return "Executes arbitrary SQL queries against the AWS resources tables. The response is returned as a JSON array of objects, with each object representing a row in the result set."
}

func (t Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			Name:        parameterQuery,
			Description: "The SQL query to execute.",
			Required:    true,
			Type:        tools.ParameterTypeString,
		},
	}
}

func (t Tool) Call(ctx context.Context, parameters map[string]any) (string, error) {
	query, _ := parameters[parameterQuery].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	rows, err := t.db.Query(ctx, query)
	if err != nil {
		return "", fmt.Errorf("error executing query: %w", err)
	}

	data := []map[string]any{}
	for rows.Next() {
		row, err := scanArbitraryRow(rows)
		if err != nil {
			return "", fmt.Errorf("error scanning row: %w", err)
		}
		data = append(data, row)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("no rows returned")
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("error marshalling data to JSON: %w", err)
	}

	return string(dataJSON), nil
}

func scanArbitraryRow(rows pgx.Rows) (map[string]any, error) {
	values := make([]any, len(rows.FieldDescriptions()))
	for i := range values {
		values[i] = new(any)
	}
	err := rows.Scan(values...)
	if err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	valuesMap := make(map[string]any)
	for i, field := range rows.FieldDescriptions() {
		valuesMap[field.Name] = values[i]
	}

	return valuesMap, nil
}
