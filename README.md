# Kinetic

Kinetic is an intelligent agent orchestrator that uses LLM-based routing to handle GitHub-related tasks, including PR reviews and cherry-pick operations. Built with Google's Agent Development Kit (ADK), it provides a flexible framework for managing multiple agent workflows.

## Features

- **LLM-Based Router**: Intelligently routes user requests to appropriate workflows using natural language understanding
- **PR Review Workflow**: Automated pull request review with code analysis and feedback
- **Cherry-Pick Workflow**: Find merged PRs and create cherry-pick PRs automatically
- **Multi-Provider Support**: Supports multiple LLM providers (DeepSeek, Google Gemini)
- **GitHub Integration**: Seamless integration with GitHub API for PR operations

## Architecture

Kinetic uses a hierarchical agent architecture:

```
┌─────────────────┐
│  Router Agent   │ (LLM-based intent detection)
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

- **Router Agent**: Analyzes user intent and routes to appropriate workflow
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
# Options: "deepseek" or "gemini" (default: deepseek)
MODEL_PROVIDER=gemini
```

### DeepSeek Configuration

```bash
DEEPSEEK_API_KEY=your_deepseek_api_key
DEEPSEEK_MODEL=deepseek-chat
```

### Google Gemini Configuration

```bash
GOOGLE_API_KEY=your_google_api_key
GEMINI_MODEL=gemini-3-pro-preview
```

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
User: "review PR 1060"
```

The router agent will detect this as a review request and route it to the PR Review workflow, which will:
1. Fetch PR details from GitHub
2. Analyze the code changes
3. Provide review feedback

#### Cherry-Pick Operations
```
User: "find merged PRs from last week"
User: "create cherry-pick for PR 1060"
```

The router agent will detect these as cherry-pick requests and route to the Cherry-Pick workflow.

## Development

### Project Structure

```
kinetic/
├── cmd/
│   └── kinetic/          # Main orchestrator entry point
│       └── main.go
├── internal/
│   ├── agents/           # Agent implementations
│   │   ├── router.go     # Router agent
│   │   ├── pr_reviewer.go
│   │   ├── pr_fetcher.go
│   │   ├── cherry_pick.go
│   │   └── workflow.go
│   ├── providers/        # LLM provider implementations
│   │   ├── factory.go
│   │   └── deepseek.go
│   └── tools/            # Tool implementations
│       └── github/
│           ├── client.go
│           └── tools.go
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

1. Create a new agent file in `internal/agents/`
2. Implement the `agent.Agent` interface
3. Register it in the router or create a new workflow

### Adding New LLM Providers

1. Create a provider implementation in `internal/providers/`
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

