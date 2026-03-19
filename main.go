package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yourusername/OpenSack/orchestrator"
)

func main() {
	// Get API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY environment variable not set")
		fmt.Println("\nSet it with:")
		fmt.Println("  export ANTHROPIC_API_KEY='your-api-key'")
		os.Exit(1)
	}

	// Configure orchestrator
	config := orchestrator.Config{
		AnthropicAPIKey:    apiKey,
		AutoMode:           true,
		ConsensusThreshold: 6, // Score >= 6 triggers consensus mode
		AllowUserOverride:  true,
		AlwaysShowAnalysis: true,
		MonthlyBudget:      300.0,
	}

	orch := orchestrator.NewOrchestrator(config)

	// Get user goal
	if len(os.Args) < 2 {
		fmt.Println("OpenSack")
		fmt.Println("==================")
		fmt.Println("\nUsage: opensack \"your goal here\"")
		fmt.Println("\nExample:")
		fmt.Println("  opensack \"Add authentication to my API\"")
		fmt.Println("  opensack \"Build a chat application with WebSockets\"")
		os.Exit(1)
	}

	goal := os.Args[1]

	// Run orchestrator
	ctx := context.Background()
	if err := orch.Execute(ctx, goal); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Complete!")
}
