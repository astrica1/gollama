# gollama

[![Go Reference](https://pkg.go.dev/badge/github.com/astrica1/gollama.svg)](https://pkg.go.dev/github.com/astrica1/gollama)
![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)

Gollama is a robust, idiomatic Go client library for the [Ollama](https://ollama.ai/) API. It provides complete, type-safe access to all Ollama endpoints, including model management, text generation, chat, embeddings, and process status.

---

## Features

- **Full API Coverage**: All Ollama endpoints implemented
- **Streaming Support**: Real-time streaming for generation, chat, and model operations
- **Type Safety**: Comprehensive Go structs for all requests and responses
- **Context-Aware**: All methods accept `context.Context`
- **Custom Error Handling**: Detailed API errors with status codes
- **Idiomatic Go**: Follows Go best practices and conventions
- **Comprehensive Tests**: 1000+ lines of test coverage

---

## Installation

```sh
go get github.com/astrica1/gollama
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/astrica1/gollama"
)

func main() {
    client, err := gollama.NewClient() // Defaults to http://localhost:11434
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    req := &gollama.GenerateRequest{
        Model:  "llama2",
        Prompt: "Why is the sky blue?",
    }
    resp, err := client.Generate(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Response:", resp.Response)
}
```

---

## API Coverage

This library provides Go types and methods for all Ollama API endpoints:

- **Model Management**: `/api/tags`, `/api/show`, `/api/copy`, `/api/delete`, `/api/pull`, `/api/create`, `/api/push`
- **Text Generation**: `/api/generate`
- **Chat**: `/api/chat`
- **Embeddings**: `/api/embeddings`
- **Process Status**: `/api/ps`

### Available Methods

#### Model Management

- `List(ctx context.Context) (*ListModelsResponse, error)`
- `Show(ctx context.Context, modelName string) (*ModelResponse, error)`
- `Copy(ctx context.Context, source, destination string) error`
- `Delete(ctx context.Context, modelName string) error`
- `Pull(ctx context.Context, modelName string, fn func(PullProgress)) error`
- `Create(ctx context.Context, modelName, modelfileContent string, fn func(CreateProgress)) error`
- `Push(ctx context.Context, modelName string, fn func(PushProgress)) error`

#### Text Generation

- `Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)`
- `GenerateStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse)) error`

#### Chat

- `Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)`
- `ChatStream(ctx context.Context, req *ChatRequest, fn func(*ChatResponse)) error`

#### Embeddings

- `Embeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)`

#### Process Status

- `PS(ctx context.Context) (*PSResponse, error)`

---

## Usage Examples

### Model Management

```go
ctx := context.Background()
models, err := client.List(ctx)
model, err := client.Show(ctx, "llama2")
err = client.Copy(ctx, "llama2", "llama2-backup")
err = client.Delete(ctx, "old-model")
err = client.Pull(ctx, "llama2", func(progress gollama.PullProgress) {
    fmt.Printf("Progress: %s\n", progress.Status)
})
modelfile := "FROM llama2\nSYSTEM You are a helpful assistant."
err = client.Create(ctx, "my-model", modelfile, func(progress gollama.CreateProgress) {
    fmt.Printf("Creating: %s\n", progress.Status)
})
err = client.Push(ctx, "my-model", func(progress gollama.PushProgress) {
    fmt.Printf("Pushing: %s\n", progress.Status)
})
```

### Text Generation

```go
req := &gollama.GenerateRequest{
    Model:  "llama2",
    Prompt: "Tell me a joke",
    Options: map[string]interface{}{
        "temperature": 0.7,
        "top_p":      0.9,
    },
}
resp, err := client.Generate(ctx, req)
err = client.GenerateStream(ctx, req, func(resp *gollama.GenerateResponse) {
    fmt.Print(resp.Response)
})
```

### Chat

```go
chatReq := &gollama.ChatRequest{
    Model: "llama2",
    Messages: []gollama.Message{
        {Role: "user", Content: "Hello!"},
    },
}
chatResp, err := client.Chat(ctx, chatReq)
err = client.ChatStream(ctx, chatReq, func(resp *gollama.ChatResponse) {
    fmt.Print(resp.Message.Content)
})
```

### Embeddings

```go
embReq := &gollama.EmbeddingRequest{
    Model:  "llama2",
    Prompt: "Hello world",
}
embResp, err := client.Embeddings(ctx, embReq)
fmt.Printf("Embedding vector length: %d\n", len(embResp.Embedding))
```

### Process Status

```go
status, err := client.PS(ctx)
fmt.Printf("Running models: %d\n", len(status.Models))
```

---

## Data Structures

- `Client` - Main API client
- `Message` - Chat messages
- `ModelResponse` / `ListModelsResponse` - Model information
- `GenerateRequest` / `GenerateResponse` - Text generation
- `ChatRequest` / `ChatResponse` - Chat completions
- `EmbeddingRequest` / `EmbeddingResponse` - Vector embeddings
- `CreateRequest` / `CreateProgress` - Model creation
- `PushRequest` / `PushProgress` - Model publishing
- `PSResponse` - Process status
- `OllamaError` - Custom error type

---

## Error Handling

All errors returned by the client implement the `error` interface. API errors are of type `*OllamaError` and include HTTP status codes and messages.

```go
resp, err := client.Generate(ctx, req)
if err != nil {
    if ollamaErr, ok := err.(*gollama.OllamaError); ok {
        fmt.Printf("Ollama API error (status %d): %s\n", ollamaErr.StatusCode, ollamaErr.Message)
    } else {
        fmt.Printf("Other error: %v\n", err)
    }
}
```

---

## Examples

See the [`examples/`](examples/) directory for complete working examples:

- [Basic Usage](examples/basic/main.go)
- [Model Management](examples/model-management/main.go)
- [Text Generation & Chat](examples/generation/main.go)
- [Complete API Demo](examples/complete-api/main.go)

---

## Requirements

- Go 1.21 or later
- Access to an Ollama server (local or remote)

---

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

---

## License

MIT License - see [LICENSE](LICENSE)

## Project Status

🚧 **Foundation Complete** - This library currently provides:

- ✅ Complete client structure with HTTP client
- ✅ All necessary data types and structs
- ✅ Proper JSON serialization/deserialization
- ✅ Custom error handling
- ✅ Comprehensive test suite
- ✅ Usage examples

🔜 **Coming Next** - API method implementations for:

- ~~Model listing and management~~ ✅ **COMPLETED**
- Text generation (streaming and non-streaming)
- Chat completions (streaming and non-streaming)  
- Embedding generation

## Model Management Usage

The library now includes complete model management functionality:

```go
ctx := context.Background()

// List all available models
models, err := client.List(ctx)
if err != nil {
    log.Fatal(err)
}

// Show details for a specific model
model, err := client.Show(ctx, "llama2")
if err != nil {
    log.Fatal(err)
}

// Copy a model
err = client.Copy(ctx, "llama2", "llama2-backup")
if err != nil {
    log.Fatal(err)
}

// Pull a model with progress tracking
err = client.Pull(ctx, "llama2", func(progress gollama.PullProgress) {
    if progress.Total > 0 {
        percent := float64(progress.Completed) / float64(progress.Total) * 100
        fmt.Printf("Progress: %.1f%% - %s\n", percent, progress.Status)
    } else {
        fmt.Printf("Status: %s\n", progress.Status)
    }
})

// Delete a model
err = client.Delete(ctx, "old-model")
if err != nil {
    log.Fatal(err)
}
```

## Data Structures

The library includes complete Go structs for all Ollama API interactions:

- `Client` - Main API client
- `Message` - Chat messages
- `ModelResponse` / `ListModelsResponse` - Model information
- `GenerateRequest` / `GenerateResponse` - Text generation
- `ChatRequest` / `ChatResponse` - Chat completions
- `EmbeddingRequest` / `EmbeddingResponse` - Vector embeddings
- `OllamaError` - Custom error type

See [doc.go](doc.go) for complete API documentation.

## Examples

Check out the [examples](examples/) directory for usage examples:

- [Basic Usage](examples/basic/main.go) - Client initialization and data structures
- [Model Management](examples/model-management/main.go) - Complete model management demo

## Development

```bash
# Run tests
go test ./...

# Run examples
go run examples/basic/main.go

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

## Requirements

- Go 1.21 or later
- Access to an Ollama server (local or remote)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## Acknowledgments

- [Ollama](https://ollama.ai/) for providing the excellent local LLM platform
- The Go community for best practices and conventions
