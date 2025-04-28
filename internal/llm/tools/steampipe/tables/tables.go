package tables

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
	"github.com/jackc/pgx/v5"
)

const (
	parameterResourceTypeFilter = "resource_type_filter"

	listTablesQuery = `
select
  foreign_table_name
from 
  information_schema.foreign_tables
where
  foreign_table_schema = 'aws'
`

	filteredListTablesQuery = listTablesQuery + `
and
  foreign_table_name LIKE $1;
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
	return "list_aws_tables"
}

func (t Tool) Description() string {
	return "Returns the list of tables containing AWS resources. The list of tables is returned as a JSON array of strings."
}

func (t Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			Name:        parameterResourceTypeFilter,
			Description: "The type of resource to filter by, e.g. \"ec2\" or \"s3\". This is used in a fuzzy search, with the query returning all tables containing the resource type in the name.",
			Required:    false,
			Type:        tools.ParameterTypeString,
		},
	}
}

func (t Tool) Call(ctx context.Context, parameters map[string]any) (string, error) {
	resourceTypeFilter, _ := parameters[parameterResourceTypeFilter].(string)

	query := listTablesQuery
	args := []any{}
	if resourceTypeFilter != "" {
		query = filteredListTablesQuery
		args = append(args, "%"+resourceTypeFilter+"%")
	}

	rows, err := t.db.Query(ctx, query, args...)
	if err != nil {
		return "", fmt.Errorf("error querying tables: %w", err)
	}

	resources := []string{}
	for rows.Next() {
		var resource string
		err := rows.Scan(&resource)
		if err != nil {
			return "", fmt.Errorf("error scanning table: %w", err)
		}
		resources = append(resources, resource)
	}

	if len(resources) == 0 {
		return "", fmt.Errorf("no resources found")
	}

	dataJSON, err := json.Marshal(resources)
	if err != nil {
		return "", fmt.Errorf("error marshalling resources to JSON: %w", err)
	}

	return string(dataJSON), nil
}
