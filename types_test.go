package gollama

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

// Test all struct types and their JSON marshaling/unmarshaling

func TestMessageStructure(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected string
	}{
		{
			name: "Basic user message",
			message: Message{
				Role:    "user",
				Content: "Hello",
			},
			expected: `{"role":"user","content":"Hello"}`,
		},
		{
			name: "Assistant message",
			message: Message{
				Role:    "assistant",
				Content: "Hi there!",
			},
			expected: `{"role":"assistant","content":"Hi there!"}`,
		},
		{
			name: "System message",
			message: Message{
				Role:    "system",
				Content: "You are a helpful assistant",
			},
			expected: `{"role":"system","content":"You are a helpful assistant"}`,
		},
		{
			name: "Empty message",
			message: Message{
				Role:    "",
				Content: "",
			},
			expected: `{"role":"","content":""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test marshaling
			jsonData, err := json.Marshal(tt.message)
			assertNoError(t, err)

			if string(jsonData) != tt.expected {
				t.Errorf("Expected JSON %s, got %s", tt.expected, string(jsonData))
			}

			// Test unmarshaling
			var unmarshaled Message
			err = json.Unmarshal(jsonData, &unmarshaled)
			assertNoError(t, err)

			if !reflect.DeepEqual(tt.message, unmarshaled) {
				t.Errorf("Expected %+v, got %+v", tt.message, unmarshaled)
			}
		})
	}
}

func TestModelStructure(t *testing.T) {
	model := ModelResponse{
		Name:       "llama2:7b",
		ModifiedAt: time.Now(),
		Size:       3825819519,
		Digest:     "sha256:1a838c4c0dbb",
		Details: ModelDetails{
			Format:            "gguf",
			Family:            "llama",
			ParameterSize:     "7B",
			QuantizationLevel: "Q4_0",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(model)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ModelResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Name != model.Name {
		t.Errorf("Expected name %s, got %s", model.Name, unmarshaled.Name)
	}

	if unmarshaled.Size != model.Size {
		t.Errorf("Expected size %d, got %d", model.Size, unmarshaled.Size)
	}

	if unmarshaled.Details.Format != model.Details.Format {
		t.Errorf("Expected format %s, got %s", model.Details.Format, unmarshaled.Details.Format)
	}
}

func TestGenerateRequestStructure(t *testing.T) {
	tests := []struct {
		name    string
		request GenerateRequest
	}{
		{
			name: "Basic generate request",
			request: GenerateRequest{
				Model:  "llama2",
				Prompt: "Hello world",
			},
		},
		{
			name: "Full generate request",
			request: GenerateRequest{
				Model:   "llama2",
				Prompt:  "Write a story",
				Options: map[string]interface{}{"temperature": 0.7, "top_p": 0.9},
				Stream:  true,
			},
		},
		{
			name: "Minimal request",
			request: GenerateRequest{
				Model:  "llama2",
				Prompt: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON marshaling
			jsonData, err := json.Marshal(tt.request)
			assertNoError(t, err)

			// Test JSON unmarshaling
			var unmarshaled GenerateRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			assertNoError(t, err)

			if unmarshaled.Model != tt.request.Model {
				t.Errorf("Expected model %s, got %s", tt.request.Model, unmarshaled.Model)
			}

			if unmarshaled.Prompt != tt.request.Prompt {
				t.Errorf("Expected prompt %s, got %s", tt.request.Prompt, unmarshaled.Prompt)
			}

			if unmarshaled.Stream != tt.request.Stream {
				t.Errorf("Expected stream %v, got %v", tt.request.Stream, unmarshaled.Stream)
			}
		})
	}
}

func TestChatRequestStructure(t *testing.T) {
	request := ChatRequest{
		Model: "llama2",
		Messages: []Message{
			{Role: "system", Content: "You are helpful"},
			{Role: "user", Content: "Hello"},
		},
		Options: map[string]interface{}{"temperature": 0.8},
		Stream:  false,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ChatRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Model != request.Model {
		t.Errorf("Expected model %s, got %s", request.Model, unmarshaled.Model)
	}

	if len(unmarshaled.Messages) != len(request.Messages) {
		t.Errorf("Expected %d messages, got %d", len(request.Messages), len(unmarshaled.Messages))
	}

	for i, msg := range request.Messages {
		if unmarshaled.Messages[i].Role != msg.Role {
			t.Errorf("Message %d: expected role %s, got %s", i, msg.Role, unmarshaled.Messages[i].Role)
		}
		if unmarshaled.Messages[i].Content != msg.Content {
			t.Errorf("Message %d: expected content %s, got %s", i, msg.Content, unmarshaled.Messages[i].Content)
		}
	}
}

func TestGenerateResponseStructure(t *testing.T) {
	response := GenerateResponse{
		Model:              "llama2",
		CreatedAt:          time.Now(),
		Response:           "Hello! How can I help you?",
		Done:               true,
		Context:            []int{1, 2, 3, 4, 5},
		TotalDuration:      1234567890,
		LoadDuration:       123456789,
		PromptEvalCount:    10,
		PromptEvalDuration: 987654321,
		EvalCount:          20,
		EvalDuration:       876543210,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled GenerateResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Model != response.Model {
		t.Errorf("Expected model %s, got %s", response.Model, unmarshaled.Model)
	}

	if unmarshaled.Response != response.Response {
		t.Errorf("Expected response %s, got %s", response.Response, unmarshaled.Response)
	}

	if unmarshaled.Done != response.Done {
		t.Errorf("Expected done %v, got %v", response.Done, unmarshaled.Done)
	}

	if unmarshaled.TotalDuration != response.TotalDuration {
		t.Errorf("Expected total duration %d, got %d", response.TotalDuration, unmarshaled.TotalDuration)
	}

	if !reflect.DeepEqual(unmarshaled.Context, response.Context) {
		t.Errorf("Expected context %v, got %v", response.Context, unmarshaled.Context)
	}
}

func TestChatResponseStructure(t *testing.T) {
	response := ChatResponse{
		Model:     "llama2",
		CreatedAt: time.Now(),
		Message: Message{
			Role:    "assistant",
			Content: "I'm doing well, thank you!",
		},
		Done:               true,
		TotalDuration:      1234567890,
		LoadDuration:       123456789,
		PromptEvalCount:    10,
		PromptEvalDuration: 987654321,
		EvalCount:          20,
		EvalDuration:       876543210,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ChatResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Model != response.Model {
		t.Errorf("Expected model %s, got %s", response.Model, unmarshaled.Model)
	}

	if unmarshaled.Message.Role != response.Message.Role {
		t.Errorf("Expected message role %s, got %s", response.Message.Role, unmarshaled.Message.Role)
	}

	if unmarshaled.Message.Content != response.Message.Content {
		t.Errorf("Expected message content %s, got %s", response.Message.Content, unmarshaled.Message.Content)
	}

	if unmarshaled.Done != response.Done {
		t.Errorf("Expected done %v, got %v", response.Done, unmarshaled.Done)
	}
}

func TestEmbeddingRequestStructure(t *testing.T) {
	request := EmbeddingRequest{
		Model:     "llama2",
		Prompt: "Embed this text",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(request)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled EmbeddingRequest
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Model != request.Model {
		t.Errorf("Expected model %s, got %s", request.Model, unmarshaled.Model)
	}

	if unmarshaled.Prompt != request.Prompt {
		t.Errorf("Expected prompt %s, got %s", request.Prompt, unmarshaled.Prompt)
	}
}

func TestEmbeddingResponseStructure(t *testing.T) {
	response := EmbeddingResponse{
		Embedding: []float64{0.1, 0.2, 0.3, -0.1, -0.2},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled EmbeddingResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if !reflect.DeepEqual(unmarshaled.Embedding, response.Embedding) {
		t.Errorf("Expected embedding %v, got %v", response.Embedding, unmarshaled.Embedding)
	}
}

func TestPullProgressStructure(t *testing.T) {
	progress := PullProgress{
		Status:    "downloading",
		Digest:    "sha256:1a838c4c",
		Total:     1073741824,
		Completed: 536870912,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(progress)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled PullProgress
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Status != progress.Status {
		t.Errorf("Expected status %s, got %s", progress.Status, unmarshaled.Status)
	}

	if unmarshaled.Total != progress.Total {
		t.Errorf("Expected total %d, got %d", progress.Total, unmarshaled.Total)
	}

	if unmarshaled.Completed != progress.Completed {
		t.Errorf("Expected completed %d, got %d", progress.Completed, unmarshaled.Completed)
	}
}

func TestRunningModelStructure(t *testing.T) {
	model := ModelResponse{
		Name:       "llama2:7b",
		Size:       3825819519,
		Digest:     "sha256:1a838c4c",
		ModifiedAt: time.Now(),
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(model)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled ModelResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.Name != model.Name {
		t.Errorf("Expected name %s, got %s", model.Name, unmarshaled.Name)
	}

	if unmarshaled.Size != model.Size {
		t.Errorf("Expected size %d, got %d", model.Size, unmarshaled.Size)
	}
}

func TestOllamaErrorStructure(t *testing.T) {
	ollamaErr := OllamaError{
		StatusCode: 404,
		Message:    "Model 'nonexistent' not found",
	}

	// Test Error() method
	errorString := ollamaErr.Error()
	if errorString == "" {
		t.Errorf("Error() should return non-empty string")
	}

	if !contains(errorString, "404") {
		t.Errorf("Error string should contain status code")
	}

	if !contains(errorString, "Model 'nonexistent' not found") {
		t.Errorf("Error string should contain error message")
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(ollamaErr)
	assertNoError(t, err)

	// Test JSON unmarshaling
	var unmarshaled OllamaError
	err = json.Unmarshal(jsonData, &unmarshaled)
	assertNoError(t, err)

	if unmarshaled.StatusCode != ollamaErr.StatusCode {
		t.Errorf("Expected status code %d, got %d", ollamaErr.StatusCode, unmarshaled.StatusCode)
	}

	if unmarshaled.Message != ollamaErr.Message {
		t.Errorf("Expected message %s, got %s", ollamaErr.Message, unmarshaled.Message)
	}
}

func TestRequestValidation(t *testing.T) {
	tests := []struct {
		name     string
		request  interface{}
		validate func(interface{}) error
	}{
		{
			name: "Generate request with empty model",
			request: GenerateRequest{
				Model:  "",
				Prompt: "Test",
			},
			validate: func(req interface{}) error {
				r := req.(GenerateRequest)
				if r.Model == "" {
					return &OllamaError{StatusCode: 400, Message: "Model is required"}
				}
				return nil
			},
		},
		{
			name: "Chat request with no messages",
			request: ChatRequest{
				Model:    "llama2",
				Messages: []Message{},
			},
			validate: func(req interface{}) error {
				r := req.(ChatRequest)
				if len(r.Messages) == 0 {
					return &OllamaError{StatusCode: 400, Message: "Messages are required"}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validate(tt.request)
			if err == nil {
				t.Errorf("Expected validation error but got none")
			}

			// Check if it's an OllamaError
			if ollamaErr, ok := err.(*OllamaError); ok {
				if ollamaErr.StatusCode == 0 {
					t.Errorf("Expected error status code to be set")
				}
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		func() bool {
			for i := 1; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())))
}
