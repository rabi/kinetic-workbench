package agents

import (
	"fmt"

	"kinetic/internal/tools/github"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// NewPRFetcherAgent creates a PR fetcher agent that fetches PR details from GitHub
func NewPRFetcherAgent(model model.LLM, githubTool *github.Tool) (agent.Agent, error) {
	// Create GitHub tools
	githubTools, err := github.CreateTools(githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create GitHub tools: %w", err)
	}

	// Debug: Log tool names
	for _, t := range githubTools {
		fmt.Printf("DEBUG: Created tool - Name: %s, Type: %T\n", t.Name(), t)
	}

	// PR Fetcher Agent - fetches PR details
	// functiontool handles execution automatically - no need for BeforeToolCallbacks
	agent, err := llmagent.New(llmagent.Config{
		Name:        "pr_fetcher",
		Model:       model,
		Description: "Fetches pull request information from GitHub including PR details, files changed, and code diffs.",
		Instruction: `You are a GitHub PR fetcher agent. Your job is to:
1. Fetch pull request details when given a PR number
2. Retrieve the list of files changed in the PR
3. Get the code diff for the PR
4. Summarize the PR changes clearly

When asked to fetch a PR, use the available GitHub tools to get all relevant information.`,
		Tools: githubTools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create PR fetcher agent: %w", err)
	}

	return agent, nil
}
