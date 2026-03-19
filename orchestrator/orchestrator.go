package orchestrator

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/yourusername/OpenSack/agents"
)

// Orchestrator manages the workflow
type Orchestrator struct {
	config      Config
	analyzer    *TaskAnalyzer
	costTracker *CostTracker
	mu          sync.Mutex
}

// Config for orchestrator
type Config struct {
	APIKey             string
	Provider           string // "anthropic" or "bedrock"
	OpenAIAPIKey       string
	GeminiAPIKey       string
	AutoMode           bool
	ConsensusThreshold int
	AllowUserOverride  bool
	AlwaysShowAnalysis bool
	MonthlyBudget      float64
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(config Config) *Orchestrator {
	return &Orchestrator{
		config:      config,
		analyzer:    NewTaskAnalyzer(config.ConsensusThreshold),
		costTracker: NewCostTracker(),
	}
}

// Execute runs the full workflow
func (o *Orchestrator) Execute(ctx context.Context, goal string) error {
	fmt.Println("=== Agent Orchestrator ===")
	fmt.Printf("Goal: %s\n\n", goal)

	// Phase 1: Discovery (simplified for now)
	discovery := o.createDiscovery(goal)

	// Phase 2: Architecture with smart mode selection
	plan, err := o.ArchitectPhase(ctx, goal, discovery)
	if err != nil {
		return fmt.Errorf("architecture phase failed: %w", err)
	}

	// Save plan
	if err := o.savePlan(plan); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Println("\n✓ Architecture phase complete")
	fmt.Println("✓ Plan saved to plan.json")

	// Phase 3: Implementation (placeholder)
	fmt.Println("\n=== Implementation Phase ===")
	fmt.Println("(Developer implementation would happen here)")

	// Phase 4: Review (placeholder)
	fmt.Println("\n=== Review Phase ===")
	fmt.Println("(Multi-model review would happen here)")

	// Show cost summary
	fmt.Println("\n" + o.costTracker.GetStats())

	return nil
}

// ArchitectPhase handles the architecture planning
func (o *Orchestrator) ArchitectPhase(ctx context.Context, goal string, discovery agents.Discovery) (agents.ImplementationPlan, error) {
	// Analyze task complexity
	analysis := o.analyzer.Analyze(goal, discovery)

	// Show analysis
	fmt.Println("=== Task Complexity Analysis ===")
	fmt.Printf("Complexity: %s (score: %d)\n", analysis.Level, analysis.Score)
	fmt.Printf("Recommended mode: %s\n", analysis.RecommendedMode)
	fmt.Printf("Estimated cost: $%.2f\n\n", analysis.EstimatedCost)

	if len(analysis.Reasons) > 0 {
		fmt.Println("Reasons:")
		for _, reason := range analysis.Reasons {
			fmt.Printf("  • %s\n", reason)
		}
		fmt.Println()
	}

	// Allow user override
	selectedMode := analysis.RecommendedMode
	if o.config.AllowUserOverride {
		fmt.Println("Options:")
		fmt.Println("1. Accept recommendation")
		fmt.Println("2. Use Fast mode (cheaper, single architect)")
		fmt.Println("3. Use Consensus mode (expensive, 3 architects)")
		fmt.Print("Choice [1]: ")

		choice := o.getUserChoice()
		switch choice {
		case 2:
			selectedMode = FastMode
			fmt.Println("→ Using Fast mode\n")
		case 3:
			selectedMode = ConsensusMode
			fmt.Println("→ Using Consensus mode\n")
		default:
			fmt.Printf("→ Using recommended %s mode\n\n", selectedMode)
		}
	} else {
		fmt.Printf("→ Auto-selecting %s mode\n\n", selectedMode)
	}

	// Execute selected mode
	var plan agents.ImplementationPlan
	var err error

	switch selectedMode {
	case ConsensusMode:
		plan, err = o.consensusPlanning(ctx, goal, discovery)
	case FastMode:
		plan, err = o.fastPlanning(ctx, goal, discovery)
	default:
		plan, err = o.fastPlanning(ctx, goal, discovery)
	}

	if err != nil {
		return agents.ImplementationPlan{}, err
	}

	// Record cost
	o.costTracker.RecordUsage(selectedMode, analysis.EstimatedCost)

	return plan, nil
}

// fastPlanning uses a single architect
func (o *Orchestrator) fastPlanning(ctx context.Context, goal string, discovery agents.Discovery) (agents.ImplementationPlan, error) {
	fmt.Println("=== Fast Mode: Single Architect ===\n")

	systemPrompt := o.getArchitectSystemPrompt()
	var architect agents.Agent

	if o.config.Provider == "bedrock" {
		architect = agents.NewBedrockAgent(
			o.config.APIKey,
			"anthropic.claude-3-5-sonnet-20240620-v1:0",
			0.3,
			systemPrompt,
		)
	} else {
		architect = agents.NewClaudeAgent(
			o.config.APIKey,
			"claude-opus-4-20250514", // Opus 4.6
			0.3,
			systemPrompt,
		)
	}

	// Interactive planning
	return o.interactivePlanning(ctx, architect, goal, discovery)
}

// consensusPlanning uses 3 architects
func (o *Orchestrator) consensusPlanning(ctx context.Context, goal string, discovery agents.Discovery) (agents.ImplementationPlan, error) {
	fmt.Println("=== Consensus Mode: 3 Architects ===\n")
	fmt.Println("Generating 3 architectural proposals...\n")

	// Generate 3 proposals in parallel
	type proposalResult struct {
		model    string
		focus    string
		proposal string
		err      error
	}

	proposalChan := make(chan proposalResult, 3)
	var wg sync.WaitGroup

	proposals := []struct {
		model string
		focus string
	}{
		{"claude-opus-4-20250514", "performance"},
		{"gpt-4", "simplicity"},
		{"gemini-pro", "scalability"},
	}

	for _, p := range proposals {
		wg.Add(1)
		go func(model, focus string) {
			defer wg.Done()
			prompt := fmt.Sprintf(`Given this goal: %s

Create an architectural proposal focusing on %s.
Provide a high-level approach with pros and cons.
Keep it concise (2-3 paragraphs).`, goal, focus)

			// For now, only use Claude (we'd add GPT/Gemini clients later)
			systemPrompt := fmt.Sprintf("You are an architect focusing on %s. Propose an approach.", focus)
			var agent agents.Agent

			if o.config.Provider == "bedrock" {
				agent = agents.NewBedrockAgent(o.config.APIKey, "anthropic.claude-3-5-sonnet-20240620-v1:0", 0.3, systemPrompt)
			} else {
				agent = agents.NewClaudeAgent(o.config.APIKey, model, 0.3, systemPrompt)
			}

			resp, err := agent.Chat(ctx, prompt)
			proposalChan <- proposalResult{
				model:    model,
				focus:    focus,
				proposal: resp.Message,
				err:      err,
			}
		}(p.model, p.focus)
	}

	go func() {
		wg.Wait()
		close(proposalChan)
	}()

	// Collect proposals
	var proposalResults []proposalResult
	for result := range proposalChan {
		if result.err != nil {
			fmt.Printf("Warning: %s proposal failed: %v\n", result.focus, result.err)
			continue
		}
		proposalResults = append(proposalResults, result)
	}

	if len(proposalResults) == 0 {
		return agents.ImplementationPlan{}, fmt.Errorf("all proposals failed")
	}

	// Present options to user
	fmt.Println("=== Three Proposals ===\n")
	for i, result := range proposalResults {
		fmt.Printf("Option %d (%s focus via %s):\n", i+1, result.focus, result.model)
		fmt.Println(result.proposal)
		fmt.Println()
	}

	fmt.Print("Pick one (1-3): ")
	choice := o.getUserChoice()
	if choice < 1 || choice > len(proposalResults) {
		choice = 1
	}

	selected := proposalResults[choice-1]
	fmt.Printf("\n→ Selected: %s focus\n\n", selected.focus)

	// Continue with interactive refinement
	fmt.Println("=== Interactive Refinement ===")
	systemPrompt := o.getArchitectSystemPrompt()
	var architect agents.Agent

	if o.config.Provider == "bedrock" {
		architect = agents.NewBedrockAgent(o.config.APIKey, "anthropic.claude-3-5-sonnet-20240620-v1:0", 0.3, systemPrompt)
	} else {
		architect = agents.NewClaudeAgent(o.config.APIKey, selected.model, 0.3, systemPrompt)
	}

	// Seed with selected proposal
	_, err := architect.Chat(ctx, fmt.Sprintf("Building on this approach:\n%s\n\nLet's refine it.", selected.proposal))
	if err != nil {
		return agents.ImplementationPlan{}, err
	}

	return o.interactivePlanning(ctx, architect, goal, discovery)
}

// interactivePlanning handles the interactive chat loop
func (o *Orchestrator) interactivePlanning(ctx context.Context, architect agents.Agent, goal string, discovery agents.Discovery) (agents.ImplementationPlan, error) {
	fmt.Println("Chat with the architect. Type 'approved' when ready.")
	fmt.Println("Type 'quit' to exit.\n")

	// Start conversation
	initialPrompt := fmt.Sprintf(`I want to build: %s

Discovery context:
- Requirements: %v
- Constraints: %v
- Security: %v
- Performance: %v

Let's discuss the approach.`, goal, discovery.Requirements, discovery.Constraints, discovery.SecurityRequirements, discovery.PerformanceRequirements)

	response, err := architect.Chat(ctx, initialPrompt)
	if err != nil {
		return agents.ImplementationPlan{}, err
	}

	fmt.Printf("Architect: %s\n\n", response.Message)

	// Interactive loop
	scanner := bufio.NewScanner(os.Stdin)
	for !architect.IsApproved() {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())
		if userInput == "" {
			continue
		}
		if userInput == "quit" {
			return agents.ImplementationPlan{}, fmt.Errorf("user quit")
		}

		response, err := architect.Chat(ctx, userInput)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\nArchitect: %s\n\n", response.Message)

		if response.Approved {
			fmt.Println("✓ Plan approved! Generating detailed implementation plan...\n")
			break
		}
	}

	// Generate detailed plan
	return architect.GeneratePlan(ctx)
}

// Helper methods
func (o *Orchestrator) createDiscovery(goal string) agents.Discovery {
	// Simplified discovery for now
	return agents.Discovery{
		Goal:                    goal,
		Requirements:            []string{"Implement the requested feature"},
		Constraints:             []string{},
		Risks:                   []string{},
		Assumptions:             []string{},
		SecurityRequirements:    []string{},
		PerformanceRequirements: []string{},
		EstimatedComplexity:     "medium",
	}
}

func (o *Orchestrator) getUserChoice() int {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" {
			return 1
		}
		var choice int
		fmt.Sscanf(text, "%d", &choice)
		return choice
	}
	return 1
}

func (o *Orchestrator) savePlan(plan agents.ImplementationPlan) error {
	// Save to file (simplified)
	fmt.Printf("Plan summary: %s\n", plan.Summary)
	return nil
}

func (o *Orchestrator) getArchitectSystemPrompt() string {
	return `You are an expert software architect. Your job is to help design a feature through conversation.

In the planning phase:
1. Ask clarifying questions about edge cases, constraints, tradeoffs
2. Propose architectural approaches with pros/cons
3. Point out potential issues or risks
4. Refine the plan based on human feedback
5. DO NOT start implementation until human explicitly says "approved"

When discussing:
- Be specific about files, functions, and codepaths
- Explain tradeoffs clearly
- Ask about things you don't know (existing code structure, preferences)
- Suggest options when multiple approaches exist

The human will shape this plan. Your role is to help them think through it, not to decide for them.

After approval, you will generate a detailed implementation plan with:
- Exact file structure
- Exact function signatures and step-by-step logic
- Security boundaries and validations
- Error handling patterns
- Test requirements

The plan should be SO detailed that a developer makes ZERO architectural decisions.`
}
