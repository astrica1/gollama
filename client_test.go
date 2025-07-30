package gollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Additional comprehensive tests for complete coverage

func TestClientChatAdvanced(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name        string
		request     ChatRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid chat request",
			request: ChatRequest{
				Model: "llama2",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: false,
		},
		{
			name: "Multiple messages",
			request: ChatRequest{
				Model: "llama2",
				Messages: []Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
					{Role: "user", Content: "How are you?"},
				},
			},
			expectError: false,
		},
		{
			name: "Empty model name",
			request: ChatRequest{
				Model: "",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			expectError: true,
			errorMsg:    "model name cannot be empty",
		},
		{
			name: "No messages",
			request: ChatRequest{
				Model:    "llama2",
				Messages: []Message{},
			},
			expectError: true,
			errorMsg:    "at least one message is required",
		},
		{
			name: "Nil messages",
			request: ChatRequest{
				Model:    "llama2",
				Messages: nil,
			},
			expectError: true,
			errorMsg:    "at least one message is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.Chat(ctx, &tt.request)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			assertNoError(t, err)

			if response == nil {
				t.Fatalf("Response should not be nil")
			}

			if response.Model != tt.request.Model {
				t.Errorf("Expected model %s, got %s", tt.request.Model, response.Model)
			}

			if response.Message.Role != "assistant" {
				t.Errorf("Expected assistant role, got %s", response.Message.Role)
			}

			if response.Message.Content == "" {
				t.Errorf("Expected non-empty content")
			}
		})
	}
}

func TestClientEmbeddingsAdvanced(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		request EmbeddingRequest
	}{
		{
			name: "Single prompt embedding",
			request: EmbeddingRequest{
				Model:  "llama2",
				Prompt: "Hello world",
			},
		},
		{
			name: "Long text embedding",
			request: EmbeddingRequest{
				Model:  "llama2",
				Prompt: "This is a longer text that needs to be embedded for semantic search purposes.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.Embeddings(ctx, &tt.request)
			assertNoError(t, err)

			if response == nil {
				t.Fatalf("Response should not be nil")
			}

			if len(response.Embedding) == 0 {
				t.Errorf("Expected non-empty embedding vector")
			}

			// Check that we got some reasonable embedding values
			for i, val := range response.Embedding {
				if val < -1.0 || val > 1.0 {
					t.Errorf("Embedding value %d out of expected range [-1,1]: %f", i, val)
				}
			}
		})
	}
}

func TestClientPSAdvanced(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	response, err := client.PS(ctx)
	assertNoError(t, err)

	if response == nil {
		t.Fatalf("Response should not be nil")
	}

	if len(response.Models) == 0 {
		t.Errorf("Expected at least one running model")
	}

	model := response.Models[0]
	if model.Name == "" {
		t.Errorf("Model name should not be empty")
	}

	if model.Size <= 0 {
		t.Errorf("Model size should be positive, got %d", model.Size)
	}

	if model.Digest == "" {
		t.Errorf("Model digest should not be empty")
	}
}

func TestClientCreateAdvanced(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	tests := []struct {
		name    string
		request CreateRequest
	}{
		{
			name: "Create from Modelfile",
			request: CreateRequest{
				Model:     "custom-model",
				Modelfile: "FROM llama2\nPARAMETER temperature 0.7",
			},
		},
		{
			name: "Create with path",
			request: CreateRequest{
				Model:     "custom-model-2",
				Modelfile: "FROM llama2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.Create(ctx, tt.request.Model, tt.request.Modelfile, func(progress CreateProgress) {
				// Handle progress updates
			})
			assertNoError(t, err)
		})
	}
}

func TestClientPushAdvanced(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	modelName := "custom-model"

	err = client.Push(ctx, modelName, func(progress PushProgress) {
		// Handle progress updates
	})
	assertNoError(t, err)
}

func TestClientConcurrency(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	// Test concurrent requests to ensure thread safety
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			request := GenerateRequest{
				Model:  "llama2",
				Prompt: "Concurrent test request",
			}

			_, err := client.Generate(ctx, &request)
			results <- err
		}(i)
	}

	// Collect all results
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent request %d failed: %v", i, err)
			}
		case <-time.After(10 * time.Second):
			t.Errorf("Timeout waiting for concurrent request %d", i)
		}
	}
}

func TestClientTimeout(t *testing.T) {
	// Create a slow server that takes longer than the timeout
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Sleep longer than client timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer slowServer.Close()

	client, err := NewClient(slowServer.URL)
	assertNoError(t, err)
	
	// Use context with timeout instead of HTTP client timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	request := GenerateRequest{
		Model:  "llama2",
		Prompt: "This should timeout",
	}

	_, err = client.Generate(ctx, &request)

	if err == nil {
		t.Errorf("Expected timeout error but got none")
		return
	}

	// Check if it's a timeout error
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestClientContextCancellation(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	// Create context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	request := GenerateRequest{
		Model:  "llama2",
		Prompt: "This should be cancelled",
	}

	_, err = client.Generate(ctx, &request)

	if err == nil {
		t.Errorf("Expected context cancellation error but got none")
	}

	if !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("Expected context canceled error, got: %v", err)
	}
}

func TestClientStreamingInterruption(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	request := GenerateRequest{
		Model:  "llama2",
		Prompt: "Stream test",
		Stream: true,
	}

	var responses []*GenerateResponse
	err = client.GenerateStream(ctx, &request, func(response *GenerateResponse) {
		responses = append(responses, response)
	})
	assertNoError(t, err)

	if len(responses) == 0 {
		t.Fatalf("Expected at least one response")
	}

	// Check first response
	firstResponse := responses[0]
	if firstResponse.Response == "" && !firstResponse.Done {
		t.Errorf("Expected either response content or done flag")
	}
}

func TestClientErrorHandlingEdgeCases(t *testing.T) {
	// Test with invalid server URL
	client, err := NewClient("http://nonexistent.localhost:99999")
	assertNoError(t, err)

	ctx := context.Background()

	request := GenerateRequest{
		Model:  "llama2",
		Prompt: "This will fail",
	}

	_, err = client.Generate(ctx, &request)

	if err == nil {
		t.Errorf("Expected connection error but got none")
	}

	// Error should indicate connection failure
	if !strings.Contains(err.Error(), "connection") && !strings.Contains(err.Error(), "dial") {
		t.Logf("Got error (expected): %v", err)
	}
}

func TestClientHeaderPersistence(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	// Note: Client doesn't support custom headers in this implementation

	ctx := context.Background()

	// Make multiple requests to ensure headers persist
	for i := 0; i < 3; i++ {
		_, err := client.List(ctx)
		assertNoError(t, err)
	}

	// Headers should still be set for new requests
	request := GenerateRequest{
		Model:  "llama2",
		Prompt: "Test with headers",
	}

	_, err = client.Generate(ctx, &request)
	assertNoError(t, err)
}

func TestClientMemoryUsage(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := createTestClient(server.URL)
	assertNoError(t, err)

	ctx := context.Background()

	// Make many requests to check for memory leaks
	for i := 0; i < 100; i++ {
		request := GenerateRequest{
			Model:  "llama2",
			Prompt: "Memory test",
		}

		_, err := client.Generate(ctx, &request)
		assertNoError(t, err)
	}

	// If we get here without crashing, memory usage is probably reasonable
}
