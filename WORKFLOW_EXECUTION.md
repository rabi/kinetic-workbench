# Workflow Execution Models

## Current Situation: ReAct Executor

With the **ReAct executor** (default), the LLM decides when and in what order to call tools based on:

1. **Instructions**: What you tell it to do
2. **Available Tools**: What tools are available
3. **Conversation Context**: What information is already available
4. **LLM Reasoning**: The model's internal decision-making

### How ReAct Works

```
User Input → LLM → [Decides to call tool?] → Tool Execution → LLM → [More tools?] → Final Response
```

**Key Point**: The LLM makes autonomous decisions. There's **no guarantee** tools will be called in a specific order or at all.

### Example: PR Review with ReAct

```yaml
tools:
  - fetch_pull_request
  - get_pull_request_diff
```

**What happens:**
- LLM sees both tools available
- Instructions say "1. Fetch PR... 2. Analyze diff..."
- **But**: LLM might call `get_pull_request_diff` first if it thinks that's better
- Or: LLM might skip `fetch_pull_request` if PR number is in context
- Or: LLM might call both in parallel (if supported)

**No Guarantees**: The order depends on LLM reasoning, not explicit workflow definition.

## Problem: Sequential Tool Dependencies

Some workflows **require** sequential execution:

```
Step 1: fetch_pull_request(123) → Get PR number, files list
Step 2: get_pull_request_diff(123) → Get actual code changes
Step 3: Analyze and review → Use both results
```

If the LLM calls `get_pull_request_diff` before `fetch_pull_request`, it might work, but:
- We lose the context from PR metadata
- The workflow is less predictable
- Harder to debug when things go wrong

## Solution: Sequential Workflows

We should support **Sequential workflows** in YAML where agents/tools run in explicit order:

```yaml
kind: Sequential

workflow:
  agents:
    - name: PRFetcher
      tools: [fetch_pull_request]
      # Output becomes input for next agent
    - name: PRReviewer  
      tools: [get_pull_request_diff]
      # Receives PR data from previous agent
```

This ensures:
- ✅ Explicit execution order
- ✅ Data flows between agents
- ✅ Predictable behavior
- ✅ Better for complex workflows

## Current Workaround: Explicit Instructions

For now, you can make instructions very explicit:

```yaml
instructions: |
  IMPORTANT: You MUST follow these steps in order:
  
  1. FIRST: Call fetch_pull_request with the PR number from user input
  2. WAIT for the result
  3. THEN: Call get_pull_request_diff with the same PR number
  4. FINALLY: Use both results to provide review
  
  Do NOT call get_pull_request_diff before fetch_pull_request.
  Do NOT skip any steps.
```

**Limitation**: Still relies on LLM following instructions - not guaranteed.

## Recommendation

1. **For simple workflows**: Use ReAct with explicit instructions (current approach)
2. **For complex workflows**: Support Sequential workflow kind in YAML (future enhancement)

Would you like me to implement Sequential workflow support in the YAML system?

