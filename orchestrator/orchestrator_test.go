package orchestrator

import (
	"context"
	"os"
	"testing"

	"github.com/yourusername/OpenSack/agents"
)

func TestDeveloperAndReviewPhase(t *testing.T) {
	// Skip if no API key set
	apiKey := getTestAPIKey()
	if apiKey == "" {
		t.Skip("Skipping: No ANTHROPIC_API_KEY or AWS_BEARER_TOKEN_BEDROCK set")
	}

	config := Config{
		APIKey:             apiKey,
		Provider:           getTestProvider(),
		AutoMode:           true,
		ConsensusThreshold: 6,
		AllowUserOverride:  false,
		AlwaysShowAnalysis: false,
		MonthlyBudget:      300.0,
	}

	orch := NewOrchestrator(config)
	ctx := context.Background()

	// Create a simple test plan
	plan := agents.ImplementationPlan{
		Summary: "Add a simple hello world function",
		Phases: []agents.Phase{
			{
				Name:        "Implementation",
				Description: "Create hello.go with HelloWorld function",
				Files:       []string{"hello.go"},
				Order:       1,
			},
		},
		Files: []agents.FileSpec{
			{
				Path:        "hello.go",
				Action:      "create",
				Description: "Simple hello world function",
				Functions: []agents.Function{
					{
						Name:        "HelloWorld",
						Signature:   "func HelloWorld() string",
						Description: "Returns 'Hello, World!'",
						Steps:       []string{"Return the string 'Hello, World!'"},
					},
				},
			},
		},
	}

	// Test Developer Phase
	t.Run("DeveloperPhase", func(t *testing.T) {
		impl, err := orch.DeveloperPhase(ctx, plan)
		if err != nil {
			t.Fatalf("DeveloperPhase failed: %v", err)
		}

		if impl.Summary == "" {
			t.Error("Implementation summary is empty")
		}

		// Verify files were created
		if len(impl.FilesCreated) == 0 && len(impl.FilesModified) == 0 {
			t.Error("No files created or modified")
		}

		// Verify actual files exist on disk
		for _, file := range impl.FilesCreated {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				t.Errorf("File %s was reported as created but doesn't exist", file)
			} else {
				// Clean up
				defer os.Remove(file)
				t.Logf("✓ Created file: %s", file)
			}
		}

		// Verify diff was generated
		if impl.Diff == "" {
			t.Log("Warning: No diff generated (might be normal if git not initialized)")
		} else {
			t.Logf("✓ Diff generated (%d bytes)", len(impl.Diff))
		}

		t.Logf("Developer summary: %s", impl.Summary[:min(len(impl.Summary), 200)])
	})

	// Test Review Phase
	t.Run("ReviewPhase", func(t *testing.T) {
		impl := agents.ImplementationResult{
			Summary:       "Implemented HelloWorld function in hello.go",
			FilesCreated:  []string{"hello.go"},
			FilesModified: []string{},
		}

		reviews, err := orch.ReviewPhase(ctx, plan, impl)
		if err != nil {
			t.Fatalf("ReviewPhase failed: %v", err)
		}

		if len(reviews) != 2 {
			t.Errorf("Expected 2 reviews, got %d", len(reviews))
		}

		for i, review := range reviews {
			if review.Summary == "" {
				t.Errorf("Review %d summary is empty", i)
			}
			t.Logf("Reviewer %d (%s): Approved=%v, Score=%.0f",
				i+1, review.Model, review.Approved, review.QualityScore)
		}
	})
}

func getTestAPIKey() string {
	// Check Bedrock first
	if token := getEnv("AWS_BEARER_TOKEN_BEDROCK"); token != "" {
		return token
	}
	// Then Anthropic
	return getEnv("ANTHROPIC_API_KEY")
}

func getTestProvider() string {
	if getEnv("AWS_BEARER_TOKEN_BEDROCK") != "" {
		return "bedrock"
	}
	return "anthropic"
}

func getEnv(key string) string {
	return os.Getenv(key)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
