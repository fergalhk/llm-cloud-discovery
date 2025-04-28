package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fergalhk/llm-cloud-discovery/internal/cmd"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/aws/get"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/aws/list"
)

func main() {
	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// create service & tools
	awsListTool, err := list.NewTool(cloudformation.NewFromConfig(awsConfig), cloudcontrol.NewFromConfig(awsConfig))
	if err != nil {
		panic(err)
	}

	cmd.Run(
		`You are a helpful assistant that can answer questions about infrastructure resources, particularly but not exclusively those in AWS cloud.

The tools provided should be called multiple times if necessary to answer the question.

For example, if the user asks for details about all EC2 instances, you should first call the list_aws_resources tool to get a list of all the resources, and then call the get_aws_resource tool once for each resource in the list to get the details.

The tools do not have any context about the previous tool calls, so you must make sure to pass the correct parameters to each tool. For example, if you have already called the list_aws_resources tool, you must pass the list of resource identifiers to the get_aws_resource tool as they were returned by the list_aws_resources tool.

Pay particular attention to the names of the properties & parameters provided to you for each tool. If you get these wrong, the tool will fail. You must also ensure that any required parameters are passed to the tool.
`,
		awsListTool,
		get.NewTool(cloudcontrol.NewFromConfig(awsConfig)),
	)
}
