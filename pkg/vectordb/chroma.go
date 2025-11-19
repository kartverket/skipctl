package vectordb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// ChromaClient handles interactions with a local Chroma vector database
type ChromaClient struct {
	baseURL    string
	httpClient *http.Client
	tenant     string
	database   string
}

// NewLocalChromaClient creates a client for a local Chroma instance
func NewLocalChromaClient(baseURL string) *ChromaClient {
	if baseURL == "" {
		baseURL = "http://localhost:8000" // Default Chroma port
	}
	return &ChromaClient{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{},
		tenant:     "default_tenant",
		database:   "default_database",
	}
}

// Collection represents a Chroma collection
type Collection struct {
	Name     string                 `json:"name"`
	ID       string                 `json:"id"`
	Metadata map[string]interface{} `json:"metadata"`
}

// QueryRequest represents a query to Chroma
type QueryRequest struct {
	QueryTexts []string `json:"query_texts"`
	NResults   int      `json:"n_results"`
}

// QueryResponse represents the response from a Chroma query
type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Documents [][]string                 `json:"documents"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Distances [][]float64                `json:"distances"`
}

// AddRequest represents adding documents to a collection
type AddRequest struct {
	Documents []string                 `json:"documents"`
	Metadatas []map[string]interface{} `json:"metadatas,omitempty"`
	IDs       []string                 `json:"ids"`
}

// GetOrCreateCollection gets or creates a collection by name
func (c *ChromaClient) GetOrCreateCollection(ctx context.Context, name string) (*Collection, error) {
	reqBody := map[string]interface{}{
		"name":          name,
		"get_or_create": true,
		"metadata": map[string]interface{}{
			"description": "ArgoKit documentation and examples with TF-IDF embeddings",
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Chroma v2 API: /api/v2/tenants/{tenant}/databases/{database}/collections
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections?get_or_create=true",
		c.baseURL, c.tenant, c.database)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var collection Collection
	if err := json.NewDecoder(resp.Body).Decode(&collection); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &collection, nil
}

// Query searches the collection for similar documents
// First we need to get the collection ID, then query using that ID
func (c *ChromaClient) Query(ctx context.Context, collectionName, queryText string, nResults int) ([]string, error) {
	if nResults <= 0 {
		nResults = 5
	}

	// Get collection to find its ID
	collection, err := c.GetOrCreateCollection(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection: %w", err)
	}

	// Generate TF-IDF embedding for query
	queryEmbedding := TFIDFEmbedding(queryText, 768)

	reqBody := map[string]interface{}{
		"query_embeddings": [][]float64{queryEmbedding},
		"n_results":        nResults,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Chroma v2 API: /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/query
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/query",
		c.baseURL, c.tenant, c.database, collection.ID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	var queryResp QueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract documents from the response
	var documents []string
	if len(queryResp.Documents) > 0 {
		documents = queryResp.Documents[0]
	}

	return documents, nil
}

// Add adds documents to a collection
func (c *ChromaClient) Add(ctx context.Context, collectionName string, documents []string, metadatas []map[string]interface{}, ids []string) error {
	if len(documents) != len(ids) {
		return fmt.Errorf("documents and ids must have the same length")
	}

	// Get collection to find its ID
	collection, err := c.GetOrCreateCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection: %w", err)
	}

	// Generate TF-IDF embeddings for all documents
	embeddings := make([][]float64, len(documents))
	for i, doc := range documents {
		embeddings[i] = TFIDFEmbedding(doc, 768)
	}

	reqBody := map[string]interface{}{
		"ids":        ids,
		"documents":  documents,
		"metadatas":  metadatas,
		"embeddings": embeddings,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Chroma v2 API: /api/v2/tenants/{tenant}/databases/{database}/collections/{collection_id}/add
	url := fmt.Sprintf("%s/api/v2/tenants/%s/databases/%s/collections/%s/add",
		c.baseURL, c.tenant, c.database, collection.ID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	return nil
}

// LoadDocumentsFromDirectory loads all .md and .jsonnet files from a directory into the collection
func (c *ChromaClient) LoadDocumentsFromDirectory(ctx context.Context, collectionName, dirPath string) (int, error) {
	var documents []string
	var metadatas []map[string]interface{}
	var ids []string

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".jsonnet" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		relPath, _ := filepath.Rel(dirPath, path)
		documents = append(documents, string(content))
		metadatas = append(metadatas, map[string]interface{}{
			"source":   relPath,
			"type":     ext,
			"filename": filepath.Base(path),
		})
		ids = append(ids, relPath)

		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to walk directory: %w", err)
	}

	if len(documents) == 0 {
		return 0, fmt.Errorf("no .md or .jsonnet files found in %s", dirPath)
	}

	// Add documents in batches (Chroma has limits)
	batchSize := 100
	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		if err := c.Add(ctx, collectionName, documents[i:end], metadatas[i:end], ids[i:end]); err != nil {
			return i, fmt.Errorf("failed to add batch: %w", err)
		}
	}

	return len(documents), nil
}

// IsAvailable checks if the Chroma server is running
func (c *ChromaClient) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/v2/heartbeat", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// TFIDFEmbedding generates an improved embedding vector using TF-IDF inspired approach
// This is better than simple hashing as it accounts for:
// - Character n-grams (captures subword information)
// - Term frequency (common words in document)
// - Lowercase normalization
func TFIDFEmbedding(text string, dimensions int) []float64 {
	if dimensions == 0 {
		dimensions = 768 // Larger dimension for better representation
	}

	vector := make([]float64, dimensions)
	text = strings.ToLower(text)

	// Extract character trigrams and words for better representation
	tokens := extractTokens(text)

	// Count frequency of each token
	freq := make(map[string]int)
	for _, token := range tokens {
		freq[token]++
	}

	// Build vector using consistent hashing with frequency weighting
	for token, count := range freq {
		// Use FNV-1a hash for better distribution
		hash := fnv1aHash(token)

		// Map to multiple positions for better coverage
		for i := 0; i < 3; i++ {
			idx := (hash + uint32(i)*2654435761) % uint32(dimensions)
			vector[idx] += float64(count) * (1.0 / (1.0 + float64(i)))
		}
	}

	// L2 normalization
	var magnitude float64
	for _, v := range vector {
		magnitude += v * v
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude > 0 {
		for i := range vector {
			vector[i] /= magnitude
		}
	}

	return vector
}

// extractTokens extracts words and character trigrams from text
func extractTokens(text string) []string {
	var tokens []string

	// Extract words
	var word strings.Builder
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			word.WriteRune(r)
		} else if word.Len() > 0 {
			tokens = append(tokens, word.String())
			word.Reset()
		}
	}
	if word.Len() > 0 {
		tokens = append(tokens, word.String())
	}

	// Extract character trigrams for subword matching
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, text)

	cleaned = strings.Join(strings.Fields(cleaned), " ")
	for i := 0; i <= len(cleaned)-3; i++ {
		trigram := cleaned[i : i+3]
		if len(strings.TrimSpace(trigram)) == 3 {
			tokens = append(tokens, trigram)
		}
	}

	return tokens
}

// fnv1aHash implements FNV-1a hash algorithm for better distribution
func fnv1aHash(s string) uint32 {
	const (
		offset32 = 2166136261
		prime32  = 16777619
	)
	hash := uint32(offset32)
	for i := 0; i < len(s); i++ {
		hash ^= uint32(s[i])
		hash *= prime32
	}
	return hash
}
