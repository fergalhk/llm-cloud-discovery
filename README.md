# Cloud discovery LLM server

A sandbox LLM app, with the goal of creating a local assistant that can reason about questions that require correlating unstructured metadata between multiple AWS resources & across service boundaries - for example, _which ECS services connect to the `users` DynamoDB table?_.

### Prerequisites

* Install Ollama:

```bash
curl -fsSL https://ollama.com/install.sh | sh
```