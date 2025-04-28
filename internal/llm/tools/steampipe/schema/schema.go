package schema

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
	"github.com/jackc/pgx/v5"
)

const (
	parameterResourceType = "resource_type"

	getSchemaQuery = `
select
  column_name,
  data_type
from
  information_schema.columns
where table_name = $1 and table_schema = 'aws'
`
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
	return "get_aws_table_schema"
}

func (t Tool) Description() string {
	return "Returns the schema for a given AWS resource. The schema is returned as a JSON object, with the key representing the column name and the value representing the column type."
}

func (t Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			Name:        parameterResourceType,
			Description: "The exact name of the resource table to get the schema for.",
			Required:    true,
			Type:        tools.ParameterTypeString,
		},
	}
}

func (t Tool) Call(ctx context.Context, parameters map[string]any) (string, error) {
	resourceType, _ := parameters[parameterResourceType].(string)
	if resourceType == "" {
		return "", fmt.Errorf("resource type is required")
	}

	rows, err := t.db.Query(ctx, getSchemaQuery, resourceType)
	if err != nil {
		return "", fmt.Errorf("error querying tables: %w", err)
	}

	columnToDataType := make(map[string]string)
	for rows.Next() {
		var columnName, dataType string
		err := rows.Scan(&columnName, &dataType)
		if err != nil {
			return "", fmt.Errorf("error scanning table: %w", err)
		}
		columnToDataType[columnName] = dataType
	}

	if len(columnToDataType) == 0 {
		return "", fmt.Errorf("no columns found for resource type %s", resourceType)
	}

	dataJSON, err := json.Marshal(columnToDataType)
	if err != nil {
		return "", fmt.Errorf("error marshalling columns to JSON: %w", err)
	}

	return string(dataJSON), nil
}
