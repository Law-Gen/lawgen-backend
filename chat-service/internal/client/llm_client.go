package client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
)

type llmClient struct {
	client *genai.GenerativeModel
	conn   *genai.Client
	cfg    *config.Config // Store config to access prompts
}

func NewLLMClient(cfg *config.Config) (domain.LLMService, error) {
	ctx := context.Background()
	conn, err := genai.NewClient(ctx, option.WithAPIKey(cfg.GoogleAPIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}
	client := conn.GenerativeModel(cfg.GeminiModel)
	return &llmClient{client: client, conn: conn, cfg: cfg}, nil
}

func (c *llmClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *llmClient) StreamGenerate(ctx context.Context, prompt string, history []domain.ChatEntry, maxWords int) (<-chan domain.LLMStreamResponse, error) {
	cs := c.client.StartChat()

	// Add history to the chat session
	for _, entry := range history {
		role := "user"
		if entry.Type == domain.MessageTypeLLM {
			role = "model"
		}
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{genai.Text(entry.Content)},
			Role:  role,
		})
	}

	iter := cs.SendMessageStream(ctx, genai.Text(prompt))

	resChan := make(chan domain.LLMStreamResponse)

	go func() {
		defer close(resChan)
		for {
			resp, err := iter.Next()
			if err == iterator.Done {
				resChan <- domain.LLMStreamResponse{Done: true}
				return
			}
			if err != nil {
				log.Printf("LLM Stream error: %v", err)
				resChan <- domain.LLMStreamResponse{Error: fmt.Errorf("LLM stream error: %w", err)}
				return
			}

			if resp != nil && len(resp.Candidates) > 0 {
				for _, part := range resp.Candidates[0].Content.Parts {
					if text, ok := part.(genai.Text); ok {
						// Stream word by word for improved readability
						words := strings.Fields(string(text))
						for _, word := range words {
							time.Sleep(30 * time.Millisecond)
							resChan <- domain.LLMStreamResponse{Chunk: word + " "}
						}
					}
				}
			}
		}
	}()

	return resChan, nil
}

func (c *llmClient) Generate(ctx context.Context, prompt string, history []domain.ChatEntry) (string, error) {
	cs := c.client.StartChat()
	for _, entry := range history {
		role := "user"
		if entry.Type == domain.MessageTypeLLM {
			role = "model"
		}
		cs.History = append(cs.History, &genai.Content{
			Parts: []genai.Part{genai.Text(entry.Content)},
			Role:  role,
		})
	}

	resp, err := cs.SendMessage(ctx, genai.Text(prompt))
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}
	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		var sb strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if t, ok := part.(genai.Text); ok {
				sb.WriteString(string(t))
			}
		}
		return sb.String(), nil
	}
	return "", fmt.Errorf("no text generated from LLM for prompt: %s", prompt)
}

func (c *llmClient) Translate(ctx context.Context, text, targetLang string) (string, error) {
	prompt := strings.ReplaceAll(c.cfg.LLMPromptConverter, "{{.Text}}", text)
	// Translation typically doesn't need prior chat history for context
	return c.Generate(ctx, prompt, nil)
}
