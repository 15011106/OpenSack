package orchestrator

import (
	"fmt"
	"sync"
)

// CostTracker tracks usage and costs
type CostTracker struct {
	mu             sync.Mutex
	totalSpent     float64
	monthlySpent   float64
	featureCount   int
	consensusCount int
	fastCount      int
	exploreCount   int
}

// NewCostTracker creates a new cost tracker
func NewCostTracker() *CostTracker {
	return &CostTracker{}
}

// RecordUsage records a usage event
func (c *CostTracker) RecordUsage(mode ArchitectMode, cost float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.totalSpent += cost
	c.monthlySpent += cost
	c.featureCount++

	switch mode {
	case ConsensusMode:
		c.consensusCount++
	case ExploreMode:
		c.exploreCount++
	case FastMode:
		c.fastCount++
	}
}

// GetStats returns formatted statistics
func (c *CostTracker) GetStats() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.featureCount == 0 {
		return "No usage yet"
	}

	avgCost := c.totalSpent / float64(c.featureCount)

	return fmt.Sprintf(`=== Cost Summary ===
Total spent:         $%.2f
This session:        $%.2f
Features built:      %d
Avg cost per feature: $%.2f

Mode usage:
  Consensus: %d (%.1f%%)
  Explore:   %d (%.1f%%)
  Fast:      %d (%.1f%%)
`,
		c.totalSpent,
		c.monthlySpent,
		c.featureCount,
		avgCost,
		c.consensusCount,
		percentage(c.consensusCount, c.featureCount),
		c.exploreCount,
		percentage(c.exploreCount, c.featureCount),
		c.fastCount,
		percentage(c.fastCount, c.featureCount),
	)
}

func percentage(count, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}
