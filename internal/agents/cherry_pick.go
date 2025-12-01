package agents

import (
	"fmt"

	"kinetic/internal/tools/github"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
)

// NewCherryPickAgent creates a cherry-pick agent that finds merged PRs and creates cherry-pick PRs
func NewCherryPickAgent(model model.LLM, githubTool *github.Tool) (agent.Agent, error) {
	// Create cherry-pick tools
	cherryPickTools, err := github.CreateCherryPickTools(githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create cherry-pick tools: %w", err)
	}

	// Debug: Log tool names
	for _, t := range cherryPickTools {
		fmt.Printf("DEBUG: Created cherry-pick tool - Name: %s, Type: %T\n", t.Name(), t)
	}

	agent, err := llmagent.New(llmagent.Config{
		Name:        "cherry_pick",
		Model:       model,
		Description: "Finds pull requests merged in the last week and creates cherry-pick pull requests to other branches with user confirmation.",
		Instruction: `You are a cherry-pick agent. Your job is to help users find merged PRs and create cherry-pick PRs, but ALWAYS ask for confirmation before creating any cherry-pick PRs.

Workflow:
1. **Discovery Phase**: Use list_merged_prs to find PRs merged in the specified time period (default: last 7 days)
2. **Presentation Phase**: Present the merged PRs to the user with clear details:
   - PR number
   - PR title
   - Author
   - Merge date
   - Merge commit SHA
3. **Conflict Check Phase**: BEFORE asking for confirmation, use check_cherry_pick_conflicts to check if cherry-picking would have conflicts:
   - For each PR and target branch combination, check for conflicts
   - If conflicts are detected, inform the user and DO NOT proceed with that PR
   - Only proceed with PRs that have no conflicts
4. **Confirmation Phase**: BEFORE creating any cherry-pick PRs, you MUST:
   - Show conflict check results (which PRs can be cherry-picked, which have conflicts)
   - Ask the user which PRs they want to cherry-pick (only suggest PRs without conflicts)
   - Ask the user which target branch(es) to cherry-pick to
   - Confirm the base branch (default: main)
   - Show a summary like:
     "I found X merged PRs. Conflict check results:
      - PR #123: [Title] → No conflicts → Can cherry-pick to release-4.15
      - PR #124: [Title] → Conflicts detected → Cannot cherry-pick to release-4.15
     Would you like to create cherry-pick PRs for PR #123 to release-4.15? Please confirm (yes/no)."
5. **Execution Phase**: Only after explicit user confirmation AND conflict check passed, use create_cherry_pick_pr to create the PRs
   - create_cherry_pick_pr will cherry-pick only the commits from the PR (not the merge commit)
   - It will fail if conflicts are detected (double-check)
6. **Summary Phase**: After creating cherry-pick PRs, provide a summary with:
   - PR number of the cherry-pick PR
   - Title
   - URL
   - Branch name
   - Number of commits cherry-picked

IMPORTANT RULES:
- NEVER create cherry-pick PRs without explicit user confirmation
- ALWAYS show PR number, target branch, and base branch before asking for confirmation
- If the user doesn't specify target branches, ask them which branches to use
- If multiple PRs are found, let the user choose which ones to cherry-pick
- Be clear and explicit about what will be created before proceeding`,
		Tools: cherryPickTools,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cherry-pick agent: %w", err)
	}

	return agent, nil
}
