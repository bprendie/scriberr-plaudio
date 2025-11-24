package vectordb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ChromaDBClient handles interactions with ChromaDB
type ChromaDBClient struct {
	baseURL string
	client  *http.Client
}

// NewChromaDBClient creates a new ChromaDB client
func NewChromaDBClient(baseURL string) *ChromaDBClient {
	// Normalize base URL: remove trailing slash
	b := baseURL
	if len(b) > 0 && b[len(b)-1] == '/' {
		b = b[:len(b)-1]
	}
	return &ChromaDBClient{
		baseURL: b,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// CollectionRequest represents a request to create/get a collection
type CollectionRequest struct {
	Name      string                 `json:"name"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	GetOrCreate bool                 `json:"get_or_create,omitempty"`
}

// AddRequest represents a request to add documents
type AddRequest struct {
	CollectionName string   `json:"collection_name"`
	IDs           []string  `json:"ids"`
	Documents     []string  `json:"documents"`
	Embeddings    [][]float32 `json:"embeddings"`
	Metadatas     []map[string]interface{} `json:"metadatas,omitempty"`
}

// QueryRequest represents a query request
type QueryRequest struct {
	CollectionName string      `json:"collection_name"`
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults       int         `json:"n_results"`
	Where          map[string]interface{} `json:"where,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	IDs       [][]string `json:"ids"`
	Documents [][]string `json:"documents"`
	Distances [][]float32 `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
}

// CreateCollection creates or gets a collection
func (c *ChromaDBClient) CreateCollection(name string, metadata map[string]interface{}) error {
	reqBody := CollectionRequest{
		Name:        name,
		Metadata:    metadata,
		GetOrCreate: true,
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/collections", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// AddDocuments adds documents with embeddings to a collection
func (c *ChromaDBClient) AddDocuments(collectionName string, ids []string, documents []string, embeddings [][]float32, metadatas []map[string]interface{}) error {
	reqBody := AddRequest{
		CollectionName: collectionName,
		IDs:           ids,
		Documents:     documents,
		Embeddings:    embeddings,
		Metadatas:     metadatas,
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/collections/"+collectionName+"/add", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	return nil
}

// Query queries a collection with embeddings
func (c *ChromaDBClient) Query(collectionName string, queryEmbeddings [][]float32, nResults int, where map[string]interface{}) (*QueryResponse, error) {
	reqBody := QueryRequest{
		CollectionName: collectionName,
		QueryEmbeddings: queryEmbeddings,
		NResults:       nResults,
		Where:          where,
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", c.baseURL+"/api/v1/collections/"+collectionName+"/query", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &queryResp, nil
}

// CountRequest represents a count request
type CountRequest struct {
	CollectionName string                 `json:"collection_name"`
	Where          map[string]interface{} `json:"where,omitempty"`
}

// CountResponse represents a count response
type CountResponse struct {
	Count int `json:"count"`
}

// CountDocuments counts documents in a collection
// ChromaDB count endpoint requires POST with collection name in body
func (c *ChromaDBClient) CountDocuments(collectionName string, where map[string]interface{}) (int, error) {
	url := c.baseURL + "/api/v1/collections/" + collectionName + "/count"
	
	// ChromaDB count endpoint requires POST
	reqBody := CountRequest{
		CollectionName: collectionName,
		Where:          where,
	}
	
	data, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
	
	var countResp CountResponse
	if err := json.NewDecoder(resp.Body).Decode(&countResp); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return countResp.Count, nil
}
