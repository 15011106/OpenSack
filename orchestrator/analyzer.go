package orchestrator

import (
	"fmt"
	"strings"

	"github.com/yourusername/OpenSack/agents"
)

// ComplexityLevel represents task complexity
type ComplexityLevel int

const (
	Simple ComplexityLevel = iota
	Medium
	Complex
)

func (c ComplexityLevel) String() string {
	switch c {
	case Simple:
		return "Simple"
	case Medium:
		return "Medium"
	case Complex:
		return "Complex"
	default:
		return "Unknown"
	}
}

// ArchitectMode represents the orchestration mode
type ArchitectMode string

const (
	FastMode      ArchitectMode = "fast"
	ExploreMode   ArchitectMode = "explore"
	ConsensusMode ArchitectMode = "consensus"
)

// ComplexityAnalysis holds the analysis result
type ComplexityAnalysis struct {
	Score           int
	Level           ComplexityLevel
	RecommendedMode ArchitectMode
	Reasons         []string
	EstimatedCost   float64
}

// TaskAnalyzer analyzes task complexity
type TaskAnalyzer struct {
	complexKeywords     []string
	multiComponentWords []string
	architecturalWords  []string
	consensusThreshold  int
}

// NewTaskAnalyzer creates a new task analyzer
func NewTaskAnalyzer(consensusThreshold int) *TaskAnalyzer {
	return &TaskAnalyzer{
		complexKeywords: []string{
			"architecture", "design", "refactor", "migrate",
			"scale", "distributed", "microservices", "system design",
			"performance optimization", "security audit",
		},
		multiComponentWords: []string{
			"multiple", "several", "across", "entire system",
			"full stack", "end-to-end", "integration",
		},
		architecturalWords: []string{
			"approach", "strategy", "pattern", "tradeoff",
			"evaluate options", "compare", "best way",
		},
		consensusThreshold: consensusThreshold,
	}
}

// Analyze analyzes the task and recommends a mode
func (t *TaskAnalyzer) Analyze(goal string, discovery agents.Discovery) ComplexityAnalysis {
	score := 0
	reasons := []string{}

	goalLower := strings.ToLower(goal)

	// Check 1: Complex keywords
	for _, keyword := range t.complexKeywords {
		if strings.Contains(goalLower, keyword) {
			score += 3
			reasons = append(reasons, fmt.Sprintf("Complex keyword: '%s'", keyword))
		}
	}

	// Check 2: Multiple components
	for _, word := range t.multiComponentWords {
		if strings.Contains(goalLower, word) {
			score += 2
			reasons = append(reasons, fmt.Sprintf("Multi-component work: '%s'", word))
		}
	}

	// Check 3: Architectural decision words
	for _, word := range t.architecturalWords {
		if strings.Contains(goalLower, word) {
			score += 2
			reasons = append(reasons, fmt.Sprintf("Architectural decision: '%s'", word))
		}
	}

	// Check 4: Discovery complexity
	if discovery.EstimatedComplexity == "high" {
		score += 3
		reasons = append(reasons, "Discovery indicates high complexity")
	}

	// Check 5: File count mentioned
	fileCount := countMentionedFiles(goal)
	if fileCount > 3 {
		score += 2
		reasons = append(reasons, fmt.Sprintf("Multiple files mentioned (%d)", fileCount))
	}

	// Check 6: Security requirements
	if len(discovery.SecurityRequirements) > 0 {
		score += 2
		reasons = append(reasons, "Security considerations required")
	}

	// Check 7: Performance requirements
	if len(discovery.PerformanceRequirements) > 0 {
		score += 2
		reasons = append(reasons, "Performance optimization needed")
	}

	// Check 8: New technology
	if containsNewTech(goal) {
		score += 2
		reasons = append(reasons, "Unfamiliar technology or pattern")
	}

	// Determine complexity level and mode
	var level ComplexityLevel
	var mode ArchitectMode

	if score >= t.consensusThreshold {
		level = Complex
		mode = ConsensusMode
	} else if score >= 3 {
		level = Medium
		mode = FastMode
	} else {
		level = Simple
		mode = FastMode
	}

	return ComplexityAnalysis{
		Score:           score,
		Level:           level,
		RecommendedMode: mode,
		Reasons:         reasons,
		EstimatedCost:   estimateCost(mode),
	}
}

func countMentionedFiles(goal string) int {
	extensions := []string{".go", ".py", ".js", ".ts", ".java", ".md", ".yml", ".yaml"}
	count := 0
	goalLower := strings.ToLower(goal)

	for _, ext := range extensions {
		count += strings.Count(goalLower, ext)
	}

	fileWords := []string{"file", "component", "module", "package", "service"}
	for _, word := range fileWords {
		if strings.Contains(goalLower, word) {
			count++
		}
	}

	return count
}

func containsNewTech(goal string) bool {
	newTechKeywords := []string{
		"new framework", "never used", "first time",
		"learning", "experiment", "prototype",
	}

	goalLower := strings.ToLower(goal)
	for _, keyword := range newTechKeywords {
		if strings.Contains(goalLower, keyword) {
			return true
		}
	}

	return false
}

func estimateCost(mode ArchitectMode) float64 {
	switch mode {
	case ConsensusMode:
		return 3.81
	case ExploreMode:
		return 3.60
	case FastMode:
		return 3.43
	default:
		return 3.43
	}
}
