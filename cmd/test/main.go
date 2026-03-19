package main

import (
	"fmt"

	"github.com/yourusername/OpenSack/agents"
	"github.com/yourusername/OpenSack/orchestrator"
)

func main() {
	analyzer := orchestrator.NewTaskAnalyzer(6)

	testCases := []struct {
		name string
		goal string
	}{
		{
			name: "Simple task",
			goal: "Add a health check endpoint at /health",
		},
		{
			name: "Medium task",
			goal: "Fix the bug where users can't upload files larger than 5MB",
		},
		{
			name: "Complex architectural task",
			goal: "Design a microservices architecture for a real-time chat application with authentication, scalability, and security considerations",
		},
		{
			name: "Architectural decision",
			goal: "Evaluate different approaches for implementing real-time notifications. Compare WebSockets, SSE, and polling with tradeoffs",
		},
		{
			name: "Refactoring task",
			goal: "Refactor the entire authentication system across multiple components to support OAuth2 and SAML",
		},
		{
			name: "Performance optimization",
			goal: "Add performance optimization and caching strategy to handle scale of 10000 concurrent users",
		},
	}

	fmt.Println("=== Complexity Analysis Test ===\n")

	for _, tc := range testCases {
		fmt.Printf("Test: %s\n", tc.name)
		fmt.Printf("Goal: %s\n", tc.goal)

		discovery := agents.Discovery{
			Goal:                tc.goal,
			EstimatedComplexity: "medium",
		}

		analysis := analyzer.Analyze(tc.goal, discovery)

		fmt.Printf("Result:\n")
		fmt.Printf("  Complexity: %s (score: %d)\n", analysis.Level, analysis.Score)
		fmt.Printf("  Mode: %s\n", analysis.RecommendedMode)
		fmt.Printf("  Cost: $%.2f\n", analysis.EstimatedCost)

		if len(analysis.Reasons) > 0 {
			fmt.Printf("  Reasons:\n")
			for _, reason := range analysis.Reasons {
				fmt.Printf("    • %s\n", reason)
			}
		}

		fmt.Println()
	}
}
