package agents

import (
	"fmt"

	"kinetic/internal/tools/github"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/workflowagents/sequentialagent"
	"google.golang.org/adk/model"
)

// CreatePRWorkflow creates a workflow with PR fetcher and reviewer agents
func CreatePRWorkflow(model model.LLM, githubTool *github.Tool) (agent.Agent, error) {
	// Create PR Fetcher Agent
	prFetcherAgent, err := NewPRFetcherAgent(model, githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR fetcher agent: %w", err)
	}

	// Create PR Reviewer Agent
	prReviewerAgent, err := NewPRReviewerAgent(model)
	if err != nil {
		return nil, fmt.Errorf("failed to create PR reviewer agent: %w", err)
	}

	// Create sequential workflow: fetch PR -> review PR
	workflow, err := sequentialagent.New(sequentialagent.Config{
		AgentConfig: agent.Config{
			Name:      "pr_review_workflow",
			SubAgents: []agent.Agent{prFetcherAgent, prReviewerAgent},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow, nil
}

// CreateCherryPickWorkflow creates a workflow with just the cherry-pick agent
func CreateCherryPickWorkflow(model model.LLM, githubTool *github.Tool) (agent.Agent, error) {
	cherryPickAgent, err := NewCherryPickAgent(model, githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create cherry-pick agent: %w", err)
	}

	return cherryPickAgent, nil
}
