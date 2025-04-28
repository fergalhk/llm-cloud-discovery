package main

import (
	"context"
	"os"

	"github.com/fergalhk/llm-cloud-discovery/internal/cmd"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/steampipe/dml"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/steampipe/schema"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/steampipe/tables"
	"github.com/jackc/pgx/v5"
)

func main() {
	dbConnStr := os.Getenv("STEAMPIPE_DB")
	if dbConnStr == "" {
		panic("STEAMPIPE_DB is not set")
	}

	db, err := pgx.Connect(context.Background(), dbConnStr)
	if err != nil {
		panic(err)
	}

	err = db.Ping(context.Background())
	if err != nil {
		panic(err)
	}

	cmd.Run(
		`You are a helpful assistant that can answer questions about infrastructure resources, particularly but not exclusively those in AWS cloud.

You have been provided with 3 tools that allow you to query a PostgreSQL database containing AWS resource data. The tables are always prefixed with "aws_".

To use the tools, you should follow this process:

1. Use the list_aws_tables tool to get a list of all the tables in the database. The number of tables is very large, so you should filter the results to narrow down the list. For example, if you're looking for EC2 data, you should use a resource_type_filter of ec2.
2. Get the schema of the table you need using the get_aws_table_schema tool.
3. Construct a SQL query, and run it using the execute_aws_query tool. Note that many resources need joins between multiple tables to return the correct data, so you should use these if necessary.

The tools provided should be called multiple times if necessary to answer the question.

`,
		dml.New(db),
		schema.New(db),
		tables.New(db),
	)
}
