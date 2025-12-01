# Kinetic

Kinetic is an intelligent agent orchestrator for handling GitHub-related tasks, including PR reviews and cherry-pick operations. Built with Google's Agent Development Kit (ADK), it provides a flexible YAML-based framework for defining and managing agent workflows.

## Features

- **YAML-Based Workflows**: Define workflows declaratively in YAML files
- **PR Review Workflow**: Automated pull request review with code analysis and feedback
- **Cherry-Pick Workflow**: Find merged PRs and create cherry-pick PRs automatically
- **Multi-Provider Support**: Supports multiple LLM providers (DeepSeek, Google Gemini, OpenAI)
- **GitHub Integration**: Seamless integration with GitHub API for PR operations
- **Workflow Composition**: Compose complex workflows from simpler components

## Architecture

Kinetic uses YAML-defined workflows that can be composed:

```
┌─────────────────┐
│  YAML Workflow  │ (Defines agent flow)
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
┌─────────┐ ┌──────────────┐
│ Review  │ │ Cherry-Pick │
│Workflow │ │  Workflow   │
└────┬────┘ └──────┬───────┘
     │             │
     ▼             ▼
┌─────────┐   ┌─────────────┐
│PR Fetcher│   │Cherry-Pick  │
│PR Reviewer│  │   Agent     │
└─────────┘   └─────────────┘
```

### Components

- **Workflow Builder**: Loads and builds agents from YAML definitions
- **PR Review Workflow**: Sequential execution of PR Fetcher → PR Reviewer agents
- **Cherry-Pick Workflow**: Handles finding merged PRs and creating cherry-pick PRs
- **GitHub Tool**: Provides GitHub API integration for agents

## Prerequisites

- Go 1.24.4 or later
- GitHub Personal Access Token (PAT) with appropriate scopes
- API key for your chosen LLM provider (DeepSeek or Google Gemini)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/rabi/kinetic.git
cd kinetic
```

2. Install dependencies:
```bash
go mod download
```

3. Build the binary:
```bash
make build
```

The binary will be created at `bin/kinetic`.

## Configuration

Create an `env` file in the project root with the following variables:

### Model Provider Configuration

```bash
# Options: "gemini", "deepseek", or "openai" (default: gemini)
MODEL_PROVIDER=gemini

# Generic model name (used if provider-specific env var is not set)
MODEL_NAME=gemini-3-pro-preview
```

### OpenAI Configuration

```bash
OPENAI_API_KEY=your_openai_api_key
OPENAI_MODEL=gpt-4o-mini  # Optional: defaults to MODEL_NAME if not set
```

### DeepSeek Configuration

```bash
DEEPSEEK_API_KEY=your_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat  # Optional: defaults to MODEL_NAME if not set
```

### Google Gemini Configuration

```bash
GOOGLE_API_KEY=your_google_api_key
GEMINI_MODEL=gemini-3-pro-preview  # Optional: defaults to MODEL_NAME if not set
```

**Note**: 
- **Gemini is the default provider** - if `provider` is not specified in workflow YAML, Gemini will be used
- Model names should NOT be hardcoded in workflow YAML files. They are read from environment variables in this order:
  1. Provider-specific env var (e.g., `GEMINI_MODEL`, `OPENAI_MODEL`, `DEEPSEEK_MODEL`)
  2. Generic `MODEL_NAME` env var
  3. Provider default (Gemini defaults to `gemini-3-pro-preview`)

### GitHub Configuration

```bash
GITHUB_TOKEN=your_github_personal_access_token
GITHUB_ORG=your_organization_or_username
GITHUB_REPO=your_repository_name
```

### Agent Configuration

```bash
LOG_LEVEL=info
DEBUG=false
```

**Note**: The `env` file is gitignored to protect your secrets. See [GITHUB_SETUP.md](GITHUB_SETUP.md) for detailed setup instructions.

## Usage

### Running the Application

```bash
make run
```

Or run directly:

```bash
go run ./cmd/kinetic
```

### Example Interactions

#### PR Review
```
$ make run WORKFLOW=examples/pr_review_composed.yaml INPUT="review PR 1060"
```

The PR Review workflow will:
1. Fetch PR details from GitHub
2. Analyze the code changes
3. Provide review feedback

#### Cherry-Pick Operations
```
$ make run WORKFLOW=examples/cherry_pick.yaml INPUT="find merged PRs from last week"
```

Workflows are defined in YAML files and can be composed together.

## Development

### Project Structure

```
kinetic/
├── cmd/
│   └── kinetic/          # Main orchestrator entry point
│       └── main.go
├── pkg/
│   ├── providers/        # LLM provider implementations
│   │   ├── factory.go
│   │   ├── deepseek.go
│   │   └── openai.go
│   ├── tools/            # Tool implementations
│   │   ├── github/       # GitHub API tools
│   │   │   ├── client.go
│   │   │   └── tools.go
│   │   ├── registry/     # Tool registry
│   │   └── search/       # Search tools (Brave)
│   └── workflow/         # Workflow builder (loads agents from YAML)
│       ├── builder.go
│       ├── loader.go
│       ├── yaml.go
│       └── parallel_agent.go
├── Makefile
├── go.mod
└── README.md
```

### Available Make Targets

- `make build` - Build the binary
- `make run` - Run the application
- `make test` - Run tests
- `make fmt` - Format code
- `make vet` - Run go vet
- `make lint` - Run golangci-lint (if installed)
- `make clean` - Remove build artifacts
- `make tidy` - Tidy go.mod
- `make help` - Show all available targets

### Adding New Agents

1. Create a new agent definition YAML file in `agents/` (or your workflow directory)
2. Define the agent configuration (name, instructions, tools, model)
3. Reference the agent definition YAML file in workflows or run it directly

Agents are created generically from YAML definitions - no Go code needed!

### Adding New LLM Providers

1. Create a provider implementation in `pkg/providers/`
2. Add the provider case to `CreateModel()` in `factory.go`
3. Update environment variable documentation

## Documentation

- [GITHUB_SETUP.md](GITHUB_SETUP.md) - Detailed GitHub tool setup guide
- [MESSAGE_FLOW.md](MESSAGE_FLOW.md) - Message flow and architecture documentation

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[Add your license here]

## Acknowledgments

- Built with [Google Agent Development Kit (ADK)](https://github.com/google/adk)
- Uses [go-github](https://github.com/google/go-github) for GitHub API integration

