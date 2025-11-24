package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OllamaEmbeddingService handles embedding generation via Ollama
type OllamaEmbeddingService struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllamaEmbeddingService creates a new Ollama embedding service
func NewOllamaEmbeddingService(baseURL, model string) *OllamaEmbeddingService {
	// Normalize base URL: remove trailing slash
	b := baseURL
	if len(b) > 0 && b[len(b)-1] == '/' {
		b = b[:len(b)-1]
	}
	return &OllamaEmbeddingService{
		baseURL: b,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// EmbeddingRequest represents an embedding request
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents an embedding response
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

// GenerateEmbedding generates an embedding for the given text
func (s *OllamaEmbeddingService) GenerateEmbedding(text string) ([]float32, error) {
	reqBody := EmbeddingRequest{
		Model:  s.model,
		Prompt: text,
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", s.baseURL+"/api/embeddings", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var embedResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return embedResp.Embedding, nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (s *OllamaEmbeddingService) GenerateEmbeddings(texts []string) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(texts))
	for _, text := range texts {
		embedding, err := s.GenerateEmbedding(text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text: %w", err)
		}
		embeddings = append(embeddings, embedding)
	}
	return embeddings, nil
}
