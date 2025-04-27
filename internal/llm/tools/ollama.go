package tools

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/ollama/ollama/api"
)

type (
	ollamaTool struct {
		Type     string             `json:"type"`
		Function ollamaToolFunction `json:"function"`
	}
	ollamaToolFunction struct {
		Name        string           `json:"name"`
		Description string           `json:"description"`
		Parameters  ollamaParameters `json:"parameters"`
	}
	ollamaParameters struct {
		Type       string                    `json:"type"`
		Defs       any                       `json:"$defs,omitempty"`
		Items      any                       `json:"items,omitempty"`
		Required   []string                  `json:"required"`
		Properties map[string]ollamaProperty `json:"properties"`
	}
	ollamaProperty struct {
		Type        api.PropertyType `json:"type"`
		Description string           `json:"description"`
		Enum        []any            `json:"enum,omitempty"`
	}
)

func toOllamaTool(function Function) (api.Tool, error) {
	parameters := function.ParameterDefinitions()
	requiredParameters := make([]string, 0, len(parameters))
	properties := make(map[string]ollamaProperty)
	for _, parameter := range parameters {
		if parameter.Required {
			requiredParameters = append(requiredParameters, parameter.Name)
		}
		properties[parameter.Name] = ollamaProperty{
			Type:        api.PropertyType{parameter.Type.String()},
			Description: parameter.Description,
			Enum:        parameter.Enum,
		}
	}

	t := ollamaTool{
		Type: "function",
		Function: ollamaToolFunction{
			Name:        function.Name(),
			Description: function.Description(),
			Parameters: ollamaParameters{
				Type:       "object",
				Required:   requiredParameters,
				Properties: properties,
			},
		},
	}

	ot := api.Tool{}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result: &ot,
	})
	if err != nil {
		return api.Tool{}, fmt.Errorf("error creating decoder: %w", err)
	}

	err = decoder.Decode(t)
	if err != nil {
		return api.Tool{}, fmt.Errorf("error decoding into ollama struct: %w", err)
	}

	return ot, nil
}
