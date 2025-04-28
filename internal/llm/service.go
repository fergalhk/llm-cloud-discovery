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
		ollamaTools  api.Tools
		ollamaClient *api.Client

		// messages is the history of messages, including the system prompt & the LLM responses
		messages []api.Message
	}
)

func NewOllamaService(log *zap.Logger, cl *api.Client, opts ...Opt) (Service, error) {
	o := newOllamaOpts(opts...)

	svc := &ollamaService{
		log:          log,
		model:        o.model,
		toolRegistry: tools.NewRegistry(o.toolFunctions...),
		ollamaClient: cl,
	}

	ollamaTools, err := svc.toolRegistry.OllamaTools()
	if err != nil {
		return nil, fmt.Errorf("error generating tools: %w", err)
	}
	svc.log.Debug("Built tools", zap.Any("tools", ollamaTools))
	svc.ollamaTools = ollamaTools

	if o.systemPrompt != "" {
		svc.log.Debug("Adding system prompt", zap.String("prompt", o.systemPrompt))
		svc.pushMessage("system", o.systemPrompt)
	}

	return svc, nil
}

func (c *ollamaService) Chat(ctx context.Context, prompt string) (string, error) {
	c.pushMessage("user", prompt)
	finalResponse, err := c.doChatWithTools(ctx)
	if err != nil {
		return "", fmt.Errorf("error calling chat: %w", err)
	}

	return finalResponse.Message.Content, nil
}

func (c *ollamaService) doChatWithTools(ctx context.Context) (api.ChatResponse, error) {
	for {
		chatResponse, err := c.callChat(ctx)
		if err != nil {
			return api.ChatResponse{}, fmt.Errorf("error calling chat API: %w", err)
		}

		if len(chatResponse.Message.ToolCalls) == 0 {
			// if there's no tool call in the response, we're done.
			c.log.Debug("Chat API returned no tool calls, returning final response")
			return chatResponse, nil
		}

		// if there's a tool call in the response, append the system message & tool calls,
		// then call the API again with the new messages.
		c.log.Debug("Chat API returned tool calls, invoking tools")
		c.pushRawMessage(chatResponse.Message)
		for i, toolCall := range chatResponse.Message.ToolCalls {
			log := c.log.With(zap.String("tool", toolCall.Function.Name), zap.Int("tool_call_index", i), zap.Any("arguments", toolCall.Function.Arguments))
			log.Debug("Invoking tool")
			toolResult, err := c.toolRegistry.Call(ctx, toolCall.Function.Name, toolCall.Function.Arguments)
			if err != nil {
				log.Warn("Tool call returned error, returning error to LLM", zap.Error(err))
				c.pushMessage("tool", fmt.Sprintf("Error calling tool %q: %s", toolCall.Function.Name, err))
			} else {
				log.Debug("Tool call complete", zap.String("call_result", toolResult))
				c.pushMessage("tool", toolResult)
			}

		}
	}
}

func (c *ollamaService) callChat(ctx context.Context) (api.ChatResponse, error) {
	chatRequest := &api.ChatRequest{
		Model:    c.model,
		Messages: c.messages,
		Stream:   ptr.To(false),
		Tools:    c.ollamaTools,
	}

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

func (c *ollamaService) pushMessage(role, content string) {
	c.pushRawMessage(api.Message{Role: role, Content: content})
}

func (c *ollamaService) pushRawMessage(message api.Message) {
	c.messages = append(c.messages, message)
}
