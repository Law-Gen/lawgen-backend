package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
)

type ragClient struct {
	apiURL string
	client *http.Client
}

func NewRAGClient(cfg *config.Config) (domain.RAGService, error) {
	return &ragClient{
		apiURL: cfg.RAGServiceAddr, // e.g. "http://127.0.0.1:8000"
		client: &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (r *ragClient) Close() error {
	// nothing to close for REST
	return nil
}

func (r *ragClient) Retrieve(ctx context.Context, query string, k int) (*domain.RAGResult, error) {
	// Prepare request body
	body, err := json.Marshal(map[string]interface{}{
		"query": query,
		"k":     k,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Send HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.apiURL+"/ask", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("REST call to /ask failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	// Parse response JSON
	var parsed struct {
		Results []struct {
			Content       string `json:"content"`
			Source        string `json:"source"`
			ArticleNumber string `json:"article_number"`
			Topics        string `json:"topics"`
		} `json:"results"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	// Convert to domain
	var sources []domain.RAGSource
	for _, r := range parsed.Results {
		sources = append(sources, domain.RAGSource{
			Content:       r.Content,
			Source:        r.Source,
			ArticleNumber: r.ArticleNumber,
			Topics:        parseTopics(r.Topics),
		})
	}

	return &domain.RAGResult{
		Results:    sources,
		Message:    parsed.Message,
		References: nil, // REST API response doesn’t include references
	}, nil
}

// parseTopics parses the topics string into a slice of strings
func parseTopics(topics string) []string {
	if topics == "" {
		return nil
	}
	var arr []string
	if err := json.Unmarshal([]byte(topics), &arr); err == nil {
		return arr
	}
	parts := strings.Split(topics, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
