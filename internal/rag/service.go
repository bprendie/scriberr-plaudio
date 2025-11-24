package rag

import (
	"context"
	"fmt"
	"strings"

	"scriberr/internal/database"
	"scriberr/internal/embeddings"
	"scriberr/internal/llm"
	"scriberr/internal/models"
	"scriberr/internal/vectordb"
)

// LLMService interface for RAG service
type LLMService interface {
	ChatCompletion(ctx context.Context, model string, messages []llm.ChatMessage, temperature float64) (*llm.ChatResponse, error)
}

// RAGService handles RAG operations
type RAGService struct {
	vectorDB   *vectordb.ChromaDBClient
	embedding  *embeddings.OllamaEmbeddingService
	llmService LLMService
	collectionName string
}

// NewRAGService creates a new RAG service
func NewRAGService(vectorDB *vectordb.ChromaDBClient, embedding *embeddings.OllamaEmbeddingService, llmService LLMService) *RAGService {
	service := &RAGService{
		vectorDB:      vectorDB,
		embedding:     embedding,
		llmService:    llmService,
		collectionName: "transcriptions",
	}
	
	// Ensure collection exists
	_ = service.vectorDB.CreateCollection(service.collectionName, map[string]interface{}{
		"description": "Transcription summaries and content",
	})
	
	return service
}

// StoreSummary stores a summary in the vector database
func (s *RAGService) StoreSummary(transcriptionID, summary, transcript string) error {
	// Combine summary and transcript for better context
	// If summary is empty, just use transcript
	var content string
	if summary != "" {
		content = fmt.Sprintf("Summary: %s\n\nTranscript: %s", summary, transcript)
	} else {
		content = fmt.Sprintf("Transcript: %s", transcript)
	}
	
	// Generate embedding
	embedding, err := s.embedding.GenerateEmbedding(content)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}
	
	// Store in vector DB
	metadata := map[string]interface{}{
		"transcription_id": transcriptionID,
		"type":            "summary",
	}
	
	err = s.vectorDB.AddDocuments(
		s.collectionName,
		[]string{transcriptionID},
		[]string{content},
		[][]float32{embedding},
		[]map[string]interface{}{metadata},
	)
	
	if err != nil {
		return fmt.Errorf("failed to store in vector DB: %w", err)
	}
	
	return nil
}

// Query performs a RAG query
func (s *RAGService) Query(ctx context.Context, query string, nResults int) ([]string, error) {
	if nResults == 0 {
		nResults = 5
	}
	
	// Generate embedding for query
	queryEmbedding, err := s.embedding.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}
	
	// Query vector DB
	results, err := s.vectorDB.Query(s.collectionName, [][]float32{queryEmbedding}, nResults, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query vector DB: %w", err)
	}
	
	if len(results.Documents) == 0 || len(results.Documents[0]) == 0 {
		return []string{}, nil
	}
	
	return results.Documents[0], nil
}

// Chat performs a RAG-enhanced chat
func (s *RAGService) Chat(ctx context.Context, query string, model string, temperature float64) (string, error) {
	// Query relevant context
	contexts, err := s.Query(ctx, query, 5)
	if err != nil {
		return "", fmt.Errorf("failed to query context: %w", err)
	}
	
	// Build prompt with context
	var prompt strings.Builder
	prompt.WriteString("You are a helpful assistant that answers questions based on the following transcription summaries and transcripts.\n\n")
	prompt.WriteString("Relevant context:\n")
	for i, ctx := range contexts {
		prompt.WriteString(fmt.Sprintf("%d. %s\n\n", i+1, ctx))
	}
	prompt.WriteString("\nUser question: ")
	prompt.WriteString(query)
	prompt.WriteString("\n\nPlease provide a helpful answer based on the context above.")
	
	// Call LLM
	messages := []llm.ChatMessage{
		{Role: "user", Content: prompt.String()},
	}
	
	response, err := s.llmService.ChatCompletion(ctx, model, messages, temperature)
	if err != nil {
		return "", fmt.Errorf("failed to get LLM response: %w", err)
	}
	
	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}
	
	return response.Choices[0].Message.Content, nil
}

// GetStats returns statistics about the RAG system
func (s *RAGService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	
	// Count completed transcriptions (each one should be in RAG)
	// This is more reliable than querying ChromaDB directly
	var count int64
	if err := database.DB.Model(&models.TranscriptionJob{}).
		Where("status = ?", models.StatusCompleted).
		Where("transcript IS NOT NULL AND transcript != ''").
		Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to count transcriptions: %w", err)
	}
	
	stats["transcript_count"] = int(count)
	stats["collection_name"] = s.collectionName
	stats["status"] = "active"
	
	return stats, nil
}
