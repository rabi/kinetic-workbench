package providers

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// CreateModel creates the appropriate model based on environment configuration
func CreateModel(ctx context.Context) (model.LLM, error) {
	modelProvider := os.Getenv("MODEL_PROVIDER")
	if modelProvider == "" {
		modelProvider = "deepseek"
	}

	switch modelProvider {
	case "gemini", "google":
		apiKey := os.Getenv("GOOGLE_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("GOOGLE_API_KEY environment variable is required")
		}
		modelName := os.Getenv("GEMINI_MODEL")
		if modelName == "" {
			modelName = "gemini-3-pro-preview"
		}
		return gemini.NewModel(ctx, modelName, &genai.ClientConfig{
			APIKey: apiKey,
		})

	case "deepseek":
		apiKey := os.Getenv("DEEPSEEK_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
		}
		modelName := os.Getenv("DEEPSEEK_MODEL")
		if modelName == "" {
			modelName = "deepseek-chat"
		}
		return NewDeepSeekModel(apiKey, modelName)

	default:
		return nil, fmt.Errorf("unsupported model provider: %s", modelProvider)
	}
}
