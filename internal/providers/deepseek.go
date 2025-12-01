package providers

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"github.com/sashabaranov/go-openai"
	"google.golang.org/adk/model"
	"google.golang.org/genai"
)

// DeepSeekModel implements the model.LLM interface for DeepSeek using OpenAI-compatible API
type DeepSeekModel struct {
	client *openai.Client
	model  string
	name   string
}

// NewDeepSeekModel creates a new DeepSeek model instance using OpenAI-compatible API
func NewDeepSeekModel(apiKey, modelName string) (*DeepSeekModel, error) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com"

	client := openai.NewClientWithConfig(config)

	if modelName == "" {
		modelName = "deepseek-chat"
	}

	return &DeepSeekModel{
		client: client,
		model:  modelName,
		name:   fmt.Sprintf("deepseek-%s", modelName),
	}, nil
}

// Name returns the model name
func (m *DeepSeekModel) Name() string {
	return m.name
}

// GenerateContent implements the model.LLM interface
func (m *DeepSeekModel) GenerateContent(ctx context.Context, req *model.LLMRequest, stream bool) iter.Seq2[*model.LLMResponse, error] {
	return func(yield func(*model.LLMResponse, error) bool) {
		// Convert genai.Content to OpenAI messages
		messages := m.convertContentsToMessages(req.Contents)

		// Create OpenAI request
		chatReq := openai.ChatCompletionRequest{
			Model:    m.model,
			Messages: messages,
			Stream:   stream,
		}

		if stream {
			// Handle streaming
			streamResp, err := m.client.CreateChatCompletionStream(ctx, chatReq)
			if err != nil {
				yield(nil, err)
				return
			}
			defer streamResp.Close()

			for {
				chunk, err := streamResp.Recv()
				if err != nil {
					// Check if it's EOF (end of stream)
					if err.Error() == "EOF" || strings.Contains(err.Error(), "stream closed") {
						return
					}
					yield(nil, err)
					return
				}

				response := m.convertStreamChunkToResponse(chunk)
				if !yield(response, nil) {
					return
				}

				// Check if stream is complete
				if len(chunk.Choices) > 0 && chunk.Choices[0].FinishReason != "" {
					return
				}
			}
		} else {
			// Handle non-streaming
			chatResp, err := m.client.CreateChatCompletion(ctx, chatReq)
			if err != nil {
				yield(nil, err)
				return
			}

			if len(chatResp.Choices) == 0 {
				yield(nil, fmt.Errorf("no choices in response"))
				return
			}

			response := m.convertChatResponseToResponse(chatResp)
			yield(response, nil)
		}
	}
}

// convertContentsToMessages converts genai.Content to OpenAI messages
func (m *DeepSeekModel) convertContentsToMessages(contents []*genai.Content) []openai.ChatCompletionMessage {
	messages := make([]openai.ChatCompletionMessage, 0, len(contents))

	for _, content := range contents {
		if content == nil {
			continue
		}

		role := openai.ChatMessageRoleUser
		if content.Role == "model" || content.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		} else if content.Role == "system" {
			role = openai.ChatMessageRoleSystem
		}

		// Convert parts to text
		var textParts []string
		for _, part := range content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}

		if len(textParts) > 0 {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    role,
				Content: strings.Join(textParts, "\n"),
			})
		}
	}

	return messages
}

// convertChatResponseToResponse converts OpenAI chat response to model.LLMResponse
func (m *DeepSeekModel) convertChatResponseToResponse(resp openai.ChatCompletionResponse) *model.LLMResponse {
	if len(resp.Choices) == 0 {
		return &model.LLMResponse{
			ErrorMessage: "no choices in response",
		}
	}

	choice := resp.Choices[0]
	content := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{Text: choice.Message.Content},
		},
	}

	response := &model.LLMResponse{
		Content:      content,
		Partial:      false,
		TurnComplete: true,
	}

	// Set finish reason if available
	switch choice.FinishReason {
	case openai.FinishReasonStop:
		response.FinishReason = genai.FinishReasonStop
	case openai.FinishReasonLength:
		response.FinishReason = genai.FinishReasonMaxTokens
	default:
		response.FinishReason = genai.FinishReasonOther
	}

	// Set usage metadata if available
	if resp.Usage.TotalTokens > 0 {
		response.UsageMetadata = &genai.GenerateContentResponseUsageMetadata{
			PromptTokenCount:     int32(resp.Usage.PromptTokens),
			CandidatesTokenCount: int32(resp.Usage.CompletionTokens),
			TotalTokenCount:      int32(resp.Usage.TotalTokens),
		}
	}

	return response
}

// convertStreamChunkToResponse converts OpenAI stream chunk to model.LLMResponse
func (m *DeepSeekModel) convertStreamChunkToResponse(chunk openai.ChatCompletionStreamResponse) *model.LLMResponse {
	if len(chunk.Choices) == 0 {
		return &model.LLMResponse{
			Partial: true,
		}
	}

	choice := chunk.Choices[0]
	delta := choice.Delta

	content := &genai.Content{
		Role: "model",
		Parts: []*genai.Part{
			{Text: delta.Content},
		},
	}

	response := &model.LLMResponse{
		Content:      content,
		Partial:      true,
		TurnComplete: choice.FinishReason != "",
	}

	if choice.FinishReason != "" {
		switch choice.FinishReason {
		case openai.FinishReasonStop:
			response.FinishReason = genai.FinishReasonStop
		case openai.FinishReasonLength:
			response.FinishReason = genai.FinishReasonMaxTokens
		default:
			response.FinishReason = genai.FinishReasonOther
		}
	}

	return response
}
