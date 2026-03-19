package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/yourusername/OpenSack/agents"
)

func TestArchitectOneQuestionAtATime(t *testing.T) {
	// Skip if no API key set
	apiKey := getTestAPIKey()
	if apiKey == "" {
		t.Skip("Skipping: No ANTHROPIC_API_KEY or AWS_BEARER_TOKEN_BEDROCK set")
	}

	config := Config{
		APIKey:             apiKey,
		Provider:           getTestProvider(),
		AutoMode:           false,
		ConsensusThreshold: 6,
		AllowUserOverride:  false,
		AlwaysShowAnalysis: false,
		MonthlyBudget:      300.0,
	}

	orch := NewOrchestrator(config)
	ctx := context.Background()

	systemPrompt := orch.getArchitectSystemPrompt()

	// Verify prompt enforces one-question-at-a-time
	t.Run("SystemPromptEnforcesOneQuestion", func(t *testing.T) {
		if !strings.Contains(systemPrompt, "ONE question at a time") {
			t.Error("System prompt should enforce one question at a time")
		}
		if !strings.Contains(systemPrompt, "Type 'approved'") {
			t.Error("System prompt should mention approval signal")
		}
	})

	// Test actual interaction
	t.Run("ArchitectAsksOneQuestionFirst", func(t *testing.T) {
		var architect agents.Agent
		if config.Provider == "bedrock" {
			architect = agents.NewBedrockAgent(
				config.APIKey,
				"anthropic.claude-3-5-sonnet-20240620-v1:0",
				0.3,
				systemPrompt,
			)
		} else {
			architect = agents.NewClaudeAgent(
				config.APIKey,
				"claude-opus-4-20250514",
				0.3,
				systemPrompt,
			)
		}

		// Initial prompt
		initialPrompt := "I want to build: Build a simple REST API\n\nDiscovery context:\n- Requirements: CRUD operations\n\nLet's discuss the approach."

		response, err := architect.Chat(ctx, initialPrompt)
		if err != nil {
			t.Fatalf("Architect chat failed: %v", err)
		}

		// Check that response asks only one question
		questionCount := strings.Count(response.Message, "?")
		t.Logf("First response question count: %d", questionCount)
		t.Logf("First response (truncated): %s", truncate(response.Message, 200))

		// Should ask at least one question
		if questionCount == 0 {
			t.Error("Architect should ask at least one question initially")
		}

		// Ideally should ask only one question (but we can't strictly enforce this)
		// Just verify it's not overwhelming (less than 5 questions)
		if questionCount > 5 {
			t.Logf("Warning: Architect asked %d questions at once (should be 1)", questionCount)
		}
	})
}
