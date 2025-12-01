package main

import (
	"context"
	"flag"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from env file
	if err := godotenv.Load("env"); err != nil {
		log.Printf("Warning: failed to load env file: %v", err)
	}

	ctx := context.Background()

	// Parse command line flags
	workflowFile := flag.String("workflow", "", "Path to workflow YAML file")
	input := flag.String("input", "", "Input text for the workflow")
	flag.Parse()

	// Validate required flags
	if *workflowFile == "" {
		log.Fatal("Error: --workflow flag is required.\n\nUsage:\n  kinetic --workflow <file> --input <input>\n\nExample:\n  kinetic --workflow research.yaml --input \"What is Rust?\"")
	}
	if *input == "" {
		log.Fatal("Error: --input flag is required.\n\nUsage:\n  kinetic --workflow <file> --input <input>")
	}

	// Run workflow
	if err := runWorkflowCommand(ctx, *workflowFile, *input); err != nil {
		log.Fatalf("Workflow execution failed: %v", err)
	}
}
