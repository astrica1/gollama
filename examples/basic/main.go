package main

import (
	"fmt"
	"log"

	"github.com/astrica1/gollama"
)

func main() {
	// Create a new client with default host (localhost:11434)
	client, err := gollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Created Ollama client with base URL: %s\n", client.BaseURL())

	// Example of creating a custom host client
	customHost := "http://127.0.0.1:11434"
	customClient, err := gollama.NewClient(customHost)
	if err != nil {
		log.Fatalf("Failed to create custom client: %v", err)
	}

	fmt.Printf("Created custom Ollama client with base URL: %s\n", customClient.BaseURL())

	// Example of model management operations (these would work with a running Ollama server)
	fmt.Println("\n--- Model Management API Examples ---")
	fmt.Println("Note: These examples require a running Ollama server")
	
	// Example: List models
	fmt.Println("\n1. Listing models:")
	fmt.Println("   ctx := context.Background()")
	fmt.Println("   models, err := client.List(ctx)")
	fmt.Println("   if err != nil { /* handle error */ }")
	fmt.Println("   for _, model := range models.Models {")
	fmt.Println("       fmt.Printf(\"Model: %s, Size: %d\\n\", model.Name, model.Size)")
	fmt.Println("   }")

	// Example: Show model details
	fmt.Println("\n2. Showing model details:")
	fmt.Println("   model, err := client.Show(ctx, \"llama2\")")
	fmt.Println("   if err != nil { /* handle error */ }")
	fmt.Println("   fmt.Printf(\"Model: %s, Family: %s\\n\", model.Name, model.Details.Family)")

	// Example: Copy a model
	fmt.Println("\n3. Copying a model:")
	fmt.Println("   err := client.Copy(ctx, \"llama2\", \"llama2-backup\")")
	fmt.Println("   if err != nil { /* handle error */ }")

	// Example: Pull a model with progress
	fmt.Println("\n4. Pulling a model with progress:")
	fmt.Println("   err := client.Pull(ctx, \"llama2\", func(progress gollama.PullProgress) {")
	fmt.Println("       if progress.Total > 0 {")
	fmt.Println("           percent := float64(progress.Completed) / float64(progress.Total) * 100")
	fmt.Println("           fmt.Printf(\"Progress: %.1f%% (%s)\\n\", percent, progress.Status)")
	fmt.Println("       } else {")
	fmt.Println("           fmt.Printf(\"Status: %s\\n\", progress.Status)")
	fmt.Println("       }")
	fmt.Println("   })")

	// Example: Delete a model
	fmt.Println("\n5. Deleting a model:")
	fmt.Println("   err := client.Delete(ctx, \"old-model\")")
	fmt.Println("   if err != nil { /* handle error */ }")

	// Example of data structures that will be used with the API
	fmt.Println("\n--- Data Structure Examples ---")
	
	// Chat message example
	message := gollama.Message{
		Role:    "user",
		Content: "Hello, how are you?",
	}
	fmt.Printf("Message: %+v\n", message)

	// Generate request example
	generateReq := gollama.GenerateRequest{
		Model:  "llama2",
		Prompt: "Tell me a joke",
		Stream: false,
	}
	fmt.Printf("Generate Request: %+v\n", generateReq)

	// Chat request example
	chatReq := gollama.ChatRequest{
		Model: "llama2",
		Messages: []gollama.Message{
			{Role: "user", Content: "Hello!"},
			{Role: "assistant", Content: "Hi there! How can I help you?"},
			{Role: "user", Content: "What's the weather like?"},
		},
		Stream: false,
	}
	fmt.Printf("Chat Request: %+v\n", chatReq)

	// Embedding request example
	embeddingReq := gollama.EmbeddingRequest{
		Model:  "llama2",
		Prompt: "The quick brown fox jumps over the lazy dog",
	}
	fmt.Printf("Embedding Request: %+v\n", embeddingReq)

	// Pull progress example
	pullProgress := gollama.PullProgress{
		Status:    "downloading",
		Digest:    "sha256:abc123def456",
		Total:     1024000,
		Completed: 512000,
	}
	fmt.Printf("Pull Progress: %+v\n", pullProgress)

	fmt.Println("\n--- Summary ---")
	fmt.Println("The gollama client library provides:")
	fmt.Println("✅ Complete model management (List, Show, Copy, Delete, Pull)")
	fmt.Println("✅ All data structures for Ollama API")
	fmt.Println("✅ Context support for timeouts and cancellation")
	fmt.Println("✅ Comprehensive error handling")
	fmt.Println("✅ Streaming support for model pulling")
	fmt.Println("\nFor a working example with actual API calls, see:")
	fmt.Println("examples/model-management/main.go")
}
