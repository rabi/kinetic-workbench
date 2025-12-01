package agents

import (
	"context"
	"fmt"
	"iter"
	"strings"
	"time"

	"kinetic/internal/tools/github"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/genai"
)

// NewRouterAgent creates a router agent that uses LLM to intelligently route to review or cherry-pick workflows
func NewRouterAgent(model model.LLM, githubTool *github.Tool) (agent.Agent, error) {
	// Create review workflow
	reviewAgent, err := CreatePRWorkflow(model, githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create review workflow: %w", err)
	}

	// Create cherry-pick workflow
	cherryPickAgent, err := CreateCherryPickWorkflow(model, githubTool)
	if err != nil {
		return nil, fmt.Errorf("failed to create cherry-pick workflow: %w", err)
	}

	// Create router LLM agent that analyzes user intent
	routerLLM, err := llmagent.New(llmagent.Config{
		Name:        "router_llm",
		Model:       model,
		Description: "Analyzes user requests to determine which workflow to use.",
		Instruction: `You are a router agent. Your job is to analyze user requests and determine which workflow should handle the request.

Available workflows:
1. **REVIEW**: Use for requests about reviewing, analyzing, or examining pull requests
   - Examples: "review PR 1060", "analyze this PR", "check the code changes", "what's in PR 123"
   - Keywords: review, analyze, check, examine, look at, what's in, feedback

2. **CHERRY_PICK**: Use for requests about finding merged PRs, creating cherry-picks, or backporting changes
   - Examples: "find merged PRs from last week", "create cherry-picks", "backport PR 1060", "what PRs were merged recently"
   - Keywords: cherry-pick, cherrypick, backport, merged, last week, recent merges

Respond with ONLY one word: either "REVIEW" or "CHERRY_PICK" (all caps, no punctuation).
Do not include any other text or explanation.`,
		Tools: []tool.Tool{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create router LLM agent: %w", err)
	}

	// Create router agent using agent.New that uses LLM to decide routing
	return agent.New(agent.Config{
		Name:        "router",
		Description: "Routes user requests to review or cherry-pick workflows using LLM-based intent detection.",
		Run: func(ctx agent.InvocationContext) iter.Seq2[*session.Event, error] {
			return func(yield func(*session.Event, error) bool) {
				// Get the user input from the context
				userContent := ctx.UserContent()
				if userContent == nil {
					yield(nil, fmt.Errorf("no user content provided"))
					return
				}

				// Extract text from user content
				var userMessage strings.Builder
				for _, part := range userContent.Parts {
					if part.Text != "" {
						userMessage.WriteString(part.Text)
						userMessage.WriteString(" ")
					}
				}
				messageText := strings.TrimSpace(userMessage.String())
				if messageText == "" {
					yield(nil, fmt.Errorf("no user message text found"))
					return
				}

				// Use LLM router to determine which workflow to use
				routerCtx := &routerInvocationContext{ctx: ctx}
				var routerDecision string
				for event, err := range routerLLM.Run(routerCtx) {
					if err != nil {
						yield(nil, fmt.Errorf("router LLM failed: %w", err))
						return
					}
					if event == nil {
						continue
					}
					// Extract text from the router's response
					if event.Content != nil {
						for _, part := range event.Content.Parts {
							if part.Text != "" {
								routerDecision += part.Text
							}
						}
					}
				}

				// Parse LLM decision (should be "REVIEW" or "CHERRY_PICK")
				routerDecision = strings.TrimSpace(strings.ToUpper(routerDecision))
				isCherryPick := strings.Contains(routerDecision, "CHERRY_PICK") || strings.Contains(routerDecision, "CHERRYPICK")

				// Select the appropriate agent based on LLM decision
				selectedAgent := reviewAgent
				if isCherryPick {
					selectedAgent = cherryPickAgent
				}

				// Run the selected agent
				for event, err := range selectedAgent.Run(ctx) {
					if !yield(event, err) {
						return
					}
					if err != nil {
						return
					}
				}
			}
		},
		SubAgents: []agent.Agent{reviewAgent, cherryPickAgent},
	})
}

// routerInvocationContext wraps agent.InvocationContext to pass user content to router LLM
type routerInvocationContext struct {
	ctx agent.InvocationContext
}

func (r *routerInvocationContext) Agent() agent.Agent {
	return r.ctx.Agent()
}

func (r *routerInvocationContext) Artifacts() agent.Artifacts {
	return r.ctx.Artifacts()
}

func (r *routerInvocationContext) Memory() agent.Memory {
	return r.ctx.Memory()
}

func (r *routerInvocationContext) Session() session.Session {
	return r.ctx.Session()
}

func (r *routerInvocationContext) InvocationID() string {
	return r.ctx.InvocationID()
}

func (r *routerInvocationContext) Branch() string {
	return r.ctx.Branch()
}

func (r *routerInvocationContext) UserContent() *genai.Content {
	// Return the user content from the original context
	return r.ctx.UserContent()
}

func (r *routerInvocationContext) Deadline() (time.Time, bool) {
	return r.ctx.Deadline()
}

func (r *routerInvocationContext) Done() <-chan struct{} {
	if ctx, ok := r.ctx.(context.Context); ok {
		return ctx.Done()
	}
	return nil
}

func (r *routerInvocationContext) Err() error {
	if ctx, ok := r.ctx.(context.Context); ok {
		return ctx.Err()
	}
	return nil
}

func (r *routerInvocationContext) Value(key interface{}) interface{} {
	if ctx, ok := r.ctx.(context.Context); ok {
		return ctx.Value(key)
	}
	return nil
}

func (r *routerInvocationContext) RunConfig() *agent.RunConfig {
	return r.ctx.RunConfig()
}

func (r *routerInvocationContext) EndInvocation() {
	r.ctx.EndInvocation()
}

func (r *routerInvocationContext) Ended() bool {
	return r.ctx.Ended()
}
