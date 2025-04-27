package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudcontrol"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/constants"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/aws/get"
	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools/aws/list"
	"github.com/ollama/ollama/api"
	"go.uber.org/zap"
)

var exitMessages = map[string]struct{}{
	"exit":   {},
	"quit":   {},
	"bye":    {},
	"cancel": {},
	"can":    {},
	"stop":   {},
}

func main() {
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "The URL of the Ollama server")
	modelName := flag.String("model", constants.DefaultModel, "The model to use for the LLM")
	debug := flag.Bool("debug", false, "Enable debug logging")
	prompt := flag.String("prompt", "", "The prompt to ask the LLM")
	flag.Parse()

	log := newLogger(*debug)
	defer log.Sync()

	awsConfig, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}

	// Create Ollama client
	ollamaClient, err := newOllamaClient(*ollamaURL)
	if err != nil {
		log.Panic("Error creating Ollama client", zap.Error(err))
	}

	// create service & tools
	awsListTool, err := list.NewTool(cloudformation.NewFromConfig(awsConfig), cloudcontrol.NewFromConfig(awsConfig))
	if err != nil {
		log.Panic("Error creating AWS list tool", zap.Error(err))
	}

	llmService, err := llm.NewOllamaService(log.Named("llmservice"), ollamaClient,
		llm.WithModel(*modelName),
		llm.WithSystemPrompt(`You are a helpful assistant that can answer questions about infrastructure resources, particularly but not exclusively those in AWS cloud.

The tools provided should be called multiple times if necessary to answer the question.

For example, if the user asks for details about all EC2 instances, you should first call the list_aws_resources tool to get a list of all the resources, and then call the get_aws_resource tool once for each resource in the list to get the details.

The tools do not have any context about the previous tool calls, so you must make sure to pass the correct parameters to each tool. For example, if you have already called the list_aws_resources tool, you must pass the list of resource identifiers to the get_aws_resource tool as they were returned by the list_aws_resources tool.

Pay particular attention to the names of the properties & parameters provided to you for each tool. If you get these wrong, the tool will fail. You must also ensure that any required parameters are passed to the tool.
`),
		// add tools
		// The DNS tool seems to trip LLMs up when they should be looking elsewhere
		// llm.WithToolFunction(dns.Tool{}),
		llm.WithToolFunction(awsListTool),
		llm.WithToolFunction(get.NewTool(cloudcontrol.NewFromConfig(awsConfig))),
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
	_, ok := exitMessages[strings.ToLower(msg)]
	return ok
}
