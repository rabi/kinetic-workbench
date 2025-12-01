# Workflow YAML System

This document explains how to create workflows from YAML definitions and how tools and workflows can be created separately and tied together.

## Terminology

- **Agent Definition** (`kind: Direct`): A YAML file that defines a single agent. It's a blueprint for one executable agent.
  - Example: `pr_fetcher.yaml` is an agent definition
  - Contains one agent configuration (name, instructions, tools, model)
  - When executed, becomes an agent (runtime entity)
  - Also called "agent executor" since it's an agent that executes
  
- **Workflow** (`kind: Sequential`, `kind: Parallel`, `kind: Composite`): A YAML file that orchestrates multiple agents/workflows.
  - Example: `pr_review_composed.yaml` is a workflow (orchestrates multiple agents)
  - Defines how multiple agents/workflows are executed (sequentially, in parallel, or composed)
  - A workflow orchestrates execution; an agent definition is just a single agent
  
- **Agent** (runtime): The executable entity built from an agent definition or workflow.
  - Built via `BuildAgent()` which takes a `WorkflowDefinition` and returns an `agent.Agent`
  - Implements the `agent.Agent` interface and can be executed via `Run()`
  
- **Tools**: Reusable functions that agents can call (e.g., `fetch_pull_request`, `get_pull_request_diff`)

**Key Distinction**:
- **Agent Definition** (`kind: Direct`) = Single agent blueprint → becomes one agent at runtime
- **Workflow** (`kind: Sequential/Parallel/Composite`) = Orchestration blueprint → becomes a workflow agent that orchestrates multiple agents at runtime

When you reference `../agents/pr_fetcher.yaml` (an agent definition) in a workflow, it gets loaded and becomes an agent that can be executed as part of that workflow.

## Architecture Overview

The workflow system follows a **separation of concerns** pattern:

1. **Tools** are created and registered independently
2. **Workflows** are defined in YAML and reference tools by name
3. **Registry** ties tools and workflows together at runtime
4. **Agents** are built from workflows and executed

## Creating Tools Separately

Tools can be created and registered independently of workflows:

```go
// Step 1: Create a tool registry
toolReg := registry.NewRegistry()

// Step 2: Register tools separately
// Option A: Auto-register from environment
toolReg.RegisterGitHubTools()  // Registers GitHub tools if credentials available
toolReg.RegisterSearchTools()  // Registers search tools if API key available

// Option B: Register custom tools programmatically
customTool := createMyCustomTool()
toolReg.RegisterTool("my_custom_tool", customTool)

// Step 3: Tools are now available for any workflow to use
```

## Creating Workflows from YAML

Workflows are defined declaratively in YAML files:

**Agent Definition (`kind: Direct`)** - Single agent (no `workflow:` wrapper needed):

```yaml
kind: Direct

name: ResearchAgent

stream: false

description: "A research agent designed to search, retrieve, and summarize information from the web."

agent:
  name: ResearchAgent
    description: "A deep research agent capable of gathering accurate information"
    instructions: |
      You are a research expert. Your task is to find accurate and up-to-date information.
      1. Search for relevant sources on the web.
      2. Extract key insights and summarize them concisely.
      3. Provide references and links to original sources.
    memory:
      kind: sliding_window
      parameters:
        window_size: 100
    model:
      kind: llm
      backend:
        kind: Cloud
      provider: OpenAI
      model_name: gpt-4o-mini
      parameters:
        temperature: 0.2
        max_tokens: 1500
    tools:
      - name: brave_search
    output:
      type: text

output:
  type: text
```

## Running Workflows

Use the CLI command to run workflows:

```bash
kind run --workflow workflow.yaml --input "What is Rust?"
```

Or programmatically:

```go
// Load workflow definition
workflowDef, err := workflow.LoadFromFile("workflow.yaml")

// Create builder with tool registry
builder := workflow.NewBuilderWithRegistry(toolReg)

// Build agent from workflow
agent, err := builder.BuildAgent(ctx, workflowDef)

// Run the agent
// ... (use agent.Run() with appropriate context)
```

## Benefits of Separation

1. **Modularity**: Tools can be developed and tested independently
2. **Reusability**: Same tools can be used across multiple workflows
3. **Flexibility**: Workflows can be defined declaratively without code changes
4. **Maintainability**: Tools and workflows can be updated independently
5. **Conditional Registration**: Tools can be registered based on environment/configuration

## Available Tools

Tools are automatically registered if their required credentials are available:

- **GitHub Tools**: `fetch_pull_request`, `get_pull_request_diff`, `list_merged_prs`, `check_cherry_pick_conflicts`, `create_cherry_pick_pr`
  - Requires: `GITHUB_TOKEN`, `GITHUB_ORG`, `GITHUB_REPO`

- **Search Tools**: `brave_search`
  - Requires: `BRAVE_API_KEY`

## Supported Model Providers

- **OpenAI**: `provider: OpenAI`
  - Requires: `OPENAI_API_KEY`
  - Models: `gpt-4o-mini`, `gpt-4o`, `gpt-4`, etc.

- **DeepSeek**: `provider: DeepSeek`
  - Requires: `DEEPSEEK_API_KEY`
  - Models: `deepseek-chat`, etc.

- **Gemini/Google**: `provider: Gemini` or `provider: Google`
  - Requires: `GOOGLE_API_KEY`
  - Models: `gemini-3-pro-preview`, etc.

## Example: Complete Workflow

```go
package main

import (
    "context"
    "kinetic-workbench/pkg/tools/registry"
    "kinetic-workbench/pkg/workflow"
)

func main() {
    ctx := context.Background()
    
    // 1. Create tool registry
    toolReg := registry.NewRegistry()
    
    // 2. Register tools (can be done separately)
    toolReg.RegisterGitHubTools()
    toolReg.RegisterSearchTools()
    
    // 3. Create workflow builder with registry
    builder := workflow.NewBuilderWithRegistry(toolReg)
    
    // 4. Load workflow from YAML
    workflowDef, _ := workflow.LoadFromFile("research_workflow.yaml")
    
    // 5. Build agent (tools are automatically resolved from registry)
    agent, _ := builder.BuildAgent(ctx, workflowDef)
    
    // 6. Use the agent
    // ... agent is ready to use
}
```

## Workflow YAML Schema

### Top-level Fields

- `kind`: Workflow type (currently supports `Direct`)
- `name`: Workflow name
- `stream`: Whether to stream responses
- `description`: Workflow description
- `workflow`: Agent definition
- `output`: Output configuration

### Agent Definition

- `name`: Agent name
- `description`: Agent description
- `instructions`: System instructions for the agent
- `memory`: Memory configuration
  - `kind`: Memory type (`sliding_window`)
  - `parameters`: Memory parameters
- `model`: Model configuration
  - `kind`: Model kind (`llm`)
  - `backend`: Backend configuration
  - `provider`: Provider name (`OpenAI`, `DeepSeek`, `Gemini`)
  - `model_name`: Specific model name
  - `parameters`: Model parameters (temperature, max_tokens, etc.)
- `tools`: List of tool names (strings) that reference registered tools
- `output`: Output configuration

