package list

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
)

const parameterResourceType = "resource_type"

type Tool struct {
	cloudcontrolClient *cloudcontrol.Client
	validResourceTypes []string
}

func NewTool(cloudformationClient *cloudformation.Client, cloudcontrolClient *cloudcontrol.Client) (tools.Function, error) {
	resourceTypes, err := validResourceTypes(context.Background(), cloudformationClient)
	if err != nil {
		return nil, err
	}

	return &Tool{
		cloudcontrolClient: cloudcontrolClient,
		validResourceTypes: resourceTypes,
	}, nil
}

func (t *Tool) Name() string {
	return "list_aws_resources"
}

func (t *Tool) Description() string {
	return "This tool retrieves a list of identifiers for all resources in AWS of a given type. The list is returned as a JSON array of strings, each of which is the identifier of a single resource."
}

func (t *Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			// note - we can't use Enum, as the resulting list is so large that it causes the input to be truncated
			Name:        parameterResourceType,
			Description: "The type of resource to list or retrieve. This is in the format of AWS::Service::ResourceType, for example AWS::EC2::Instance. Resource types always begin with `AWS::`.",
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

	if !slices.Contains(t.validResourceTypes, resourceType) {
		return "", fmt.Errorf("%s is not a valid resource type", resourceType)
	}

	paginator := cloudcontrol.NewListResourcesPaginator(t.cloudcontrolClient, &cloudcontrol.ListResourcesInput{
		TypeName: &resourceType,
	})

	resourceIdentifiers := []string{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return "", fmt.Errorf("error listing resources: %w", err)
		}

		for _, r := range page.ResourceDescriptions {
			resourceIdentifiers = append(resourceIdentifiers, *r.Identifier)
		}
	}

	resourceIdentifiersJSON, err := json.Marshal(resourceIdentifiers)
	if err != nil {
		return "", fmt.Errorf("error marshalling resource identifiers to JSON: %w", err)
	}

	return string(resourceIdentifiersJSON), nil
}

func validResourceTypes(ctx context.Context, cloudformationClient *cloudformation.Client) ([]string, error) {
	paginator := cloudformation.NewListTypesPaginator(cloudformationClient, &cloudformation.ListTypesInput{
		Visibility: types.VisibilityPublic,
	})

	resourceTypes := []string{}
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, t := range page.TypeSummaries {
			if t.Type == types.RegistryTypeResource {
				resourceTypes = append(resourceTypes, *t.TypeName)
			}
		}
	}

	return resourceTypes, nil
}

func toAnySlice[T any](sl []T) []any {
	out := []any{}
	for _, v := range sl {
		out = append(out, v)
	}
	return out
}
