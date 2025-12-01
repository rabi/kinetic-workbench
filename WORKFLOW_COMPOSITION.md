# Workflow Composition

This document explains how to compose workflows from reusable components, inspired by AutoAgents patterns.

## Terminology

- **Agent Definition** (`kind: Direct`): A YAML file defining a single agent (e.g., `pr_fetcher.yaml`). Also called "agent executor" since it's an executable agent.
- **Workflow** (`kind: Sequential/Parallel/Composite`): A YAML file orchestrating multiple agents/workflows.
- **Component**: A reusable agent definition or workflow that can be referenced by other workflows.

## Overview

Workflow composition allows you to:
- **Reuse agent definitions**: Define common agents once, use them everywhere
- **Reuse workflows**: Compose workflows from other workflows
- **Build complex workflows**: Compose simple components into complex ones
- **Execute sequentially or in parallel**: Control execution order and concurrency
- **Override configurations**: Customize referenced components with overrides
- **Maintain modularity**: Update component logic in one place

## Execution Modes

Composite workflows support two execution modes:

1. **Sequential** (default): Workflows execute one after another, with output flowing from one to the next
2. **Parallel**: Workflows execute concurrently, with outputs combined after all complete

**When to use Sequential:**
- Workflows depend on each other's output
- You need guaranteed execution order
- Example: Fetch PR → Analyze → Review (each step needs previous output)

**When to use Parallel:**
- Workflows are independent (don't depend on each other's output)
- You want faster execution
- Example: Fetch PR metadata and Fetch PR diff concurrently (both use same PR number from input), then combine results

## Composition Patterns

### 1. Agent Definition References in Sequential/Composite Workflows

Reference agent definitions (or workflows) as agents. `kind: Sequential` is an alias for `kind: Composite` with `execution: sequential`:

```yaml
kind: Sequential  # Alias for Composite with execution: sequential
name: ComposedWorkflow

workflow:
  agents:
    # Reference an agent definition (kind: Direct)
    - workflow: ../agents/pr_fetcher.yaml  # This is an agent definition
      name: PRFetcher
    
    # Reference with overrides
    - workflow:
        file: ../agents/pr_reviewer.yaml  # This is also an agent definition
        overrides:
          model:
            provider: DeepSeek
            model_name: deepseek-chat
      name: PRReviewer
```

**Equivalent Composite form:**
```yaml
kind: Composite
name: ComposedWorkflow
execution: sequential  # Explicit sequential mode

workflow:
  agents:
    - workflow: ../agents/pr_fetcher.yaml  # Agent definition
      name: PRFetcher
    - workflow: ../agents/pr_reviewer.yaml  # Agent definition
      name: PRReviewer
```

**How it works:**
- Each agent in the `agents` list can reference an agent definition (`kind: Direct`) or workflow file
- The referenced file is loaded and built into an agent
- Output flows sequentially from one agent to the next
- Overrides allow customizing the referenced component's configuration

### 2. Composite Workflows

Compose multiple workflows together with configurable execution mode:

**Sequential Execution (default):**
```yaml
kind: Composite
name: CompositeWorkflow

# execution: sequential  # This is the default
workflow:
  workflows:
    - file: ../agents/pr_fetcher.yaml
      name: FetchPR
    - file: ../agents/pr_reviewer.yaml
      name: ReviewPR
```

**Parallel Execution:**
```yaml
kind: Composite
name: ParallelWorkflow

execution: parallel
workflow:
  workflows:
    - file: ../agents/pr_fetcher.yaml
      name: FetchPR
    - file: ../agents/pr_reviewer.yaml
      name: ReviewPR
```

**How it works:**
- **Sequential (default)**: Workflows execute one after another, output flows from one to the next
- **Parallel**: Workflows execute concurrently, outputs are combined after all complete
- Useful for composing independent workflows
- Sequential is better when workflows depend on each other's output
- Parallel is better when workflows are independent and you want faster execution

### 3. Workflow References in Direct Workflows

A single agent can reference another agent definition or workflow:

```yaml
kind: Direct
name: SingleAgentWorkflow

agent:
  workflow: ../agents/pr_fetcher.yaml
  name: CustomPRFetcher
```

## Reference Resolution

### Path Resolution

Workflow references support both relative and absolute paths:

```yaml
# Relative path (resolved from the current workflow's directory)
workflow: ../agents/pr_fetcher.yaml

# Relative path with parent directory
workflow: ../shared/workflows/base.yaml

# Absolute path
workflow: /opt/workflows/custom.yaml
```

### Circular Reference Detection

The system automatically detects and prevents circular references:

```yaml
# workflow_a.yaml
kind: Direct
agent:
  workflow: ./workflow_b.yaml  # ❌ Error if workflow_b.yaml references workflow_a.yaml
```

## Overrides

### Workflow-Level Overrides

Override configuration for all agents/components in a workflow:

```yaml
kind: Sequential
name: MyWorkflow

# Workflow-level overrides apply to all agents/components
overrides:
  model:
    provider: DeepSeek        # Override model provider
    model_name: deepseek-chat  # Override model name
    parameters:
      temperature: 0.5        # Override temperature

workflow:
  agents:
    - workflow: ../agents/pr_fetcher.yaml
    - workflow: ../agents/pr_reviewer.yaml
```

**Benefits:**
- Apply the same override to all agents/components at once
- Cleaner YAML structure
- Easier to maintain consistent model configuration across a workflow

### Per-Reference Overrides

Override configuration for a specific workflow reference:

```yaml
workflow:
  agents:
    - workflow:
        file: ../agents/pr_fetcher.yaml
        overrides:
          model:
            provider: DeepSeek
            model_name: deepseek-chat
      name: PRFetcher
    - workflow: ../agents/pr_reviewer.yaml  # Uses default model
```

**Use Cases:**
- Override only specific components
- Mix different models in the same workflow
- Fine-grained control over individual components

**Supported Overrides:**
- `model.provider` - Change the LLM provider
- `model.model_name` - Change the model name
- `model.parameters` - Override model parameters (temperature, max_tokens, etc.)

**Precedence:**
1. Per-reference overrides (highest priority)
2. Workflow-level overrides
3. Component defaults (lowest priority)

## Examples

### Example 1: Reusable Agent Definitions

Create reusable agent definitions (components):

**`agents/pr_fetcher.yaml`** (Agent Definition):
```yaml
kind: Direct  # Single agent definition
name: PRFetcher
agent:  # For Direct kind, agent is at top level (no workflow: wrapper)
  name: PRFetcher
  tools: [fetch_pull_request, get_pull_request_diff]
  # ... agent config
```

**`pr_review_composed.yaml`** (Workflow):
```yaml
kind: Sequential  # Workflow orchestrating multiple agents
name: PRReviewComposed
workflow:
  agents:
    - workflow: ../agents/pr_fetcher.yaml  # References agent definition
    - workflow: ../agents/pr_reviewer.yaml  # References agent definition
```

### Example 2: Workflow-Level Model Override

Apply the same model override to all components:

```yaml
kind: Sequential
name: PRReviewWithDeepSeek

# Override model for all agents/components
overrides:
  model:
    provider: DeepSeek
    model_name: deepseek-chat

workflow:
  agents:
    - workflow: ../agents/pr_fetcher.yaml    # Uses DeepSeek
    - workflow: ../agents/pr_reviewer.yaml  # Uses DeepSeek
```

### Example 3: Per-Reference Override

Override model for specific components:

```yaml
kind: Sequential
workflow:
  agents:
    # Use DeepSeek for fetcher
    - workflow:
        file: ../agents/pr_fetcher.yaml
        overrides:
          model:
            provider: DeepSeek
    # Use OpenAI (default) for reviewer
    - workflow: ../agents/pr_reviewer.yaml
```

### Example 4: Sequential Composite Workflow

Compose multiple workflows that depend on each other:

```yaml
kind: Composite
name: MultiStageReview
execution: sequential  # Explicit sequential mode
workflow:
  workflows:
    - file: ./workflows/fetch_stage.yaml
      name: Fetch
    - file: ./workflows/analyze_stage.yaml
      name: Analyze
    - file: ./workflows/review_stage.yaml
      name: Review
```

### Example 5: Parallel Composite Workflow

Execute independent workflows concurrently. These workflows are independent because they all use the same input (PR number) but fetch different data:

```yaml
kind: Composite
name: ParallelPRDataCollection
execution: parallel
workflow:
  workflows:
    - file: ../agents/pr_metadata_fetcher.yaml
      name: FetchMetadata
    - file: ../agents/pr_diff_fetcher.yaml
      name: FetchDiff
```

Both workflows run concurrently since they're independent - neither depends on the other's output. They both extract the PR number from the input and fetch different data.

### Example 6: Parallel Then Sequential (Two-Stage Workflow)

Combine parallel execution with sequential processing:

```yaml
kind: Sequential
name: ParallelDataThenCombine
workflow:
  agents:
    # Stage 1: Parallel data collection
    - workflow: ./parallel_workflow.yaml  # Contains parallel execution
      name: ParallelDataFetch
    
    # Stage 2: Combine the parallel results sequentially
    - workflow: ../agents/pr_combiner.yaml
      name: CombineResults
```

This shows how to use parallel execution for independent data fetching, then sequential execution to process the combined results.

## Best Practices

### 1. Component Design
- **Single Responsibility**: Each component should do one thing well
- **Reusability**: Design components to be reusable across workflows
- **Clear Interfaces**: Components should have clear input/output expectations

### 2. Workflow Organization
```
workflows/
├── agents/              # Reusable agent definitions
│   ├── pr_fetcher.yaml  # Fetches both PR details and diff
│   └── pr_reviewer.yaml
├── composed/            # Composed workflows
│   ├── pr_review_composed.yaml
│   └── pr_review_with_overrides.yaml
└── standalone/          # Standalone workflows
    └── research_agent.yaml
```

### 3. Naming Conventions
- Use descriptive names: `pr_fetcher.yaml` not `wf1.yaml`
- Include purpose in name: `pr_reviewer.yaml` not `reviewer.yaml`
- Use consistent naming: `pr_*` for PR-related workflows

### 4. Documentation
- Document component purpose in `description` field
- Document expected inputs/outputs in agent `instructions`
- Include usage examples in component comments

## Reference Format

### Simple Reference (String)
```yaml
workflow: "./path/to/workflow.yaml"
```

### Object Reference (with overrides)
```yaml
workflow:
  file: "./path/to/workflow.yaml"
  name: "CustomName"  # Optional: override workflow name
  overrides:
    model:
      provider: DeepSeek
```

## Limitations

1. **Override Depth**: Currently supports model overrides only. Tool and memory overrides coming soon.
2. **Validation**: Override validation is basic - invalid overrides may cause runtime errors
3. **Circular References**: Detected but error messages could be more helpful

## Migration Guide

### From Inline to Composed

**Before (inline):**
```yaml
kind: Sequential
workflow:
  agents:
    - name: PRFetcher
      tools: [fetch_pull_request]
      # ... full agent config
    - name: PRReviewer
      tools: []
      # ... full agent config
```

Note: For `kind: Direct`, use `agent:` at top level (no `workflow:` wrapper).

**After (composed):**
```yaml
kind: Sequential
workflow:
  agents:
    - workflow: ../agents/pr_fetcher.yaml
    - workflow: ../agents/pr_reviewer.yaml
```

Benefits:
- ✅ Reusable components
- ✅ Easier maintenance
- ✅ Consistent behavior
- ✅ Less duplication

