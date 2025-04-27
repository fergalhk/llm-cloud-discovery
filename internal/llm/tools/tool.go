package tools

import (
	"context"
)

type (
	// Function defines an interface for a tool that can be used by an LLM.
	Function interface {
		Name() string
		Description() string
		ParameterDefinitions() []ParameterDefinition
		Call(ctx context.Context, parameters map[string]any) (string, error)
	}

	// ParameterDefinition defines an input parameter for a tool.
	ParameterDefinition struct {
		// The name of the parameter.
		Name string
		// The description of the parameter.
		Description string
		// The type of the parameter.
		Type ParameterType
		// If true, the parameter is marked as required for the tool.
		Required bool
		// If set, the parameter is marked as an enum and the values must be one of the values in the enum.
		Enum []any
	}

	ParameterType string
)

const (
	ParameterTypeString ParameterType = "string"
)

func (p ParameterType) String() string {
	return string(p)
}
