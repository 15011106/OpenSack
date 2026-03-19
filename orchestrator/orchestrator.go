package orchestrator

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Phase 1: Optional Discovery
	fmt.Println("\n=== Discovery Phase ===")
	fmt.Print("Run detailed discovery? (helps architect understand better) [y/N]: ")

	scanner := bufio.NewScanner(os.Stdin)
	runDiscovery := false
	if scanner.Scan() {
		answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
		runDiscovery = answer == "y" || answer == "yes"
	}

	var discovery agents.Discovery
	if runDiscovery {
		fmt.Println("Let's understand your requirements in 3 quick steps...")
		fmt.Println()
		discovery = o.interactiveDiscovery(ctx, goal)
	} else {
		fmt.Println("Skipping discovery, going straight to architecture...")
		fmt.Println()
		discovery = o.createSimpleDiscovery(goal)
	}

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

	// Phase 3: Implementation
	fmt.Println("\n=== Implementation Phase ===")
	impl, err := o.DeveloperPhase(ctx, plan)
	if err != nil {
		return fmt.Errorf("developer phase failed: %w", err)
	}
	fmt.Println("\n✓ Implementation complete")

	// Phase 4: Review
	fmt.Println("\n=== Review Phase ===")
	reviews, err := o.ReviewPhase(ctx, plan, impl)
	if err != nil {
		return fmt.Errorf("review phase failed: %w", err)
	}

	// Handle review feedback
	if err := o.handleReviewFeedback(ctx, plan, impl, reviews); err != nil {
		return fmt.Errorf("review handling failed: %w", err)
	}

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
			fmt.Println("→ Using Fast mode")
			fmt.Println()
		case 3:
			selectedMode = ConsensusMode
			fmt.Println("→ Using Consensus mode")
			fmt.Println()
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
	fmt.Println("=== Fast Mode: Single Architect ===")
	fmt.Println()

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
	fmt.Println("=== Consensus Mode: 3 Architects ===")
	fmt.Println()
	fmt.Println("Generating 3 architectural proposals...")
	fmt.Println()

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
	fmt.Println("=== Three Proposals ===")
	fmt.Println()
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
	fmt.Println("Type 'quit' to exit.")
	fmt.Println()

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
			fmt.Println("✓ Plan approved! Generating detailed implementation plan...")
			fmt.Println()
			break
		}
	}

	// Generate detailed plan
	return architect.GeneratePlan(ctx)
}

// Helper methods
func (o *Orchestrator) createSimpleDiscovery(goal string) agents.Discovery {
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

func (o *Orchestrator) interactiveDiscovery(ctx context.Context, goal string) agents.Discovery {
	scanner := bufio.NewScanner(os.Stdin)

	discovery := agents.Discovery{
		Goal:                    goal,
		Requirements:            []string{},
		Constraints:             []string{},
		Risks:                   []string{},
		Assumptions:             []string{},
		SecurityRequirements:    []string{},
		PerformanceRequirements: []string{},
		EstimatedComplexity:     "medium",
	}

	// Step 1: Requirements
	fmt.Println("📋 Step 1/3: Requirements")
	fmt.Print("   Key features or what it should do (comma-separated, or Enter to skip): ")
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			features := strings.Split(text, ",")
			for _, f := range features {
				if trimmed := strings.TrimSpace(f); trimmed != "" {
					discovery.Requirements = append(discovery.Requirements, trimmed)
				}
			}
		}
	}

	// Step 2: Constraints (combined: technical, security, performance)
	fmt.Println("\n⚙️  Step 2/3: Constraints & Requirements")
	fmt.Print("   Technical limits, security needs, or performance requirements (comma-separated, or Enter to skip): ")
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			items := strings.Split(text, ",")
			for _, item := range items {
				trimmed := strings.TrimSpace(item)
				if trimmed != "" {
					lower := strings.ToLower(trimmed)
					if strings.Contains(lower, "auth") || strings.Contains(lower, "security") || strings.Contains(lower, "permission") {
						discovery.SecurityRequirements = append(discovery.SecurityRequirements, trimmed)
					} else if strings.Contains(lower, "performance") || strings.Contains(lower, "scale") || strings.Contains(lower, "load") {
						discovery.PerformanceRequirements = append(discovery.PerformanceRequirements, trimmed)
					} else {
						discovery.Constraints = append(discovery.Constraints, trimmed)
					}
				}
			}
		}
	}

	// Step 3: Other concerns
	fmt.Println("\n🎯 Step 3/3: Other Concerns")
	fmt.Print("   Any risks, assumptions, or concerns? (comma-separated, or Enter to skip): ")
	if scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text != "" {
			items := strings.Split(text, ",")
			for _, item := range items {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					discovery.Risks = append(discovery.Risks, trimmed)
				}
			}
		}
	}

	// Estimate complexity based on gathered info
	complexity := len(discovery.Requirements) + len(discovery.SecurityRequirements) + len(discovery.PerformanceRequirements)
	if complexity < 3 {
		discovery.EstimatedComplexity = "low"
	} else if complexity < 6 {
		discovery.EstimatedComplexity = "medium"
	} else {
		discovery.EstimatedComplexity = "high"
	}

	fmt.Println("\n✓ Discovery complete!")
	return discovery
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
	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal plan: %w", err)
	}

	if err := os.WriteFile("plan.json", planJSON, 0644); err != nil {
		return fmt.Errorf("failed to write plan.json: %w", err)
	}

	fmt.Printf("Plan summary: %s\n", plan.Summary)
	return nil
}

// DeveloperPhase implements the plan using Claude Haiku
func (o *Orchestrator) DeveloperPhase(ctx context.Context, plan agents.ImplementationPlan) (agents.ImplementationResult, error) {
	systemPrompt := `You are a developer implementing an approved plan.

Your job:
1. Follow the plan strictly - no creative decisions
2. Implement the exact files, functions, and logic specified
3. Write clean, working code
4. Provide COMPLETE file content for each file

CRITICAL: For each file, use this exact format:

FILE: path/to/file.ext
` + "```" + `language
[complete file content here]
` + "```" + `

DO NOT:
- Make architectural decisions (that's done)
- Add features not in the plan
- Skip specified requirements
- Provide partial file content or diffs - ALWAYS provide complete files`

	var developer agents.Agent
	if o.config.Provider == "bedrock" {
		developer = agents.NewBedrockAgent(
			o.config.APIKey,
			"anthropic.claude-3-haiku-20240307-v1:0",
			0.3,
			systemPrompt,
		)
	} else {
		developer = agents.NewClaudeAgent(
			o.config.APIKey,
			"claude-3-haiku-20240307",
			0.3,
			systemPrompt,
		)
	}

	// Send plan to developer
	planJSON, _ := json.MarshalIndent(plan, "", "  ")
	prompt := fmt.Sprintf(`Implement this plan:

%s

For each file, provide COMPLETE content using the format:

FILE: path/to/file.ext
`+"`"+`go
[complete file content]
`+"`"+`

After all files, provide a brief summary.`, string(planJSON))

	response, err := developer.Chat(ctx, prompt)
	if err != nil {
		return agents.ImplementationResult{}, err
	}

	// Write files and generate diff
	result, err := o.writeFilesFromResponse(response.Message)
	if err != nil {
		return agents.ImplementationResult{}, fmt.Errorf("failed to write files: %w", err)
	}

	return result, nil
}

// ReviewPhase runs parallel reviews with Opus and Sonnet
func (o *Orchestrator) ReviewPhase(ctx context.Context, plan agents.ImplementationPlan, impl agents.ImplementationResult) ([]agents.ReviewResult, error) {
	type reviewJob struct {
		model string
		focus string
	}

	jobs := []reviewJob{
		{"claude-opus-4-20250514", "quality and architecture"},
		{"claude-sonnet-4-20250514", "practical implementation"},
	}

	reviewChan := make(chan agents.ReviewResult, len(jobs))
	var wg sync.WaitGroup

	for _, job := range jobs {
		wg.Add(1)
		go func(model, focus string) {
			defer wg.Done()

			systemPrompt := fmt.Sprintf(`You are a code reviewer focusing on %s.

Review the implementation against the plan:
1. Does it match the plan?
2. Are there bugs or issues?
3. Code quality concerns?
4. Security issues?

Be constructive but thorough.`, focus)

			var reviewer agents.Agent
			if o.config.Provider == "bedrock" {
				reviewer = agents.NewBedrockAgent(
					o.config.APIKey,
					"anthropic.claude-3-5-sonnet-20240620-v1:0",
					0.3,
					systemPrompt,
				)
			} else {
				reviewer = agents.NewClaudeAgent(
					o.config.APIKey,
					model,
					0.3,
					systemPrompt,
				)
			}

			planJSON, _ := json.MarshalIndent(plan, "", "  ")
			prompt := fmt.Sprintf(`Review this implementation:

PLAN:
%s

IMPLEMENTATION:
%s

Provide:
1. Critical issues (blockers)
2. Minor issues (suggestions)
3. Overall quality score (0-100)
4. Approve? (yes/no)`, string(planJSON), impl.Summary)

			response, err := reviewer.Chat(ctx, prompt)
			if err != nil {
				fmt.Printf("Warning: %s review failed: %v\n", model, err)
				return
			}

			reviewChan <- agents.ReviewResult{
				Model:        model,
				Summary:      response.Message,
				Approved:     strings.Contains(strings.ToLower(response.Message), "approve"),
				QualityScore: 75.0,
			}
		}(job.model, job.focus)
	}

	go func() {
		wg.Wait()
		close(reviewChan)
	}()

	var reviews []agents.ReviewResult
	for review := range reviewChan {
		reviews = append(reviews, review)
	}

	return reviews, nil
}

// handleReviewFeedback decides what to do with review results
func (o *Orchestrator) handleReviewFeedback(ctx context.Context, plan agents.ImplementationPlan, impl agents.ImplementationResult, reviews []agents.ReviewResult) error {
	fmt.Println("=== Review Results ===")
	fmt.Println()

	approvedCount := 0
	for i, review := range reviews {
		fmt.Printf("Reviewer %d (%s):\n", i+1, review.Model)
		fmt.Printf("  Approved: %v\n", review.Approved)
		fmt.Printf("  Quality Score: %.0f/100\n", review.QualityScore)
		fmt.Printf("  Summary: %s\n\n", truncate(review.Summary, 200))

		if review.Approved {
			approvedCount++
		}
	}

	// Consensus decision
	if approvedCount == len(reviews) {
		fmt.Println("✅ All reviewers approved!")
		return nil
	} else if approvedCount > 0 {
		fmt.Println("⚠️  Mixed reviews - some concerns raised")
		fmt.Println("   (In full implementation, would iterate with developer)")
		return nil
	} else {
		fmt.Println("❌ No approvals - significant issues found")
		fmt.Println("   (In full implementation, would escalate to architect)")
		return nil
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// writeFilesFromResponse parses the developer's response and writes files to disk
func (o *Orchestrator) writeFilesFromResponse(response string) (agents.ImplementationResult, error) {
	result := agents.ImplementationResult{
		FilesCreated:  []string{},
		FilesModified: []string{},
		Summary:       "",
	}

	// Parse response for FILE: markers and code blocks
	lines := strings.Split(response, "\n")
	var currentFile string
	var currentContent strings.Builder
	inCodeBlock := false

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for FILE: marker
		if strings.HasPrefix(line, "FILE:") {
			// Write previous file if any
			if currentFile != "" && inCodeBlock {
				if err := o.writeFile(currentFile, currentContent.String(), &result); err != nil {
					return result, err
				}
			}

			// Start new file
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "FILE:"))
			currentContent.Reset()
			inCodeBlock = false
			continue
		}

		// Check for code block markers
		if strings.HasPrefix(line, "```") {
			if !inCodeBlock {
				// Starting code block
				inCodeBlock = true
			} else {
				// Ending code block - write file
				if currentFile != "" {
					if err := o.writeFile(currentFile, currentContent.String(), &result); err != nil {
						return result, err
					}
					currentFile = ""
					currentContent.Reset()
				}
				inCodeBlock = false
			}
			continue
		}

		// Collect content inside code block
		if inCodeBlock && currentFile != "" {
			currentContent.WriteString(line)
			currentContent.WriteString("\n")
		}
	}

	// Write last file if any
	if currentFile != "" && inCodeBlock {
		if err := o.writeFile(currentFile, currentContent.String(), &result); err != nil {
			return result, err
		}
	}

	// Generate git diff
	if len(result.FilesCreated) > 0 || len(result.FilesModified) > 0 {
		diff, err := o.generateDiff()
		if err == nil {
			result.Diff = diff
		}
	}

	// Extract summary (text after all code blocks)
	summaryStarted := false
	var summary strings.Builder
	for _, line := range lines {
		if !strings.HasPrefix(line, "FILE:") && !strings.HasPrefix(line, "```") {
			if summaryStarted || (!strings.HasPrefix(line, "FILE:") && len(strings.TrimSpace(line)) > 0) {
				summaryStarted = true
				summary.WriteString(line)
				summary.WriteString("\n")
			}
		}
	}
	result.Summary = strings.TrimSpace(summary.String())

	if result.Summary == "" {
		result.Summary = fmt.Sprintf("Created %d files, modified %d files",
			len(result.FilesCreated), len(result.FilesModified))
	}

	return result, nil
}

// writeFile writes a single file and tracks whether it was created or modified
func (o *Orchestrator) writeFile(filePath, content string, result *agents.ImplementationResult) error {
	// Check if file exists
	_, err := os.Stat(filePath)
	fileExists := err == nil

	// Create parent directories if needed
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}

	// Track creation vs modification
	if fileExists {
		result.FilesModified = append(result.FilesModified, filePath)
		fmt.Printf("✓ Modified: %s\n", filePath)
	} else {
		result.FilesCreated = append(result.FilesCreated, filePath)
		fmt.Printf("✓ Created: %s\n", filePath)
	}

	return nil
}

// generateDiff generates a git diff of uncommitted changes
func (o *Orchestrator) generateDiff() (string, error) {
	cmd := fmt.Sprintf("cd %s && git diff HEAD", os.Getenv("PWD"))
	output, err := exec.CommandContext(context.Background(), "sh", "-c", cmd).Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (o *Orchestrator) getArchitectSystemPrompt() string {
	return `You are an expert software architect. Your job is to help design a feature through conversation.

CRITICAL RULES:
1. Ask ONE question at a time - never dump multiple questions at once
2. Wait for the human's answer before asking the next question
3. After gathering all information (usually 5-10 questions), say: "I have enough information. Type 'approved' when ready for my proposal."
4. Only show your proposal AFTER the human types "approved"
5. After showing the proposal, allow discussion and refinement
6. Generate the final detailed plan only when human types "approved" again

Question types to ask (one at a time):
- What problem does this solve?
- Expected load/scale?
- Authentication/authorization needs?
- Existing systems to integrate with?
- Performance requirements?
- Security concerns?
- Edge cases to handle?
- Technology preferences?

When proposing:
- Be specific about files, functions, and codepaths
- Explain tradeoffs clearly
- Point out potential issues or risks
- Suggest options when multiple approaches exist

The human will shape this plan. Your role is to help them think through it, not to decide for them.

After final approval, generate a detailed implementation plan with:
- Exact file structure
- Exact function signatures and step-by-step logic
- Security boundaries and validations
- Error handling patterns
- Test requirements

The plan should be SO detailed that a developer makes ZERO architectural decisions.`
}
