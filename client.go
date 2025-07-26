// Package gollama provides a client for interacting with the Ollama API.
// It allows for operations such as generating text, chatting with models,
// and retrieving embeddings.
package gollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents an Ollama API client. It holds the HTTP client
// used for requests and the base URL of the Ollama server.
type Client struct {
	// httpClient is the underlying HTTP client used for making requests
	httpClient *http.Client
	// baseURL is the base URL of the Ollama server
	baseURL string
}

// NewClient creates a new Ollama API client.
//
// It accepts optional host URL as a parameter. If no host is provided or an empty string
// is given, it defaults to "http://localhost:11434".
//
// Examples:
//   client, err := gollama.NewClient()                           // Uses default localhost:11434
//   client, err := gollama.NewClient("http://192.168.1.100:11434") // Custom host
//
// It returns a pointer to a `Client` and an error if the client cannot be initialized.
func NewClient(host ...string) (*Client, error) {
	baseURL := "http://localhost:11434"

	if len(host) > 0 && host[0] != "" {
		baseURL = host[0]
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}, nil
}

// BaseURL returns the base URL of the Ollama server that the client is configured to use.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// do is an internal helper method for making HTTP requests to the Ollama API.
// It handles request creation, JSON serialization/deserialization, and error parsing.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - method: HTTP method (GET, POST, DELETE, etc.)
//   - path: API endpoint path (e.g., "/api/tags")
//   - reqBody: Request body to be JSON-serialized (can be nil)
//   - resBody: Response body to deserialize JSON into (can be nil)
//
// Returns an error if the request fails or the response indicates an error.
func (c *Client) do(ctx context.Context, method, path string, reqBody, resBody interface{}) error {
	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	var body io.Reader
	if reqBody != nil {
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Deserialize response body if a target is provided
	if resBody != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, resBody); err != nil {
			return fmt.Errorf("failed to unmarshal response body: %w", err)
		}
	}

	return nil
}

// List retrieves all available models from the Ollama server.
// It makes a GET request to the `/api/tags` endpoint.
//
// Returns a ListModelsResponse containing information about all available models,
// or an error if the request fails.
func (c *Client) List(ctx context.Context) (*ListModelsResponse, error) {
	var response ListModelsResponse
	err := c.do(ctx, http.MethodGet, "/api/tags", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}
	return &response, nil
}

// Show retrieves detailed information about a specific model.
// It makes a POST request to the `/api/show` endpoint with the model name.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - modelName: The name of the model to show details for
//
// Returns a ModelResponse with detailed model information, or an error if the request fails.
func (c *Client) Show(ctx context.Context, modelName string) (*ModelResponse, error) {
	if modelName == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	req := ShowRequest{Model: modelName}
	var response ModelResponse
	err := c.do(ctx, http.MethodPost, "/api/show", req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to show model %q: %w", modelName, err)
	}
	return &response, nil
}

// Copy creates a copy of an existing model with a new name.
// It makes a POST request to the `/api/copy` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - source: The name of the source model to copy
//   - destination: The name for the new copied model
//
// Returns an error if the copy operation fails.
func (c *Client) Copy(ctx context.Context, source, destination string) error {
	if source == "" {
		return fmt.Errorf("source model name cannot be empty")
	}
	if destination == "" {
		return fmt.Errorf("destination model name cannot be empty")
	}

	req := CopyRequest{Source: source, Destination: destination}
	err := c.do(ctx, http.MethodPost, "/api/copy", req, nil)
	if err != nil {
		return fmt.Errorf("failed to copy model from %q to %q: %w", source, destination, err)
	}
	return nil
}

// Delete removes a model from the Ollama server.
// It makes a DELETE request to the `/api/delete` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - modelName: The name of the model to delete
//
// Returns an error if the deletion fails.
func (c *Client) Delete(ctx context.Context, modelName string) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}

	req := DeleteRequest{Model: modelName}
	err := c.do(ctx, http.MethodDelete, "/api/delete", req, nil)
	if err != nil {
		return fmt.Errorf("failed to delete model %q: %w", modelName, err)
	}
	return nil
}

// Pull downloads a model from the Ollama model registry with streaming progress updates.
// It makes a POST request to the `/api/pull` endpoint and streams the response.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - modelName: The name of the model to pull/download
//   - fn: Callback function that receives progress updates during the pull operation
//
// The callback function is called for each progress update received from the server.
// Returns an error if the pull operation fails.
func (c *Client) Pull(ctx context.Context, modelName string, fn func(PullProgress)) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("progress callback function cannot be nil")
	}

	req := PullRequest{Model: modelName}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, "/api/pull")
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute pull request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("pull request failed with status %d and could not read response body: %w", resp.StatusCode, readErr)
		}
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Stream the response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var progress PullProgress
		if err := json.Unmarshal([]byte(line), &progress); err != nil {
			// Log the error but continue processing other lines
			continue
		}

		// Call the callback function with the progress update
		fn(progress)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading pull response stream: %w", err)
	}

	return nil
}

// Create creates a new model from a Modelfile with streaming progress updates.
// It makes a POST request to the `/api/create` endpoint and streams the response.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - modelName: The name for the new model to create
//   - modelfileContent: The content of the Modelfile defining the model
//   - fn: Callback function that receives progress updates during the creation operation
//
// The callback function is called for each progress update received from the server.
// Returns an error if the create operation fails.
func (c *Client) Create(ctx context.Context, modelName, modelfileContent string, fn func(CreateProgress)) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if modelfileContent == "" {
		return fmt.Errorf("modelfile content cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("progress callback function cannot be nil")
	}

	req := CreateRequest{Model: modelName, Modelfile: modelfileContent}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal create request: %w", err)
	}

	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, "/api/create")
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute create request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("create request failed with status %d and could not read response body: %w", resp.StatusCode, readErr)
		}
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Stream the response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var progress CreateProgress
		if err := json.Unmarshal([]byte(line), &progress); err != nil {
			// Log the error but continue processing other lines
			continue
		}

		// Call the callback function with the progress update
		fn(progress)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading create response stream: %w", err)
	}

	return nil
}

// Push uploads a model to a registry with streaming progress updates.
// It makes a POST request to the `/api/push` endpoint and streams the response.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - modelName: The name of the model to push to the registry
//   - fn: Callback function that receives progress updates during the push operation
//
// The callback function is called for each progress update received from the server.
// Returns an error if the push operation fails.
func (c *Client) Push(ctx context.Context, modelName string, fn func(PushProgress)) error {
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("progress callback function cannot be nil")
	}

	req := PushRequest{Model: modelName}
	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal push request: %w", err)
	}

	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, "/api/push")
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute push request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("push request failed with status %d and could not read response body: %w", resp.StatusCode, readErr)
		}
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Stream the response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var progress PushProgress
		if err := json.Unmarshal([]byte(line), &progress); err != nil {
			// Log the error but continue processing other lines
			continue
		}

		// Call the callback function with the progress update
		fn(progress)
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading push response stream: %w", err)
	}

	return nil
}

// Generate performs text generation using the specified model and prompt.
// This method handles non-streaming requests where the complete response is returned at once.
// It makes a POST request to the `/api/generate` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: The generation request containing model, prompt, and options
//
// Returns a GenerateResponse with the generated text and metadata, or an error if the request fails.
func (c *Client) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("generate request cannot be nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}

	// Ensure this is a non-streaming request
	reqCopy := *req
	reqCopy.Stream = false

	var response GenerateResponse
	err := c.do(ctx, http.MethodPost, "/api/generate", &reqCopy, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to generate text: %w", err)
	}
	return &response, nil
}

// GenerateStream performs streaming text generation using the specified model and prompt.
// This method handles streaming requests where partial responses are delivered via callback.
// It makes a POST request to the `/api/generate` endpoint with streaming enabled.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: The generation request containing model, prompt, and options
//   - fn: Callback function that receives each partial response during generation
//
// The callback function is called for each partial response received from the server.
// Returns an error if the generation fails or if the request/callback parameters are invalid.
func (c *Client) GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse)) error {
	if req == nil {
		return fmt.Errorf("generate request cannot be nil")
	}
	if req.Model == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if fn == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	// Ensure this is a streaming request
	reqCopy := *req
	reqCopy.Stream = true

	jsonData, err := json.Marshal(&reqCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal generate request: %w", err)
	}

	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, "/api/generate")
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute generate request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("generate request failed with status %d and could not read response body: %w", resp.StatusCode, readErr)
		}
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Stream the response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Check if context was canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var response GenerateResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			// Log the error but continue processing other lines
			continue
		}

		// Call the callback function with the response
		fn(&response)

		// Check if generation is complete
		if response.Done {
			break
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading generate response stream: %w", err)
	}

	return nil
}

// Chat performs a chat conversation using the specified model and message history.
// This method handles non-streaming requests where the complete response is returned at once.
// It makes a POST request to the `/api/chat` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: The chat request containing model, messages, and options
//
// Returns a ChatResponse with the assistant's message and metadata, or an error if the request fails.
func (c *Client) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("chat request cannot be nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("at least one message is required")
	}

	// Ensure this is a non-streaming request
	reqCopy := *req
	reqCopy.Stream = false

	var response ChatResponse
	err := c.do(ctx, http.MethodPost, "/api/chat", &reqCopy, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to chat: %w", err)
	}
	return &response, nil
}

// ChatStream performs a streaming chat conversation using the specified model and message history.
// This method handles streaming requests where partial responses are delivered via callback.
// It makes a POST request to the `/api/chat` endpoint with streaming enabled.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: The chat request containing model, messages, and options
//   - fn: Callback function that receives each partial response during the conversation
//
// The callback function is called for each partial response received from the server.
// Returns an error if the chat fails or if the request/callback parameters are invalid.
func (c *Client) ChatStream(ctx context.Context, req *ChatRequest, fn func(*ChatResponse)) error {
	if req == nil {
		return fmt.Errorf("chat request cannot be nil")
	}
	if req.Model == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	if len(req.Messages) == 0 {
		return fmt.Errorf("at least one message is required")
	}
	if fn == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	// Ensure this is a streaming request
	reqCopy := *req
	reqCopy.Stream = true

	jsonData, err := json.Marshal(&reqCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal chat request: %w", err)
	}

	// Construct the full URL
	u, err := url.JoinPath(c.baseURL, "/api/chat")
	if err != nil {
		return fmt.Errorf("failed to construct URL: %w", err)
	}

	// Create the HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	// Execute the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to execute chat request: %w", err)
	}
	defer resp.Body.Close()

	// Check for non-2xx status codes
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("chat request failed with status %d and could not read response body: %w", resp.StatusCode, readErr)
		}
		return parseErrorResponse(resp.StatusCode, respBody)
	}

	// Stream the response line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// Check if context was canceled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var response ChatResponse
		if err := json.Unmarshal([]byte(line), &response); err != nil {
			// Log the error but continue processing other lines
			continue
		}

		// Call the callback function with the response
		fn(&response)

		// Check if conversation is complete
		if response.Done {
			break
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading chat response stream: %w", err)
	}

	return nil
}

// Embeddings generates vector embeddings for the given text using the specified model.
// It makes a POST request to the `/api/embeddings` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//   - req: The embedding request containing model and text to embed
//
// Returns an EmbeddingResponse containing the generated embedding vector, or an error if the request fails.
func (c *Client) Embeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("embedding request cannot be nil")
	}
	if req.Model == "" {
		return nil, fmt.Errorf("model name cannot be empty")
	}
	if req.Prompt == "" {
		return nil, fmt.Errorf("prompt cannot be empty")
	}

	var response EmbeddingResponse
	err := c.do(ctx, http.MethodPost, "/api/embeddings", req, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings: %w", err)
	}
	return &response, nil
}

// PS retrieves information about currently running models and processes.
// It makes a GET request to the `/api/ps` endpoint.
//
// Parameters:
//   - ctx: Context for request cancellation and timeouts
//
// Returns a PSResponse containing information about running models, or an error if the request fails.
func (c *Client) PS(ctx context.Context) (*PSResponse, error) {
	var response PSResponse
	err := c.do(ctx, http.MethodGet, "/api/ps", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to get process status: %w", err)
	}
	return &response, nil
}

// Message represents a single chat message, comprising a role (e.g., "user", "assistant")
// and the content of the message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelDetails contains specific metadata about an Ollama model, such as
// its parameter size, quantization level, and family.
type ModelDetails struct {
	ParameterSize     string `json:"parameter_size,omitempty"`
	QuantizationLevel string `json:"quantization_level,omitempty"`
	Family            string `json:"family,omitempty"`
	ParentModel       string `json:"parent_model,omitempty"`
	Format            string `json:"format,omitempty"`
}

// ModelResponse represents the detailed information for a single model
// returned by the Ollama API's list models endpoint.
type ModelResponse struct {
	Name       string       `json:"name"`
	ModifiedAt time.Time    `json:"modified_at"`
	Size       int64        `json:"size"`
	Digest     string       `json:"digest"`
	Details    ModelDetails `json:"details,omitempty"`
}

// ListModelsResponse encapsulates the response structure for listing
// all available models from the Ollama API.
type ListModelsResponse struct {
	Models []ModelResponse `json:"models"`
}

// GenerateRequest defines the structure for a request to the Ollama API's
// `/api/generate` endpoint, used for generating text completions.
type GenerateRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream,omitempty"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// GenerateResponse represents the response structure from the Ollama API's
// `/api/generate` endpoint. It includes the generated text, model information,
// and performance metrics.
type GenerateResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// ChatRequest defines the structure for a request to the Ollama API's
// `/api/chat` endpoint, used for multi-turn conversations with models.
type ChatRequest struct {
	Model    string                 `json:"model"`
	Messages []Message              `json:"messages"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ChatResponse represents the response structure from the Ollama API's
// `/api/chat` endpoint. It contains the model's message, along with
// creation time and performance statistics.
type ChatResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// EmbeddingRequest defines the structure for a request to the Ollama API's
// `/api/embeddings` endpoint, used for generating vector embeddings of text.
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response structure from the Ollama API's
// `/api/embeddings` endpoint, containing the generated embedding as a slice of float64.
type EmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// ShowRequest defines the structure for a request to show model details.
type ShowRequest struct {
	Model string `json:"model"`
}

// CopyRequest defines the structure for copying a model.
type CopyRequest struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
}

// DeleteRequest defines the structure for deleting a model.
type DeleteRequest struct {
	Model string `json:"model"`
}

// PullRequest defines the structure for pulling a model.
type PullRequest struct {
	Model string `json:"model"`
}

// PullProgress represents the progress information during model pulling.
type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// CreateRequest defines the structure for creating a model from a Modelfile.
type CreateRequest struct {
	Model     string `json:"name"`
	Modelfile string `json:"modelfile"`
}

// CreateProgress represents the progress information during model creation.
type CreateProgress struct {
	Status string `json:"status"`
}

// PushRequest defines the structure for pushing a model to a registry.
type PushRequest struct {
	Model string `json:"name"`
}

// PushProgress represents the progress information during model pushing.
type PushProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

// PSResponse represents the response from the process status endpoint.
type PSResponse struct {
	Models []ModelResponse `json:"models"`
}

// OllamaError represents a custom error type for errors returned by the Ollama API.
// It includes the HTTP status code and a descriptive message.
type OllamaError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

// Error implements the error interface for OllamaError, providing a formatted
// string representation of the error.
func (e *OllamaError) Error() string {
	return fmt.Sprintf("Ollama API error (status %d): %s", e.StatusCode, e.Message)
}

// ErrorResponse represents the generic error response structure from the Ollama API.
type ErrorResponse struct {
	Error string `json:"error"`
}

// parseErrorResponse attempts to parse a raw byte slice into an OllamaError.
// It takes the HTTP status code and the response body. If the body can be
// unmarshaled into an `ErrorResponse`, its `Error` field is used as the message.
// Otherwise, the raw body content is used as the message.
func parseErrorResponse(statusCode int, body []byte) error {
	var errorResp ErrorResponse
	if err := json.Unmarshal(body, &errorResp); err != nil {
		return &OllamaError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}

	return &OllamaError{
		StatusCode: statusCode,
		Message:    errorResp.Error,
	}
}
