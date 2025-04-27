package llm

import (
	"context"
	"fmt"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"
)

type (
	Service interface {
		Chat(ctx context.Context, prompt string) (string, error)
	}
	ollamaService struct {
		log          *zap.Logger
		model        string
		toolRegistry *tools.Registry
		ollamaClient *api.Client
		systemPrompt string
	}
)

func NewOllamaService(log *zap.Logger, cl *api.Client, opts ...Opt) Service {
	o := newOllamaOpts(opts...)

	return &ollamaService{
		log:          log,
		model:        o.model,
		toolRegistry: tools.NewRegistry(o.toolFunctions...),
		ollamaClient: cl,
		systemPrompt: o.systemPrompt,
	}
}

func (c *ollamaService) Chat(ctx context.Context, prompt string) (string, error) {
	ollamaTools, err := c.toolRegistry.OllamaTools()
	if err != nil {
		return "", fmt.Errorf("error generating tools: %w", err)
	}
	c.log.Debug("Built tools", zap.Any("tools", ollamaTools))

	baseMessages := []api.Message{}
	if c.systemPrompt != "" {
		c.log.Debug("Adding system prompt", zap.String("prompt", c.systemPrompt))
		baseMessages = append(baseMessages, api.Message{Role: "system", Content: c.systemPrompt})
	}

	finalResponse, err := c.doChatWithTools(ctx, ollamaTools, append(baseMessages, api.Message{Role: "user", Content: prompt})...)
	if err != nil {
		return "", fmt.Errorf("error calling chat: %w", err)
	}

	return finalResponse.Message.Content, nil
}

func (c *ollamaService) doChatWithTools(ctx context.Context, ollamaTools api.Tools, messages ...api.Message) (api.ChatResponse, error) {
	for {
		chatRequest := &api.ChatRequest{
			Model:    c.model,
			Messages: messages,
			Stream:   ptr.To(false),
			Tools:    ollamaTools,
		}

		chatResponse, err := c.callChat(ctx, chatRequest)
		if err != nil {
			return api.ChatResponse{}, fmt.Errorf("error calling chat API: %w", err)
		}

		if len(chatResponse.Message.ToolCalls) > 0 {
			// if there's a tool call in the response, append the system message & tool calls,
			// then call the API again with the new messages.
			c.log.Debug("Chat API returned tool calls, invoking tools")
			messages = append(messages, chatResponse.Message)
			for i, toolCall := range chatResponse.Message.ToolCalls {
				c.log.Debug("Invoking tool", zap.String("tool", toolCall.Function.Name), zap.Int("tool_call_index", i), zap.Any("arguments", toolCall.Function.Arguments))
				toolResult, err := c.toolRegistry.Call(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
				if err != nil {
					return api.ChatResponse{}, fmt.Errorf("error calling tool %q (%d): %w", toolCall.Function.Name, i, err)
				}

				messages = append(messages, api.Message{
					Role:    "tool",
					Content: toolResult,
				})
				c.log.Debug("Tool call complete", zap.String("call_result", toolResult))
			}
		} else {
			// if there's no tool call in the response, we're done.
			c.log.Debug("Chat API returned no tool calls, returning final response")
			return chatResponse, nil
		}
	}
}

func (c *ollamaService) callChat(ctx context.Context, chatRequest *api.ChatRequest) (api.ChatResponse, error) {
	c.log.Debug("Calling chat API", zap.Any("messages", chatRequest.Messages))
	var chatResponse api.ChatResponse
	err := c.ollamaClient.Chat(ctx, chatRequest, func(cr api.ChatResponse) error {
		chatResponse = cr
		return nil
	})

	if err != nil {
		return api.ChatResponse{}, fmt.Errorf("error calling ollama: %w", err)
	}

	return chatResponse, nil
}
