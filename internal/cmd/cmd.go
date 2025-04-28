package cmd

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/constants"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

var (
	exitMessages = map[string]struct{}{
		"exit":   {},
		"quit":   {},
		"bye":    {},
		"cancel": {},
		"can":    {},
		"stop":   {},
	}
	resetMessages = map[string]struct{}{
		"reset":  {},
		"clear":  {},
		"forget": {},
	}
)

func Run(systemPrompt string, toolFunctions ...tools.Function) {
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "The URL of the Ollama server")
	modelName := flag.String("model", constants.DefaultModel, "The model to use for the LLM")
	debug := flag.Bool("debug", false, "Enable debug logging")
	prompt := flag.String("prompt", "", "The prompt to ask the LLM")
	flag.Parse()

	log := newLogger(*debug)
	defer log.Sync()

	// Create Ollama client
	ollamaClient, err := newOllamaClient(*ollamaURL)
	if err != nil {
		log.Panic("Error creating Ollama client", zap.Error(err))
	}

	llmService, err := llm.NewOllamaService(log.Named("llmservice"), ollamaClient,
		llm.WithModel(*modelName),
		llm.WithSystemPrompt(systemPrompt),
		llm.WithToolFunction(toolFunctions...),
	)

	if err != nil {
		log.Panic("Error creating LLM service", zap.Error(err))
	}

	oneShot := *prompt != ""
	if oneShot {
		log.Debug("Running in one-shot mode", zap.String("prompt", *prompt))
		// non-interactive mode, ask question then exit after response
		response, err := llmService.Chat(context.Background(), *prompt)
		if err != nil {
			log.Panic("Error calling chat", zap.Error(err))
		}
		fmt.Println(response)
		return
	}

	// interactive mode, accept input in a REPL
	log.Debug("Running in interactive mode")
	for {
		promptMsg, err := readStdin()
		if err != nil {
			log.Panic("Error reading input", zap.Error(err))
		}

		if isExitMessage(promptMsg) {
			log.Info("Exiting")
			break
		}

		if isResetMessage(promptMsg) {
			log.Info("Resetting")
			llmService.Reset()
			continue
		}

		response, err := llmService.Chat(context.Background(), promptMsg)
		if err != nil {
			log.Error("Error calling chat", zap.Error(err))
		}
		fmt.Println(response)
	}
}

func newOllamaClient(ollamaURL string) (*api.Client, error) {
	apiURL, err := url.Parse(ollamaURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing ollama URL: %w", err)
	}

	return api.NewClient(apiURL, new(http.Client)), nil
}

func newLogger(debug bool) *zap.Logger {
	cfg := zap.NewProductionConfig()
	if debug {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func readStdin() (string, error) {
	fmt.Print("> ")
	reader := bufio.NewReader(os.Stdin)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading input: %w", err)
	}
	return strings.TrimSpace(msg), nil
}

func isExitMessage(msg string) bool {
	return setContains(msg, exitMessages)
}

func isResetMessage(msg string) bool {
	return setContains(msg, resetMessages)
}

func setContains(elem string, set map[string]struct{}) bool {
	_, ok := set[strings.ToLower(elem)]
	return ok
}
