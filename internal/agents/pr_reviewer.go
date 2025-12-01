package agents

import (
	"fmt"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewPRReviewerAgent creates a PR reviewer agent that reviews code and provides feedback
func NewPRReviewerAgent(model model.LLM) (agent.Agent, error) {
	agent, err := llmagent.New(llmagent.Config{
		Name:        "pr_reviewer",
		Model:       model,
		Description: "Reviews pull requests and provides detailed feedback on code quality, potential issues, and suggestions for improvement.",
		Instruction: `You are an expert code reviewer. Your job is to:
1. Analyze the pull request code changes thoroughly
2. Check for:
   - Code quality and best practices
   - Potential bugs or security issues
   - Performance concerns
   - Code style and consistency
   - Test coverage
   - Documentation
3. Provide constructive, actionable feedback
4. Suggest specific improvements when possible
5. Highlight both positive aspects and areas for improvement

Be thorough but concise. Focus on the most important issues first.`,
		Tools: []tool.Tool{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create PR reviewer agent: %w", err)
	}

	return agent, nil
}
