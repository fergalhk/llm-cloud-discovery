package dns

import (
	"context"
	"fmt"
	"net"

	"github.com/fergalhk/llm-cloud-discovery/internal/llm/tools"
)

type Tool struct{}

func (t Tool) Name() string {
	return "dns_record"
}

func (t Tool) Description() string {
	return "Get the DNS record for a given domain"
}

func (t Tool) ParameterDefinitions() []tools.ParameterDefinition {
	return []tools.ParameterDefinition{
		{
			Name:        "domain",
			Description: "The domain to get the DNS record for, e.g. google.com",
			Required:    true,
			Type:        tools.ParameterTypeString,
		},
	}
}

func (t Tool) Call(ctx context.Context, parameters map[string]any) (string, error) {
	domain, ok := parameters["domain"].(string)
	if !ok {
		return "", fmt.Errorf("format is not a string")
	}

	addrs, err := net.DefaultResolver.LookupHost(ctx, domain)
	if err != nil {
		return "", fmt.Errorf("error resolving host: %w", err)
	}

	out := ""
	for _, addr := range addrs {
		out += "* " + addr + "\n"
	}

	return out, nil
}
