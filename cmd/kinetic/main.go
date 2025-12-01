package main

import (
	"context"
	"log"
	"os"

	"kinetic/internal/agents"
	"kinetic/internal/providers"
	"kinetic/internal/tools/github"

	"github.com/joho/godotenv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/session"
)

func main() {
	// Load environment variables from env file
	if err := godotenv.Load("env"); err != nil {
		log.Printf("Warning: failed to load env file: %v", err)
	}

	ctx := context.Background()

	// Create model
	model, err := providers.CreateModel(ctx)
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// Load GitHub tool
	githubTool, err := github.LoadFromEnv()
	if err != nil {
		log.Fatalf("Failed to load GitHub tool: %v", err)
	}

	// Create router agent that routes to review or cherry-pick workflows based on keywords
	workflowAgent, err := agents.NewRouterAgent(model, githubTool)
	if err != nil {
		log.Fatalf("Failed to create router agent: %v", err)
	}

	log.Println("PR Review Agent initialized with LLM-based router:")
	log.Println("  - Router Agent (uses LLM to intelligently route requests)")
	log.Println("  - Review Workflow (for PR review requests)")
	log.Println("    - PR Fetcher Agent (fetches PR details from GitHub)")
	log.Println("    - PR Reviewer Agent (reviews code and provides feedback)")
	log.Println("  - Cherry-Pick Workflow (for cherry-pick/backport requests)")
	log.Println("    - Cherry-Pick Agent (finds merged PRs and creates cherry-pick PRs)")

	config := &launcher.Config{
		SessionService: session.InMemoryService(),
		AgentLoader:    agent.NewSingleLoader(workflowAgent),
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}

