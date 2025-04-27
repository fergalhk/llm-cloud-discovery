package get

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
)

const (
	parameterResourceType       = "resource_type"
	parameterResourceIdentifier = "resource_identifier"
)

type Tool struct {
	cloudcontrolClient *cloudcontrol.Client
}

func NewTool(cloudcontrolClient *cloudcontrol.Client) tools.Function {
	return &Tool{
		cloudcontrolClient: cloudcontrolClient,
	}
}

func (t *Tool) Name() string {
	return "get_aws_resource"
}

func (t *Tool) Description() string {
	return fmt.Sprintf(`This tool allows all of the properties of a specific AWS resource to be retrieved.
The tool returns a JSON object containing the resource's properties.
You must provide both the %q and %q parameters.
The %q parameter is the same resource type used for the list_aws_resources tool.`,
		parameterResourceIdentifier, parameterResourceType, parameterResourceType)
}

func (t *Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			Name:        parameterResourceType,
			Description: "The type of resource to retrieve. This is in the format of AWS::Service::ResourceType, for example AWS::EC2::Instance.",
			Required:    true,
			Type:        tools.ParameterTypeString,
		},
		{
			Name:        parameterResourceIdentifier,
			Description: "The identifier of the resource to retrieve.",
			Required:    true,
			Type:        tools.ParameterTypeString,
		},
	}
}

func (t *Tool) Call(ctx context.Context, parameters map[string]any) (string, error) {
	resourceType, ok := parameters[parameterResourceType].(string)
	if !ok {
		return "", fmt.Errorf("%s is not a valid string", parameterResourceType)
	}

	resourceIdentifier, ok := parameters[parameterResourceIdentifier].(string)
	if !ok {
		return "", fmt.Errorf("%s is not a valid string", parameterResourceIdentifier)
	}

	resp, err := t.cloudcontrolClient.GetResource(ctx, &cloudcontrol.GetResourceInput{
		TypeName:   &resourceType,
		Identifier: &resourceIdentifier,
	})
	if err != nil {
		return "", fmt.Errorf("error getting resource: %w", err)
	}

	return *resp.ResourceDescription.Properties, nil
}
