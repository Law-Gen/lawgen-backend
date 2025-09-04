package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/LAWGEN/lawgen-backend/chat-service/internal/pb"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
)

type ragClient struct {
	client pb.LegalAssistantClient
	conn   *grpc.ClientConn
}

func NewRAGClient(cfg *config.Config) (domain.RAGService, error) {
	conn, err := grpc.Dial(cfg.RAGServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial RAG gRPC service at %s: %w", cfg.RAGServiceAddr, err)
	}
	client := pb.NewLegalAssistantClient(conn)
	return &ragClient{client: client, conn: conn}, nil
}

func (r *ragClient) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *ragClient) Retrieve(ctx context.Context, query string, k int) (*domain.RAGResult, error) {
	req := &pb.QuestionRequest{
		Query: query,
		K:     int32(k),
	}
	stream, err := r.client.AskQuestion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("gRPC call to AskQuestion failed: %w", err)
	}

	var (
		fullText      string
		allReferences []string
	)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("gRPC stream Recv failed: %w", err)
		}
		// Accumulate text and references from each streamed response
		fullText = resp.Text // Only the last one will be used (final result)
		allReferences = resp.References
	}

	// Parse the JSON in fullText to extract results and message
	// Example: { "results": [...], "message": "..." }
	var parsed struct {
		Results []struct {
			Content  string            `json:"content"`
			Metadata map[string]string `json:"metadata"`
		} `json:"results"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(fullText), &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse RAG response JSON: %w", err)
	}

	var sources []domain.RAGSource
	for _, r := range parsed.Results {
		sources = append(sources, domain.RAGSource{
			Content:       r.Content,
			Source:        r.Metadata["source"],
			ArticleNumber: r.Metadata["article_number"],
			Topics:        parseTopics(r.Metadata["topics"]),
		})
	}

	return &domain.RAGResult{
		Results:    sources,
		Message:    parsed.Message,
		References: allReferences,
	}, nil
}

// parseTopics parses the topics string into a slice of strings
func parseTopics(topics string) []string {
	if topics == "" {
		return nil
	}
	// Try to parse as JSON array
	var arr []string
	if err := json.Unmarshal([]byte(topics), &arr); err == nil {
		return arr
	}
	// Otherwise, split by comma
	parts := strings.Split(topics, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
