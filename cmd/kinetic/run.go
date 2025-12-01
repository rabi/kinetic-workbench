package main

import (
	"context"
	"fmt"
	"strings"

	"kinetic-workbench/pkg/workflow"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

// runWorkflowCommand handles the `kinetic --workflow <file> --input <input>` command
func runWorkflowCommand(ctx context.Context, workflowFile, input string) error {
	// Load workflow definition
	workflowDef, err := workflow.LoadFromFile(workflowFile)
	if err != nil {
		return fmt.Errorf("failed to load workflow file: %w", err)
	}

	// Create workflow builder
	builder, err := workflow.NewBuilder()
	if err != nil {
		return fmt.Errorf("failed to create workflow builder: %w", err)
	}

	// Build agent from workflow definition
	workflowAgent, err := builder.BuildAgent(ctx, workflowDef, workflowFile)
	if err != nil {
		return fmt.Errorf("failed to build agent: %w", err)
	}

	// Print workflow info
	printWorkflowInfo(workflowDef)

	// Create user content
	userContent := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{{Text: input}},
	}

	// Create services
	sessionService := session.InMemoryService()
	memoryService := buildMemoryService(builder.GetMemoryConfig(workflowDef))

	// Run the agent
	return runAgent(ctx, workflowAgent, userContent, sessionService, memoryService)
}

// printWorkflowInfo prints workflow structure information
func printWorkflowInfo(def *workflow.WorkflowDefinition) {
	fmt.Printf("Loaded workflow: %s (kind: %s)\n", def.Name, def.Kind)
	if def.Description != "" {
		fmt.Printf("Description: %s\n", def.Description)
	}

	switch def.Kind {
	case "Composite":
		printCompositeInfo(def)
	case "Direct":
		printDirectInfo(def)
	}
	fmt.Println()
}

// printCompositeInfo prints composite workflow details
func printCompositeInfo(def *workflow.WorkflowDefinition) {
	if len(def.Workflow.Workflows) > 0 {
		fmt.Printf("Composite workflow with %d workflows:\n", len(def.Workflow.Workflows))
		for i, wf := range def.Workflow.Workflows {
			fmt.Printf("  Workflow %d: %s (%s)\n", i+1, wf.Name, wf.File)
		}
		return
	}

	if len(def.Workflow.Agents) > 0 {
		fmt.Printf("Composite workflow with %d agents:\n", len(def.Workflow.Agents))
		for i, ag := range def.Workflow.Agents {
			ref := ag.Workflow
			if ref == nil {
				ref = ag.Defn
			}
			if ref != nil {
				fmt.Printf("  Agent %d: %s (references: %s)\n", i+1, ag.Name, ref.File)
			} else {
				fmt.Printf("  Agent %d: %s\n", i+1, ag.Name)
			}
		}
		if def.Workflow.Execution != "" {
			fmt.Printf("  Execution mode: %s\n", def.Workflow.Execution)
		}
	}
}

// printDirectInfo prints direct workflow details
func printDirectInfo(def *workflow.WorkflowDefinition) {
	ag := def.Agent
	if ag == nil {
		ag = def.Workflow.Agent
	}
	if ag == nil {
		return
	}

	if ag.Workflow != nil {
		fmt.Printf("Agent: %s (references workflow: %s)\n", ag.Name, ag.Workflow.File)
	} else {
		fmt.Printf("Agent: %s\n", ag.Name)
		if ag.Model.Provider != "" || ag.Model.ModelName != "" {
			fmt.Printf("Model: %s (%s)\n", ag.Model.Provider, ag.Model.ModelName)
		}
		if len(ag.Tools) > 0 {
			fmt.Printf("Tools: %v\n", ag.Tools)
		}
	}
}

// buildMemoryService creates a memory service based on configuration
func buildMemoryService(kind string, _ map[string]interface{}) memory.Service {
	if kind == "" {
		return nil
	}
	// TODO: Implement sliding window memory that respects window_size parameter
	return memory.InMemoryService()
}

// runAgent runs an agent with user input
func runAgent(ctx context.Context, ag agent.Agent, userContent *genai.Content, sessionService session.Service, memoryService memory.Service) error {
	// Create session
	createResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "kinetic",
		UserID:  "default-user",
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create runner
	runnerConfig := runner.Config{
		AppName:        "kinetic",
		SessionService: sessionService,
		Agent:          ag,
	}
	if memoryService != nil {
		runnerConfig.MemoryService = memoryService
	}

	r, err := runner.New(runnerConfig)
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}

	// Run the agent
	fmt.Printf("Running workflow with input: %s\n\n", userContent.Parts[0].Text)

	for event, err := range r.Run(ctx, "default-user", createResp.Session.ID(), userContent, agent.RunConfig{
		StreamingMode: agent.StreamingModeNone,
	}) {
		if err != nil {
			return formatAPIError(err)
		}
		if event != nil {
			printEvent(event)
		}
	}

	fmt.Println()
	return nil
}

// formatAPIError provides helpful error messages for common API errors
func formatAPIError(err error) error {
	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "503") || strings.Contains(errMsg, "overloaded") || strings.Contains(errMsg, "unavailable") {
		return fmt.Errorf("model API temporarily unavailable (503): %w\n\n"+
			"Suggestions:\n"+
			"  1. Wait a few moments and try again\n"+
			"  2. Switch to a different model provider\n"+
			"  3. Try a different model", err)
	}
	return fmt.Errorf("agent execution error: %w", err)
}

// printEvent prints an event's content
func printEvent(event *session.Event) {
	if event.Content == nil {
		return
	}

	for _, part := range event.Content.Parts {
		if part.FunctionCall != nil {
			printFunctionCall(part.FunctionCall)
		}
		if part.FunctionResponse != nil {
			printFunctionResponse(part.FunctionResponse)
		}
		if part.Text != "" {
			fmt.Print(part.Text)
		}
	}
}

// printFunctionCall prints a function call
func printFunctionCall(fc *genai.FunctionCall) {
	fmt.Printf("\n[Tool Call] %s(", fc.Name)
	if fc.Args != nil {
		args := make([]string, 0, len(fc.Args))
		for k, v := range fc.Args {
			args = append(args, fmt.Sprintf("%s=%v", k, v))
		}
		fmt.Print(strings.Join(args, ", "))
	}
	fmt.Println(")")
}

// printFunctionResponse prints a function response
func printFunctionResponse(fr *genai.FunctionResponse) {
	fmt.Printf("[Tool Complete] %s() finished\n", fr.Name)
	if fr.Response != nil {
		if _, hasError := fr.Response["error"]; hasError {
			fmt.Printf("  ⚠️  Error: %v\n", fr.Response["error"])
		} else {
			fmt.Printf("  ✓ Success\n")
		}
	}
}
