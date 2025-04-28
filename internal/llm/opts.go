package llm

import (
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/constants"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
)

type (
	Opt        func(*ollamaOpts)
	ollamaOpts struct {
		model         string
		systemPrompt  string
		toolFunctions []tools.Function
	}
)

func newOllamaOpts(opts ...Opt) ollamaOpts {
	o := ollamaOpts{
		model: constants.DefaultModel,
	}

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func WithModel(model string) Opt {
	return func(o *ollamaOpts) {
		o.model = model
	}
}

func WithSystemPrompt(systemPrompt string) Opt {
	return func(o *ollamaOpts) {
		o.systemPrompt = systemPrompt
	}
}

func WithToolFunction(toolFunctions ...tools.Function) Opt {
	return func(o *ollamaOpts) {
		o.toolFunctions = append(o.toolFunctions, toolFunctions...)
	}
}
