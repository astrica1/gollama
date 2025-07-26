package gollama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		host     []string
		expected string
	}{
		{
			name:     "no host defaults to localhost",
			host:     nil,
			expected: "http://localhost:11434",
		},
		{
			name:     "empty string host defaults to localhost",
			host:     []string{""},
			expected: "http://localhost:11434",
		},
		{
			name:     "custom host is used",
			host:     []string{"http://example.com:8080"},
			expected: "http://example.com:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var client *Client
			var err error

			if tt.host == nil {
				client, err = NewClient()
			} else {
				client, err = NewClient(tt.host...)
			}

			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			if client.BaseURL() != tt.expected {
				t.Errorf("NewClient() baseURL = %v, want %v", client.BaseURL(), tt.expected)
			}

			if client.httpClient == nil {
				t.Error("NewClient() httpClient is nil")
			}
		})
	}
}

func TestMessageJSON(t *testing.T) {
	message := Message{
		Role:    "user",
		Content: "Hello, world!",
	}

	data, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	expected := `{"role":"user","content":"Hello, world!"}`
	if string(data) != expected {
		t.Errorf("json.Marshal() = %v, want %v", string(data), expected)
	}

	var unmarshaled Message
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled != message {
		t.Errorf("json.Unmarshal() = %v, want %v", unmarshaled, message)
	}
}

func TestGenerateRequestJSON(t *testing.T) {
	req := GenerateRequest{
		Model:  "llama2",
		Prompt: "Tell me a joke",
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled GenerateRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Model != req.Model {
		t.Errorf("Model = %v, want %v", unmarshaled.Model, req.Model)
	}
	if unmarshaled.Prompt != req.Prompt {
		t.Errorf("Prompt = %v, want %v", unmarshaled.Prompt, req.Prompt)
	}
	if unmarshaled.Stream != req.Stream {
		t.Errorf("Stream = %v, want %v", unmarshaled.Stream, req.Stream)
	}
}

func TestChatRequestJSON(t *testing.T) {
	req := ChatRequest{
		Model: "llama2",
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
			{Role: "assistant", Content: "Hi there!"},
		},
		Stream: false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ChatRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Model != req.Model {
		t.Errorf("Model = %v, want %v", unmarshaled.Model, req.Model)
	}
	if len(unmarshaled.Messages) != len(req.Messages) {
		t.Errorf("Messages length = %v, want %v", len(unmarshaled.Messages), len(req.Messages))
	}
}

func TestModelResponseJSON(t *testing.T) {
	now := time.Now()
	model := ModelResponse{
		Name:       "llama2",
		ModifiedAt: now,
		Size:       1234567890,
		Digest:     "sha256:abc123",
		Details: ModelDetails{
			ParameterSize:     "7B",
			QuantizationLevel: "Q4_0",
			Family:            "llama",
		},
	}

	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ModelResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Name != model.Name {
		t.Errorf("Name = %v, want %v", unmarshaled.Name, model.Name)
	}
	if unmarshaled.Size != model.Size {
		t.Errorf("Size = %v, want %v", unmarshaled.Size, model.Size)
	}
	if unmarshaled.Digest != model.Digest {
		t.Errorf("Digest = %v, want %v", unmarshaled.Digest, model.Digest)
	}
}

func TestOllamaError(t *testing.T) {
	err := &OllamaError{
		StatusCode: 404,
		Message:    "Model not found",
	}

	expected := "Ollama API error (status 404): Model not found"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestParseErrorResponse(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		expected   string
	}{
		{
			name:       "valid JSON error",
			statusCode: 400,
			body:       []byte(`{"error":"Invalid model name"}`),
			expected:   "Ollama API error (status 400): Invalid model name",
		},
		{
			name:       "invalid JSON error",
			statusCode: 500,
			body:       []byte("Internal Server Error"),
			expected:   "Ollama API error (status 500): Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseErrorResponse(tt.statusCode, tt.body)
			if err.Error() != tt.expected {
				t.Errorf("parseErrorResponse() = %v, want %v", err.Error(), tt.expected)
			}
		})
	}
}

func TestClientList(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" || r.Method != http.MethodGet {
			t.Errorf("Expected GET /api/tags, got %s %s", r.Method, r.URL.Path)
		}

		response := ListModelsResponse{
			Models: []ModelResponse{
				{
					Name:       "llama2",
					ModifiedAt: time.Now(),
					Size:       1234567890,
					Digest:     "sha256:abc123",
					Details: ModelDetails{
						ParameterSize: "7B",
						Family:        "llama",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(result.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(result.Models))
	}

	if result.Models[0].Name != "llama2" {
		t.Errorf("Expected model name 'llama2', got %s", result.Models[0].Name)
	}
}

func TestClientShow(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/show" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/show, got %s %s", r.Method, r.URL.Path)
		}

		var req ShowRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		response := ModelResponse{
			Name:       "llama2",
			ModifiedAt: time.Now(),
			Size:       1234567890,
			Digest:     "sha256:abc123",
			Details: ModelDetails{
				ParameterSize: "7B",
				Family:        "llama",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.Show(ctx, "llama2")
	if err != nil {
		t.Fatalf("Show() error = %v", err)
	}

	if result.Name != "llama2" {
		t.Errorf("Expected model name 'llama2', got %s", result.Name)
	}
}

func TestClientShowEmptyModel(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	_, err = client.Show(ctx, "")
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}

	expectedMsg := "model name cannot be empty"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got %q", expectedMsg, err.Error())
	}
}

func TestClientCopy(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/copy" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/copy, got %s %s", r.Method, r.URL.Path)
		}

		var req CopyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Source != "llama2" {
			t.Errorf("Expected source 'llama2', got %s", req.Source)
		}

		if req.Destination != "llama2-copy" {
			t.Errorf("Expected destination 'llama2-copy', got %s", req.Destination)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Copy(ctx, "llama2", "llama2-copy")
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
}

func TestClientCopyValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty source
	err = client.Copy(ctx, "", "dest")
	if err == nil {
		t.Error("Expected error for empty source, got nil")
	}

	// Test empty destination
	err = client.Copy(ctx, "source", "")
	if err == nil {
		t.Error("Expected error for empty destination, got nil")
	}
}

func TestClientDelete(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/delete" || r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE /api/delete, got %s %s", r.Method, r.URL.Path)
		}

		var req DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Delete(ctx, "llama2")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestClientDeleteEmptyModel(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	err = client.Delete(ctx, "")
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}
}

func TestClientPull(t *testing.T) {
	// Create a mock server that returns streaming progress
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/pull" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/pull, got %s %s", r.Method, r.URL.Path)
		}

		var req PullRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simulate streaming progress responses
		progress1 := PullProgress{Status: "downloading", Digest: "sha256:abc", Total: 1000, Completed: 100}
		progress2 := PullProgress{Status: "downloading", Digest: "sha256:abc", Total: 1000, Completed: 500}
		progress3 := PullProgress{Status: "complete", Digest: "sha256:abc", Total: 1000, Completed: 1000}

		json.NewEncoder(w).Encode(progress1)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress2)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress3)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	var progressUpdates []PullProgress

	err = client.Pull(ctx, "llama2", func(progress PullProgress) {
		progressUpdates = append(progressUpdates, progress)
	})

	if err != nil {
		t.Fatalf("Pull() error = %v", err)
	}

	if len(progressUpdates) != 3 {
		t.Errorf("Expected 3 progress updates, got %d", len(progressUpdates))
	}

	// Check the final progress
	if len(progressUpdates) > 0 {
		final := progressUpdates[len(progressUpdates)-1]
		if final.Status != "complete" {
			t.Errorf("Expected final status 'complete', got %s", final.Status)
		}
		if final.Completed != 1000 {
			t.Errorf("Expected final completed 1000, got %d", final.Completed)
		}
	}
}

func TestClientPullValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty model name
	err = client.Pull(ctx, "", func(PullProgress) {})
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}

	// Test nil callback function
	err = client.Pull(ctx, "llama2", nil)
	if err == nil {
		t.Error("Expected error for nil callback function, got nil")
	}
}

func TestClientErrorHandling(t *testing.T) {
	// Create a mock server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ErrorResponse{Error: "model not found"})
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test List error
	_, err = client.List(ctx)
	if err == nil {
		t.Error("Expected error from List(), got nil")
	}

	var ollamaErr *OllamaError
	if !errors.As(err, &ollamaErr) {
		t.Errorf("Expected OllamaError, got %T: %v", err, err)
	} else {
		if ollamaErr.StatusCode != 404 {
			t.Errorf("Expected status code 404, got %d", ollamaErr.StatusCode)
		}
		if ollamaErr.Message != "model not found" {
			t.Errorf("Expected message 'model not found', got %q", ollamaErr.Message)
		}
	}
}

func TestPullProgressJSON(t *testing.T) {
	progress := PullProgress{
		Status:    "downloading",
		Digest:    "sha256:abc123",
		Total:     1000,
		Completed: 500,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled PullProgress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Status != progress.Status {
		t.Errorf("Status = %v, want %v", unmarshaled.Status, progress.Status)
	}
	if unmarshaled.Digest != progress.Digest {
		t.Errorf("Digest = %v, want %v", unmarshaled.Digest, progress.Digest)
	}
	if unmarshaled.Total != progress.Total {
		t.Errorf("Total = %v, want %v", unmarshaled.Total, progress.Total)
	}
	if unmarshaled.Completed != progress.Completed {
		t.Errorf("Completed = %v, want %v", unmarshaled.Completed, progress.Completed)
	}
}

func TestClientGenerate(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/generate, got %s %s", r.Method, r.URL.Path)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		if req.Prompt != "Tell me a joke" {
			t.Errorf("Expected prompt 'Tell me a joke', got %s", req.Prompt)
		}

		if req.Stream != false {
			t.Errorf("Expected Stream false for non-streaming request, got %t", req.Stream)
		}

		response := GenerateResponse{
			Model:     "llama2",
			CreatedAt: time.Now(),
			Response:  "Why don't scientists trust atoms? Because they make up everything!",
			Done:      true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &GenerateRequest{
		Model:  "llama2",
		Prompt: "Tell me a joke",
	}

	result, err := client.Generate(ctx, req)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result.Model != "llama2" {
		t.Errorf("Expected model 'llama2', got %s", result.Model)
	}

	if !result.Done {
		t.Errorf("Expected Done true, got %t", result.Done)
	}

	if !strings.Contains(result.Response, "atoms") {
		t.Errorf("Expected response to contain 'atoms', got %s", result.Response)
	}
}

func TestClientGenerateValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil request
	_, err = client.Generate(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	// Test empty model
	req := &GenerateRequest{
		Model:  "",
		Prompt: "test",
	}
	_, err = client.Generate(ctx, req)
	if err == nil {
		t.Error("Expected error for empty model, got nil")
	}
}

func TestClientGenerateStream(t *testing.T) {
	// Create a mock server that returns streaming responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/generate" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/generate, got %s %s", r.Method, r.URL.Path)
		}

		var req GenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Stream != true {
			t.Errorf("Expected Stream true for streaming request, got %t", req.Stream)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simulate streaming responses
		responses := []GenerateResponse{
			{Model: "llama2", Response: "Why", Done: false},
			{Model: "llama2", Response: " don't", Done: false},
			{Model: "llama2", Response: " scientists", Done: false},
			{Model: "llama2", Response: " trust atoms?", Done: true},
		}

		for i, resp := range responses {
			json.NewEncoder(w).Encode(resp)
			if i < len(responses)-1 {
				w.Write([]byte("\n"))
			}
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &GenerateRequest{
		Model:  "llama2",
		Prompt: "Tell me a joke",
	}

	var responses []GenerateResponse
	err = client.GenerateStream(ctx, req, func(resp *GenerateResponse) {
		responses = append(responses, *resp)
	})

	if err != nil {
		t.Fatalf("GenerateStream() error = %v", err)
	}

	if len(responses) != 4 {
		t.Errorf("Expected 4 responses, got %d", len(responses))
	}

	// Check the final response
	if len(responses) > 0 {
		final := responses[len(responses)-1]
		if !final.Done {
			t.Errorf("Expected final response Done true, got %t", final.Done)
		}
	}

	// Concatenate all response text
	var fullResponse strings.Builder
	for _, resp := range responses {
		fullResponse.WriteString(resp.Response)
	}
	fullText := fullResponse.String()

	if !strings.Contains(fullText, "scientists") {
		t.Errorf("Expected full response to contain 'scientists', got %s", fullText)
	}
}

func TestClientGenerateStreamValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil request
	err = client.GenerateStream(ctx, nil, func(*GenerateResponse) {})
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	// Test empty model
	req := &GenerateRequest{
		Model:  "",
		Prompt: "test",
	}
	err = client.GenerateStream(ctx, req, func(*GenerateResponse) {})
	if err == nil {
		t.Error("Expected error for empty model, got nil")
	}

	// Test nil callback
	req = &GenerateRequest{
		Model:  "llama2",
		Prompt: "test",
	}
	err = client.GenerateStream(ctx, req, nil)
	if err == nil {
		t.Error("Expected error for nil callback, got nil")
	}
}

func TestClientChat(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/chat, got %s %s", r.Method, r.URL.Path)
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		if len(req.Messages) != 1 {
			t.Errorf("Expected 1 message, got %d", len(req.Messages))
		}

		if req.Stream != false {
			t.Errorf("Expected Stream false for non-streaming request, got %t", req.Stream)
		}

		response := ChatResponse{
			Model:     "llama2",
			CreatedAt: time.Now(),
			Message: Message{
				Role:    "assistant",
				Content: "Hello! How can I help you today?",
			},
			Done: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &ChatRequest{
		Model: "llama2",
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	result, err := client.Chat(ctx, req)
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}

	if result.Model != "llama2" {
		t.Errorf("Expected model 'llama2', got %s", result.Model)
	}

	if !result.Done {
		t.Errorf("Expected Done true, got %t", result.Done)
	}

	if result.Message.Role != "assistant" {
		t.Errorf("Expected role 'assistant', got %s", result.Message.Role)
	}

	if !strings.Contains(result.Message.Content, "help") {
		t.Errorf("Expected response to contain 'help', got %s", result.Message.Content)
	}
}

func TestClientChatValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil request
	_, err = client.Chat(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	// Test empty model
	req := &ChatRequest{
		Model:    "",
		Messages: []Message{{Role: "user", Content: "test"}},
	}
	_, err = client.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error for empty model, got nil")
	}

	// Test empty messages
	req = &ChatRequest{
		Model:    "llama2",
		Messages: []Message{},
	}
	_, err = client.Chat(ctx, req)
	if err == nil {
		t.Error("Expected error for empty messages, got nil")
	}
}

func TestClientChatStream(t *testing.T) {
	// Create a mock server that returns streaming responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/chat, got %s %s", r.Method, r.URL.Path)
		}

		var req ChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Stream != true {
			t.Errorf("Expected Stream true for streaming request, got %t", req.Stream)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simulate streaming responses
		responses := []ChatResponse{
			{Model: "llama2", Message: Message{Role: "assistant", Content: "Hello"}, Done: false},
			{Model: "llama2", Message: Message{Role: "assistant", Content: "! How"}, Done: false},
			{Model: "llama2", Message: Message{Role: "assistant", Content: " can I help"}, Done: false},
			{Model: "llama2", Message: Message{Role: "assistant", Content: " you?"}, Done: true},
		}

		for i, resp := range responses {
			json.NewEncoder(w).Encode(resp)
			if i < len(responses)-1 {
				w.Write([]byte("\n"))
			}
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &ChatRequest{
		Model: "llama2",
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	var responses []ChatResponse
	err = client.ChatStream(ctx, req, func(resp *ChatResponse) {
		responses = append(responses, *resp)
	})

	if err != nil {
		t.Fatalf("ChatStream() error = %v", err)
	}

	if len(responses) != 4 {
		t.Errorf("Expected 4 responses, got %d", len(responses))
	}

	// Check the final response
	if len(responses) > 0 {
		final := responses[len(responses)-1]
		if !final.Done {
			t.Errorf("Expected final response Done true, got %t", final.Done)
		}
	}

	// Concatenate all message content
	var fullResponse strings.Builder
	for _, resp := range responses {
		fullResponse.WriteString(resp.Message.Content)
	}
	fullText := fullResponse.String()

	if !strings.Contains(fullText, "help") {
		t.Errorf("Expected full response to contain 'help', got %s", fullText)
	}
}

func TestClientChatStreamValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil request
	err = client.ChatStream(ctx, nil, func(*ChatResponse) {})
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	// Test empty model
	req := &ChatRequest{
		Model:    "",
		Messages: []Message{{Role: "user", Content: "test"}},
	}
	err = client.ChatStream(ctx, req, func(*ChatResponse) {})
	if err == nil {
		t.Error("Expected error for empty model, got nil")
	}

	// Test empty messages
	req = &ChatRequest{
		Model:    "llama2",
		Messages: []Message{},
	}
	err = client.ChatStream(ctx, req, func(*ChatResponse) {})
	if err == nil {
		t.Error("Expected error for empty messages, got nil")
	}

	// Test nil callback
	req = &ChatRequest{
		Model: "llama2",
		Messages: []Message{
			{Role: "user", Content: "test"},
		},
	}
	err = client.ChatStream(ctx, req, nil)
	if err == nil {
		t.Error("Expected error for nil callback, got nil")
	}
}

func TestClientCreate(t *testing.T) {
	// Create a mock server that returns streaming progress
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/create" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/create, got %s %s", r.Method, r.URL.Path)
		}

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "my-model" {
			t.Errorf("Expected model 'my-model', got %s", req.Model)
		}

		if !strings.Contains(req.Modelfile, "FROM llama2") {
			t.Errorf("Expected Modelfile to contain 'FROM llama2', got %s", req.Modelfile)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simulate streaming progress responses
		progress1 := CreateProgress{Status: "reading model metadata"}
		progress2 := CreateProgress{Status: "creating model layer"}
		progress3 := CreateProgress{Status: "success"}

		json.NewEncoder(w).Encode(progress1)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress2)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress3)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	var progressUpdates []CreateProgress

	modelfile := "FROM llama2\nSYSTEM You are a helpful assistant."
	err = client.Create(ctx, "my-model", modelfile, func(progress CreateProgress) {
		progressUpdates = append(progressUpdates, progress)
	})

	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if len(progressUpdates) != 3 {
		t.Errorf("Expected 3 progress updates, got %d", len(progressUpdates))
	}

	// Check the final progress
	if len(progressUpdates) > 0 {
		final := progressUpdates[len(progressUpdates)-1]
		if final.Status != "success" {
			t.Errorf("Expected final status 'success', got %s", final.Status)
		}
	}
}

func TestClientCreateValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty model name
	err = client.Create(ctx, "", "FROM llama2", func(CreateProgress) {})
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}

	// Test empty modelfile content
	err = client.Create(ctx, "my-model", "", func(CreateProgress) {})
	if err == nil {
		t.Error("Expected error for empty modelfile content, got nil")
	}

	// Test nil callback function
	err = client.Create(ctx, "my-model", "FROM llama2", nil)
	if err == nil {
		t.Error("Expected error for nil callback function, got nil")
	}
}

func TestClientPush(t *testing.T) {
	// Create a mock server that returns streaming progress
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/push" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/push, got %s %s", r.Method, r.URL.Path)
		}

		var req PushRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "my-model" {
			t.Errorf("Expected model 'my-model', got %s", req.Model)
		}

		w.Header().Set("Content-Type", "application/json")

		// Simulate streaming progress responses
		progress1 := PushProgress{Status: "preparing", Digest: "sha256:abc", Total: 1000, Completed: 0}
		progress2 := PushProgress{Status: "pushing", Digest: "sha256:abc", Total: 1000, Completed: 500}
		progress3 := PushProgress{Status: "success", Digest: "sha256:abc", Total: 1000, Completed: 1000}

		json.NewEncoder(w).Encode(progress1)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress2)
		w.Write([]byte("\n"))
		json.NewEncoder(w).Encode(progress3)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	var progressUpdates []PushProgress

	err = client.Push(ctx, "my-model", func(progress PushProgress) {
		progressUpdates = append(progressUpdates, progress)
	})

	if err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	if len(progressUpdates) != 3 {
		t.Errorf("Expected 3 progress updates, got %d", len(progressUpdates))
	}

	// Check the final progress
	if len(progressUpdates) > 0 {
		final := progressUpdates[len(progressUpdates)-1]
		if final.Status != "success" {
			t.Errorf("Expected final status 'success', got %s", final.Status)
		}
		if final.Completed != 1000 {
			t.Errorf("Expected final completed 1000, got %d", final.Completed)
		}
	}
}

func TestClientPushValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test empty model name
	err = client.Push(ctx, "", func(PushProgress) {})
	if err == nil {
		t.Error("Expected error for empty model name, got nil")
	}

	// Test nil callback function
	err = client.Push(ctx, "my-model", nil)
	if err == nil {
		t.Error("Expected error for nil callback function, got nil")
	}
}

func TestClientEmbeddings(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" || r.Method != http.MethodPost {
			t.Errorf("Expected POST /api/embeddings, got %s %s", r.Method, r.URL.Path)
		}

		var req EmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if req.Model != "llama2" {
			t.Errorf("Expected model 'llama2', got %s", req.Model)
		}

		if req.Prompt != "Hello world" {
			t.Errorf("Expected prompt 'Hello world', got %s", req.Prompt)
		}

		response := EmbeddingResponse{
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	req := &EmbeddingRequest{
		Model:  "llama2",
		Prompt: "Hello world",
	}

	result, err := client.Embeddings(ctx, req)
	if err != nil {
		t.Fatalf("Embeddings() error = %v", err)
	}

	if len(result.Embedding) != 5 {
		t.Errorf("Expected embedding length 5, got %d", len(result.Embedding))
	}

	expectedEmbedding := []float64{0.1, 0.2, 0.3, 0.4, 0.5}
	for i, val := range result.Embedding {
		if val != expectedEmbedding[i] {
			t.Errorf("Expected embedding[%d] = %f, got %f", i, expectedEmbedding[i], val)
		}
	}
}

func TestClientEmbeddingsValidation(t *testing.T) {
	client, err := NewClient()
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test nil request
	_, err = client.Embeddings(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil request, got nil")
	}

	// Test empty model
	req := &EmbeddingRequest{
		Model:  "",
		Prompt: "test",
	}
	_, err = client.Embeddings(ctx, req)
	if err == nil {
		t.Error("Expected error for empty model, got nil")
	}

	// Test empty prompt
	req = &EmbeddingRequest{
		Model:  "llama2",
		Prompt: "",
	}
	_, err = client.Embeddings(ctx, req)
	if err == nil {
		t.Error("Expected error for empty prompt, got nil")
	}
}

func TestClientPS(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ps" || r.Method != http.MethodGet {
			t.Errorf("Expected GET /api/ps, got %s %s", r.Method, r.URL.Path)
		}

		response := PSResponse{
			Models: []ModelResponse{
				{
					Name:       "llama2",
					ModifiedAt: time.Now(),
					Size:       1234567890,
					Digest:     "sha256:abc123",
					Details: ModelDetails{
						ParameterSize: "7B",
						Family:        "llama",
					},
				},
				{
					Name:       "codellama",
					ModifiedAt: time.Now(),
					Size:       987654321,
					Digest:     "sha256:def456",
					Details: ModelDetails{
						ParameterSize: "13B",
						Family:        "codellama",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, err := NewClient(server.URL)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	result, err := client.PS(ctx)
	if err != nil {
		t.Fatalf("PS() error = %v", err)
	}

	if len(result.Models) != 2 {
		t.Errorf("Expected 2 running models, got %d", len(result.Models))
	}

	// Check first model
	if result.Models[0].Name != "llama2" {
		t.Errorf("Expected first model name 'llama2', got %s", result.Models[0].Name)
	}

	// Check second model
	if result.Models[1].Name != "codellama" {
		t.Errorf("Expected second model name 'codellama', got %s", result.Models[1].Name)
	}
}

func TestCreateProgressJSON(t *testing.T) {
	progress := CreateProgress{
		Status: "creating model layer",
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CreateProgress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Status != progress.Status {
		t.Errorf("Status = %v, want %v", unmarshaled.Status, progress.Status)
	}
}

func TestPushProgressJSON(t *testing.T) {
	progress := PushProgress{
		Status:    "pushing",
		Digest:    "sha256:abc123",
		Total:     1000,
		Completed: 500,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled PushProgress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Status != progress.Status {
		t.Errorf("Status = %v, want %v", unmarshaled.Status, progress.Status)
	}
	if unmarshaled.Digest != progress.Digest {
		t.Errorf("Digest = %v, want %v", unmarshaled.Digest, progress.Digest)
	}
	if unmarshaled.Total != progress.Total {
		t.Errorf("Total = %v, want %v", unmarshaled.Total, progress.Total)
	}
	if unmarshaled.Completed != progress.Completed {
		t.Errorf("Completed = %v, want %v", unmarshaled.Completed, progress.Completed)
	}
}
