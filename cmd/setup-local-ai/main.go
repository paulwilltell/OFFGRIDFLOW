package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/example/offgridflow/internal/ai"
)

func main() {
	var (
		modelName  = flag.String("model", "llama3.2:3b", "Model to download")
		aiURL      = flag.String("url", "http://localhost:11434", "Ollama URL")
		testPrompt = flag.String("test", "", "Test prompt after setup")
	)
	flag.Parse()

	fmt.Println("OffGridFlow Local AI Setup")
	fmt.Println("===========================")

	// Create local AI provider
	provider, err := ai.NewLocalOfflineProvider(ai.LocalOfflineConfig{
		BaseURL: *aiURL,
		Model:   *modelName,
	})
	if err != nil {
		log.Fatalf("Failed to create local AI provider: %v", err)
	}
	defer provider.Close()

	ctx := context.Background()

	// Check if AI engine is available
	fmt.Printf("\nChecking Ollama availability at %s...\n", *aiURL)
	if err := provider.HealthCheck(ctx); err != nil {
		log.Fatalf("Ollama is not available: %v\nPlease ensure Ollama is running.", err)
	}
	fmt.Println("✓ Ollama is running")

	// List available models
	fmt.Println("\nChecking installed models...")
	models, err := provider.ListModels(ctx)
	if err != nil {
		log.Fatalf("Failed to list models: %v", err)
	}

	modelInstalled := false
	for _, m := range models {
		if m == *modelName {
			modelInstalled = true
			break
		}
	}

	if modelInstalled {
		fmt.Printf("✓ Model '%s' is already installed\n", *modelName)
	} else {
		fmt.Printf("\nModel '%s' not found. Downloading...\n", *modelName)
		fmt.Println("This may take several minutes depending on model size.")

		pullCtx, cancel := context.WithTimeout(ctx, 30*time.Minute)
		defer cancel()

		if err := provider.PullModel(pullCtx, *modelName); err != nil {
			log.Fatalf("Failed to download model: %v", err)
		}
		fmt.Printf("✓ Model '%s' downloaded successfully\n", *modelName)
	}

	// Test the model if requested
	if *testPrompt != "" {
		fmt.Printf("\nTesting model with prompt: %s\n", *testPrompt)

		resp, err := provider.Chat(ctx, ai.ChatRequest{
			Prompt: *testPrompt,
		})
		if err != nil {
			log.Fatalf("Test failed: %v", err)
		}

		fmt.Println("\nResponse:")
		fmt.Println("─────────")
		fmt.Println(resp.Output)
		fmt.Println("─────────")
		fmt.Printf("Latency: %v\n", resp.Latency)
	}

	fmt.Println("\n✓ Local AI setup complete!")
	fmt.Println("\nYou can now run OffGridFlow in offline mode.")
	fmt.Printf("The local AI engine will be used automatically when offline.\n")
}
