package tools

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
)

type Registry struct {
	nameToTool  map[string]Function
	ollamaTools api.Tools
}

func NewRegistry(tools ...Function) *Registry {
	nameToTool := make(map[string]Function)
	for _, tool := range tools {
		nameToTool[tool.Name()] = tool
	}

	return &Registry{
		nameToTool: nameToTool,
	}
}

func (r *Registry) OllamaTools() (api.Tools, error) {
	err := r.initOllamaTools()
	if err != nil {
		return nil, fmt.Errorf("error initializing ollama tools: %w", err)
	}

	return r.ollamaTools, nil
}

func (r *Registry) initOllamaTools() error {
	if len(r.ollamaTools) >= len(r.nameToTool) {
		return nil
	}

	tools := make(api.Tools, 0, len(r.nameToTool))
	for _, tool := range r.nameToTool {
		ollamaTool, err := toOllamaTool(tool)
		if err != nil {
			return fmt.Errorf("error generating ollama tool for %q: %w", tool.Name(), err)
		}

		tools = append(tools, ollamaTool)
	}

	r.ollamaTools = tools

	return nil
}

func (r *Registry) Call(ctx context.Context, name string, parameters map[string]any) (string, error) {
	tool, ok := r.nameToTool[name]
	if !ok {
		return "", fmt.Errorf("tool %q not found", name)
	}

	return tool.Call(ctx, parameters)
}
