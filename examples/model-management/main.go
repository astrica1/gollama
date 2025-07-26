package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/astrica1/gollama"
)

func main() {
	// Create a new client with default host (localhost:11434)
	client, err := gollama.NewClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	fmt.Printf("Connected to Ollama at: %s\n", client.BaseURL())

	// Create a context with timeout for API calls
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// List all available models
	fmt.Println("\n=== Listing Available Models ===")
	models, err := client.List(ctx)
	if err != nil {
		log.Printf("Failed to list models: %v", err)
		return
	}

	if len(models.Models) == 0 {
		fmt.Println("No models found. You might need to pull a model first.")
		
		// Example of pulling a model
		fmt.Println("\n=== Pulling a Model (llama2:latest) ===")
		fmt.Println("This will download the model if it's not already available...")
		
		err = client.Pull(ctx, "llama2:latest", func(progress gollama.PullProgress) {
			if progress.Total > 0 {
				percent := float64(progress.Completed) / float64(progress.Total) * 100
				fmt.Printf("Progress: %.1f%% - %s\n", percent, progress.Status)
			} else {
				fmt.Printf("Status: %s\n", progress.Status)
			}
		})
		
		if err != nil {
			log.Printf("Failed to pull model: %v", err)
			return
		}
		
		fmt.Println("Model pulled successfully!")
		
		// List models again after pulling
		models, err = client.List(ctx)
		if err != nil {
			log.Printf("Failed to list models after pull: %v", err)
			return
		}
	}

	for i, model := range models.Models {
		fmt.Printf("%d. Model: %s\n", i+1, model.Name)
		fmt.Printf("   Size: %.2f GB\n", float64(model.Size)/(1024*1024*1024))
		fmt.Printf("   Modified: %s\n", model.ModifiedAt.Format(time.RFC3339))
		fmt.Printf("   Digest: %s\n", model.Digest)
		if model.Details.ParameterSize != "" {
			fmt.Printf("   Parameters: %s\n", model.Details.ParameterSize)
		}
		if model.Details.Family != "" {
			fmt.Printf("   Family: %s\n", model.Details.Family)
		}
		fmt.Println()
	}

	// Show details for the first model if available
	if len(models.Models) > 0 {
		modelName := models.Models[0].Name
		fmt.Printf("=== Showing Details for Model: %s ===\n", modelName)
		
		modelDetails, err := client.Show(ctx, modelName)
		if err != nil {
			log.Printf("Failed to show model details: %v", err)
		} else {
			fmt.Printf("Name: %s\n", modelDetails.Name)
			fmt.Printf("Size: %.2f GB\n", float64(modelDetails.Size)/(1024*1024*1024))
			fmt.Printf("Modified: %s\n", modelDetails.ModifiedAt.Format(time.RFC3339))
			fmt.Printf("Digest: %s\n", modelDetails.Digest)
			if modelDetails.Details.ParameterSize != "" {
				fmt.Printf("Parameter Size: %s\n", modelDetails.Details.ParameterSize)
			}
			if modelDetails.Details.QuantizationLevel != "" {
				fmt.Printf("Quantization: %s\n", modelDetails.Details.QuantizationLevel)
			}
			if modelDetails.Details.Family != "" {
				fmt.Printf("Family: %s\n", modelDetails.Details.Family)
			}
		}

		// Example of copying a model
		copyName := modelName + "-backup"
		fmt.Printf("\n=== Copying Model: %s -> %s ===\n", modelName, copyName)
		
		err = client.Copy(ctx, modelName, copyName)
		if err != nil {
			log.Printf("Failed to copy model: %v", err)
		} else {
			fmt.Printf("Successfully copied %s to %s\n", modelName, copyName)
			
			// List models again to show the copy
			fmt.Println("\n=== Updated Model List ===")
			updatedModels, err := client.List(ctx)
			if err != nil {
				log.Printf("Failed to list models: %v", err)
			} else {
				for i, model := range updatedModels.Models {
					fmt.Printf("%d. %s (%.2f GB)\n", i+1, model.Name, float64(model.Size)/(1024*1024*1024))
				}
			}
			
			// Clean up: delete the copied model
			fmt.Printf("\n=== Cleaning Up: Deleting %s ===\n", copyName)
			err = client.Delete(ctx, copyName)
			if err != nil {
				log.Printf("Failed to delete copied model: %v", err)
			} else {
				fmt.Printf("Successfully deleted %s\n", copyName)
			}
		}
	}

	fmt.Println("\n=== Model Management Demo Complete ===")
}
