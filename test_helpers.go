package gollama

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Test utilities and helper functions

// setupMockServer creates a test HTTP server that simulates Ollama API responses
func setupMockServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set common headers
		w.Header().Set("Content-Type", "application/json")

		// Route based on URL path
		switch r.URL.Path {
		case "/api/tags":
			handleListModels(w, r)
		case "/api/show":
			handleShowModel(w, r)
		case "/api/generate":
			handleGenerate(w, r)
		case "/api/chat":
			handleChat(w, r)
		case "/api/embeddings":
			handleEmbeddings(w, r)
		case "/api/copy":
			handleCopyModel(w, r)
		case "/api/delete":
			handleDeleteModel(w, r)
		case "/api/pull":
			handlePullModel(w, r)
		case "/api/create":
			handleCreateModel(w, r)
		case "/api/push":
			handlePushModel(w, r)
		case "/api/ps":
			handlePS(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
}

// Mock handlers for different API endpoints

func handleListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := ListModelsResponse{
		Models: []ModelResponse{
			{
				Name:       "llama2",
				ModifiedAt: time.Now(),
				Size:       3825819519,
				Digest:     "sha256:1a838c4c",
			},
			{
				Name:       "codellama",
				ModifiedAt: time.Now(),
				Size:       3825819519,
				Digest:     "sha256:2b947d5f",
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

func handleShowModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Model == "nonexistent" {
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	response := ModelResponse{
		Name:       req.Model,
		ModifiedAt: time.Now(),
		Size:       7323310500,
		Digest:     "sha256:bc07c81de745",
	}

	json.NewEncoder(w).Encode(response)
}

func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, "Model name required", http.StatusBadRequest)
		return
	}

	if req.Prompt == "error" {
		http.Error(w, "Generation failed", http.StatusInternalServerError)
		return
	}

	response := GenerateResponse{
		Model:              req.Model,
		CreatedAt:          time.Now(),
		Response:           "This is a test response to: " + req.Prompt,
		Done:               true,
		TotalDuration:      1234567890,
		LoadDuration:       123456789,
		PromptEvalCount:    10,
		PromptEvalDuration: 987654321,
		EvalCount:          20,
		EvalDuration:       876543210,
	}

	json.NewEncoder(w).Encode(response)
}

func handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Model == "" {
		http.Error(w, "Model name required", http.StatusBadRequest)
		return
	}

	if len(req.Messages) == 0 {
		http.Error(w, "Messages required", http.StatusBadRequest)
		return
	}

	response := ChatResponse{
		Model:     req.Model,
		CreatedAt: time.Now(),
		Message: Message{
			Role:    "assistant",
			Content: "This is a test chat response",
		},
		Done: true,
	}

	json.NewEncoder(w).Encode(response)
}

func handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req EmbeddingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	response := EmbeddingResponse{
		Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
	}

	json.NewEncoder(w).Encode(response)
}

func handleCopyModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CopyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Source == "nonexistent" {
		http.Error(w, "Source model not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleDeleteModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Model == "nonexistent" {
		http.Error(w, "Model not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handlePullModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PullRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simulate pull progress
	progress := PullProgress{
		Status:    "downloading",
		Digest:    "sha256:1a838c4c",
		Total:     1000,
		Completed: 500,
	}

	json.NewEncoder(w).Encode(progress)
}

func handleCreateModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	progress := CreateProgress{
		Status: "creating model layer",
	}

	json.NewEncoder(w).Encode(progress)
}

func handlePushModel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PushRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	progress := PushProgress{
		Status:    "pushing",
		Digest:    "sha256:1a838c4c",
		Total:     1000,
		Completed: 250,
	}

	json.NewEncoder(w).Encode(progress)
}

func handlePS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := PSResponse{
		Models: []ModelResponse{
			{
				Name:       "llama2",
				Size:       3825819519,
				Digest:     "sha256:1a838c4c",
				ModifiedAt: time.Now().Add(-time.Hour),
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}

// createTestClient creates a client pointed at the test server
func createTestClient(serverURL string) (*Client, error) {
	return NewClient(serverURL)
}

// assertJSONResponse compares expected and actual JSON responses
func assertJSONResponse(t *testing.T, expected, actual interface{}) {
	t.Helper()

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to marshal expected response: %v", err)
	}

	actualJSON, err := json.Marshal(actual)
	if err != nil {
		t.Fatalf("Failed to marshal actual response: %v", err)
	}

	if !bytes.Equal(expectedJSON, actualJSON) {
		t.Errorf("JSON responses don't match.\nExpected: %s\nActual: %s", expectedJSON, actualJSON)
	}
}

// assertErrorContains checks if error contains expected substring
func assertErrorContains(t *testing.T, err error, expected string) {
	t.Helper()

	if err == nil {
		t.Fatalf("Expected error containing '%s', got nil", expected)
	}

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error to contain '%s', got '%s'", expected, err.Error())
	}
}

// assertNoError checks that no error occurred
func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
}
