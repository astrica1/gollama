# Gollama - Go Client Library for Ollama API

![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![License](https://img.shields.io/badge/license-LICENSE-green.svg)

Gollama is a robust and idiomatic Go client library for interacting with the [Ollama](https://ollama.ai/) API. It provides a simple interface for working with local language models through Ollama.

## Features

- 🚀 **Simple API**: Clean and intuitive Go interface
- 📦 **Complete Type Safety**: Fully typed request/response structures
- 🔧 **Flexible Configuration**: Customizable client with sensible defaults
- 🎯 **Idiomatic Go**: Follows Go best practices and conventions
- 📊 **Comprehensive**: Supports all major Ollama API endpoints
- 🧪 **Well Tested**: Comprehensive test suite included

## Installation

```bash
go get github.com/astrica1/gollama
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/astrica1/gollama"
)

func main() {
    // Create a new client (defaults to http://localhost:11434)
    client, err := gollama.NewClient(nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Connected to Ollama at: %s\n", client.BaseURL())
}
```

## API Coverage

This library provides Go types and methods for the following Ollama API endpoints:

- ✅ **Models** (`/api/tags`, `/api/show`, `/api/copy`, `/api/delete`, `/api/pull`) - Complete model management
- 🔜 **Generate** (`/api/generate`) - Text generation
- 🔜 **Chat** (`/api/chat`) - Chat completions
- 🔜 **Embeddings** (`/api/embeddings`) - Vector embeddings

### Available Methods

#### Model Management

- `List(ctx context.Context) (*ListModelsResponse, error)` - List all models
- `Show(ctx context.Context, modelName string) (*ModelResponse, error)` - Show model details
- `Copy(ctx context.Context, source, destination string) error` - Copy a model
- `Delete(ctx context.Context, modelName string) error` - Delete a model
- `Pull(ctx context.Context, modelName string, fn func(PullProgress)) error` - Pull model with progress

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
