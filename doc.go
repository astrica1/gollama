// Package gollama provides a Go client library for interacting with the Ollama API.
//
// Gollama is a robust and idiomatic Go client library for working with local
// language models through Ollama. It provides a complete interface for text
// generation, chat completions, embeddings, model management, and process monitoring.
//
// # Quick Start
//
// Create a new client (defaults to http://localhost:11434):
//
//   client, err := gollama.NewClient()
//   if err != nil {
//       log.Fatal(err)
//   }
//
// Or create a client with custom host:
//
//   client, err = gollama.NewClient("http://your-ollama-server:11434")
//   if err != nil {
//       log.Fatal(err)
//   }
//
// # Complete API Coverage
//
// The library provides complete Go structs and methods for all Ollama API endpoints:
//
// Model Management:
//   - List() - List all available models
//   - Show() - Show detailed model information
//   - Copy() - Copy/duplicate a model
//   - Delete() - Remove a model
//   - Pull() - Download models with streaming progress
//   - Create() - Create models from Modelfile with streaming progress
//   - Push() - Upload models to registry with streaming progress
//
// Text Generation:
//   - Generate() - Generate text completions (non-streaming)
//   - GenerateStream() - Generate text with real-time streaming
//
// Chat Conversations:
//   - Chat() - Multi-turn conversations (non-streaming)
//   - ChatStream() - Multi-turn conversations with real-time streaming
//
// Embeddings & Status:
//   - Embeddings() - Generate vector embeddings from text
//   - PS() - Get status of currently running models
//
// # Data Structures
//
// The library provides complete Go structs for all Ollama API interactions:
//
//   - Client: Main API client with HTTP configuration
//   - Message: Chat messages with role and content
//   - ModelResponse/ListModelsResponse: Model information and metadata
//   - GenerateRequest/GenerateResponse: Text generation with options
//   - ChatRequest/ChatResponse: Chat completions with conversation history
//   - EmbeddingRequest/EmbeddingResponse: Vector embeddings
//   - CreateRequest/CreateProgress: Model creation from Modelfile
//   - PushRequest/PushProgress: Model publishing to registries
//   - PSResponse: Running model process status
//   - OllamaError: Custom error type for API errors
//
// # Options
//
// Common options for generation and chat requests:
//
//   - temperature (float): Controls randomness (0.0 to 2.0)
//   - top_p (float): Nucleus sampling threshold (0.0 to 1.0)
//   - top_k (int): Top-K sampling limit
//   - repeat_penalty (float): Penalty for repeated tokens
//   - seed (int): Random seed for deterministic output
//   - num_ctx (int): Context window size
//   - num_predict (int): Maximum tokens to generate
//
// Example:
//
//	request := gollama.GenerateRequest{
//		Model:  "llama2",
//		Prompt: "Tell me a story",
//		Options: map[string]interface{}{
//			"temperature":    0.7,
//			"top_p":         0.9,
//			"repeat_penalty": 1.1,
//			"seed":          42,
//		},
//	}
//
// # Streaming Support
//
// All long-running operations support streaming with callback functions:
//
//	// Streaming text generation
//	err := client.GenerateStream(ctx, req, func(resp *gollama.GenerateResponse) {
//		fmt.Print(resp.Response) // Print each chunk as it arrives
//	})
//
//	// Streaming model download with progress
//	err := client.Pull(ctx, "llama2", func(progress gollama.PullProgress) {
//		if progress.Total > 0 {
//			percent := float64(progress.Completed) / float64(progress.Total) * 100
//			fmt.Printf("Progress: %.1f%%\n", percent)
//		}
//	})
//
// # Error Handling
//
// The library provides a custom OllamaError type that includes HTTP status
// codes and detailed error messages from the Ollama API:
//
//	if err != nil {
//		if ollamaErr, ok := err.(*gollama.OllamaError); ok {
//			fmt.Printf("Ollama API error (status %d): %s\n", 
//				ollamaErr.StatusCode, ollamaErr.Message)
//		}
//	}
//
// # Context Support
//
// All methods accept a context.Context for cancellation and timeout control:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	response, err := client.Generate(ctx, req)
//	if err != nil {
//		// Handle timeout or cancellation
//	}
package gollama
