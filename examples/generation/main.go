package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/astrica1/gollama"
)

func main() {
	// Create a new client
	client, err := gollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Connected to Ollama at: %s\n", client.BaseURL())

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Note: These examples require a running Ollama server with models
	fmt.Println("\n=== Text Generation Examples ===")
	fmt.Println("Note: These examples require a running Ollama server with the 'llama2' model")

	// Example 1: Simple text generation (non-streaming)
	fmt.Println("\n--- 1. Simple Text Generation (Non-streaming) ---")
	generateReq := &gollama.GenerateRequest{
		Model:  "llama2",
		Prompt: "Explain quantum computing in simple terms:",
	}

	fmt.Println("Making generate request...")
	response, err := client.Generate(ctx, generateReq)
	if err != nil {
		log.Printf("Generate failed: %v", err)
		fmt.Println("(This is expected if Ollama server is not running)")
	} else {
		fmt.Printf("Generated text: %s\n", response.Response)
		fmt.Printf("Model used: %s\n", response.Model)
		if response.TotalDuration > 0 {
			fmt.Printf("Total duration: %.2f seconds\n", float64(response.TotalDuration)/1e9)
		}
	}

	// Example 2: Streaming text generation
	fmt.Println("\n--- 2. Streaming Text Generation ---")
	streamReq := &gollama.GenerateRequest{
		Model:  "llama2",
		Prompt: "Write a short poem about artificial intelligence:",
		Options: map[string]interface{}{
			"temperature": 0.7,
		},
	}

	fmt.Println("Starting streaming generation...")
	var fullText strings.Builder
	err = client.GenerateStream(ctx, streamReq, func(resp *gollama.GenerateResponse) {
		fmt.Print(resp.Response) // Print each chunk as it arrives
		fullText.WriteString(resp.Response)
		
		if resp.Done {
			fmt.Printf("\n[Generation complete - Total duration: %.2f seconds]\n", 
				float64(resp.TotalDuration)/1e9)
		}
	})

	if err != nil {
		log.Printf("GenerateStream failed: %v", err)
		fmt.Println("(This is expected if Ollama server is not running)")
	}

	// Example 3: Chat conversation (non-streaming)
	fmt.Println("\n--- 3. Chat Conversation (Non-streaming) ---")
	chatReq := &gollama.ChatRequest{
		Model: "llama2",
		Messages: []gollama.Message{
			{Role: "user", Content: "Hello! What can you tell me about Go programming language?"},
		},
	}

	fmt.Println("Making chat request...")
	chatResponse, err := client.Chat(ctx, chatReq)
	if err != nil {
		log.Printf("Chat failed: %v", err)
		fmt.Println("(This is expected if Ollama server is not running)")
	} else {
		fmt.Printf("Assistant: %s\n", chatResponse.Message.Content)
		fmt.Printf("Model used: %s\n", chatResponse.Model)
	}

	// Example 4: Multi-turn chat with streaming
	fmt.Println("\n--- 4. Multi-turn Chat Conversation (Streaming) ---")
	
	// Simulate a conversation history
	conversation := []gollama.Message{
		{Role: "user", Content: "What is machine learning?"},
		{Role: "assistant", Content: "Machine learning is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed."},
		{Role: "user", Content: "Can you give me a simple example?"},
	}

	multiChatReq := &gollama.ChatRequest{
		Model:    "llama2",
		Messages: conversation,
		Options: map[string]interface{}{
			"temperature": 0.8,
			"max_tokens":  150,
		},
	}

	fmt.Println("User: Can you give me a simple example?")
	fmt.Print("Assistant: ")
	
	var chatFullText strings.Builder
	err = client.ChatStream(ctx, multiChatReq, func(resp *gollama.ChatResponse) {
		fmt.Print(resp.Message.Content) // Print each chunk as it arrives
		chatFullText.WriteString(resp.Message.Content)
		
		if resp.Done {
			fmt.Printf("\n[Chat complete]\n")
		}
	})

	if err != nil {
		log.Printf("ChatStream failed: %v", err)
		fmt.Println("(This is expected if Ollama server is not running)")
	}

	// Example 5: Generation with custom options
	fmt.Println("\n--- 5. Generation with Custom Options ---")
	customReq := &gollama.GenerateRequest{
		Model:  "llama2",
		Prompt: "List 3 benefits of renewable energy:",
		Options: map[string]interface{}{
			"temperature":     0.5,  // Lower temperature for more focused responses
			"top_p":          0.9,   // Nucleus sampling
			"repeat_penalty": 1.1,   // Avoid repetition
			"seed":           42,    // For reproducible results
		},
	}

	fmt.Println("Making generation request with custom options...")
	customResponse, err := client.Generate(ctx, customReq)
	if err != nil {
		log.Printf("Custom generate failed: %v", err)
		fmt.Println("(This is expected if Ollama server is not running)")
	} else {
		fmt.Printf("Generated response: %s\n", customResponse.Response)
		if customResponse.PromptEvalCount > 0 {
			fmt.Printf("Prompt tokens: %d\n", customResponse.PromptEvalCount)
		}
		if customResponse.EvalCount > 0 {
			fmt.Printf("Response tokens: %d\n", customResponse.EvalCount)
		}
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("Available generation methods:")
	fmt.Println("✅ Generate(ctx, req) - Non-streaming text generation")
	fmt.Println("✅ GenerateStream(ctx, req, callback) - Streaming text generation")
	fmt.Println("✅ Chat(ctx, req) - Non-streaming chat conversation")
	fmt.Println("✅ ChatStream(ctx, req, callback) - Streaming chat conversation")
	fmt.Println("\nFeatures demonstrated:")
	fmt.Println("• Context-based cancellation and timeouts")
	fmt.Println("• Custom generation options (temperature, top_p, etc.)")
	fmt.Println("• Multi-turn conversation support")
	fmt.Println("• Real-time streaming responses")
	fmt.Println("• Performance metrics (duration, token counts)")
}
