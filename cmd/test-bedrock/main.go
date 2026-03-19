package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yourusername/OpenSack/agents"
)

func main() {
	bearerToken := os.Getenv("AWS_BEARER_TOKEN_BEDROCK")
	if bearerToken == "" {
		fmt.Println("Error: AWS_BEARER_TOKEN_BEDROCK not set")
		os.Exit(1)
	}

	fmt.Println("Testing Bedrock API connection...")
	fmt.Printf("Bearer Token (first 20 chars): %s...\n", bearerToken[:20])

	agent := agents.NewBedrockAgent(
		bearerToken,
		"anthropic.claude-3-5-sonnet-20240620-v1:0",
		0.3,
		"You are a helpful assistant.",
	)

	fmt.Println("\nSending test message...")
	ctx := context.Background()
	response, err := agent.Chat(ctx, "Say hello in one sentence.")

	if err != nil {
		fmt.Printf("\n❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Success!\nResponse: %s\n", response.Message)
}
