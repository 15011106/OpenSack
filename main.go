package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yourusername/OpenSack/orchestrator"
)

func main() {
	// Check for Bedrock mode first (Option 3)
	useBedrock := os.Getenv("CLAUDE_CODE_USE_BEDROCK") == "1"
	bedrockToken := os.Getenv("AWS_BEARER_TOKEN_BEDROCK")

	var apiKey string
	var provider string

	if useBedrock && bedrockToken != "" {
		// Option 3: Using Bedrock
		provider = "bedrock"
		apiKey = bedrockToken
		fmt.Println("OpenSack (Bedrock Mode)")
		fmt.Println("==================")
	} else {
		// Option 1 or 2: Using Anthropic
		provider = "anthropic"
		apiKey = os.Getenv("ANTHROPIC_API_KEY")

		if apiKey == "" {
			// Option 2: No env var - expect CLI arg
			if len(os.Args) < 2 {
				fmt.Println("OpenSack")
				fmt.Println("==================")
				fmt.Println("\nUsage:")
				fmt.Println("  Option 1 (recommended): Set environment variable")
				fmt.Println("    export ANTHROPIC_API_KEY='your-api-key'")
				fmt.Println("    opensack")
				fmt.Println("\n  Option 2: Provide API key as argument")
				fmt.Println("    opensack 'your-api-key'")
				fmt.Println("\n  Option 3: Use AWS Bedrock")
				fmt.Println("    export CLAUDE_CODE_USE_BEDROCK=1")
				fmt.Println("    export AWS_BEARER_TOKEN_BEDROCK='your-bearer-token'")
				fmt.Println("    opensack")
				os.Exit(1)
			}
			apiKey = os.Args[1]
			fmt.Println("⚠️  Warning: API key provided via command line (will appear in shell history)")
			fmt.Println("   Consider using: export ANTHROPIC_API_KEY='...'")
		}
		fmt.Println("OpenSack")
		fmt.Println("==================")
	}

	// Prompt for goal after credentials are set
	fmt.Println("\nWhat would you like to build?")
	fmt.Print("> ")

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		fmt.Println("\nError: No goal provided")
		os.Exit(1)
	}

	goal := strings.TrimSpace(scanner.Text())
	if goal == "" {
		fmt.Println("Error: Goal cannot be empty")
		os.Exit(1)
	}

	fmt.Println()

	// Configure orchestrator
	config := orchestrator.Config{
		APIKey:             apiKey,
		Provider:           provider,
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
