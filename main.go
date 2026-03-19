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
	fmt.Println("OpenSack")
	fmt.Println("==================")

	// Ask user which provider to use
	fmt.Println("\nWhich provider would you like to use?")
	fmt.Println("1. Anthropic API")
	fmt.Println("2. AWS Bedrock")
	fmt.Print("\nChoice [1]: ")

	scanner := bufio.NewScanner(os.Stdin)
	var choice string
	if scanner.Scan() {
		choice = strings.TrimSpace(scanner.Text())
	}
	if choice == "" {
		choice = "1"
	}

	var apiKey string
	var provider string

	switch choice {
	case "2":
		// User chose Bedrock
		provider = "bedrock"
		apiKey = os.Getenv("AWS_BEARER_TOKEN_BEDROCK")

		if apiKey == "" {
			fmt.Println("\nError: AWS_BEARER_TOKEN_BEDROCK not set")
			fmt.Println("\nPlease set it:")
			fmt.Println("  export AWS_BEARER_TOKEN_BEDROCK='your-bearer-token'")
			os.Exit(1)
		}
		fmt.Println(apiKey)
		fmt.Println("\n✓ Using AWS Bedrock")

	case "1":
		fallthrough
	default:
		// User chose Anthropic (default)
		provider = "anthropic"
		apiKey = os.Getenv("ANTHROPIC_API_KEY")

		if apiKey == "" {
			fmt.Println("\nError: ANTHROPIC_API_KEY not set")
			fmt.Println("\nPlease set it:")
			fmt.Println("  export ANTHROPIC_API_KEY='your-api-key'")
			os.Exit(1)
		}
		fmt.Println("\n✓ Using Anthropic API")
	}

	// Prompt for goal after provider and credentials are set
	fmt.Println("\nWhat would you like to build?")
	fmt.Print("> ")

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
