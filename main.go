package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yourusername/OpenSack/orchestrator"
)

func main() {
	// Get API key from environment or command line
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	var goal string

	if apiKey == "" {
		// No env var - expect: opensack 'api-key' "goal"
		if len(os.Args) < 3 {
			fmt.Println("OpenSack")
			fmt.Println("==================")
			fmt.Println("\nUsage:")
			fmt.Println("  Option 1 (recommended): Set environment variable")
			fmt.Println("    export ANTHROPIC_API_KEY='your-api-key'")
			fmt.Println("    opensack \"your goal here\"")
			fmt.Println("\n  Option 2: Provide API key as first argument")
			fmt.Println("    opensack 'your-api-key' \"your goal here\"")
			fmt.Println("\nExample:")
			fmt.Println("  opensack \"Add authentication to my API\"")
			fmt.Println("  opensack 'sk-ant-...' \"Build a chat application\"")
			os.Exit(1)
		}
		apiKey = os.Args[1]
		goal = os.Args[2]
		fmt.Println("⚠️  Warning: API key provided via command line (will appear in shell history)")
		fmt.Println("   Consider using: export ANTHROPIC_API_KEY='...'")
		fmt.Println()
	} else {
		// Env var exists - expect: opensack "goal"
		if len(os.Args) < 2 {
			fmt.Println("OpenSack")
			fmt.Println("==================")
			fmt.Println("\nUsage: opensack \"your goal here\"")
			fmt.Println("\nExample:")
			fmt.Println("  opensack \"Add authentication to my API\"")
			fmt.Println("  opensack \"Build a chat application with WebSockets\"")
			os.Exit(1)
		}
		goal = os.Args[1]
	}

	// Configure orchestrator
	config := orchestrator.Config{
		AnthropicAPIKey:    apiKey,
		AutoMode:           true,
		ConsensusThreshold: 6,
		AllowUserOverride:  true,
		AlwaysShowAnalysis: true,
		MonthlyBudget:      300.0,
	}

	orch := orchestrator.NewOrchestrator(config)

	// Run orchestrator
	ctx := context.Background()
	if err := orch.Execute(ctx, goal); err != nil {
		fmt.Printf("\nError: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Complete!")
}
