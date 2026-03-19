# OpenSack

A Go-based AI agent orchestrator that intelligently routes tasks between single-architect (fast) and multi-architect (consensus) modes based on complexity analysis.

## Features

- **Smart Mode Selection**: Automatically analyzes task complexity and selects the optimal mode
- **Consensus Mode**: Uses 3 different AI models (Claude, GPT-4, Gemini) for complex architectural decisions
- **Fast Mode**: Single architect for straightforward tasks
- **Interactive Planning**: Chat with the architect until you approve the plan
- **Cost Tracking**: Monitor spending and mode usage
- **User Override**: Always allows manual mode selection

## Architecture

```
Discovery → Architecture (Fast/Consensus) → Implementation → Review
               ↓
        Complexity Analysis
        ├─ Simple/Medium → Fast Mode (1 architect)
        └─ Complex → Consensus Mode (3 architects)
```

## Installation

1. Clone the repository:
```bash
git clone <repo-url>
cd opensack
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up API keys:
```bash
export ANTHROPIC_API_KEY='your-anthropic-key'
# Optional for consensus mode:
export OPENAI_API_KEY='your-openai-key'
export GEMINI_API_KEY='your-gemini-key'
```

4. Build:
```bash
go build -o opensack
```

## Usage

### Basic Usage

```bash
./opensack "Your goal here"
```

### Examples

**Simple task (Fast mode):**
```bash
./opensack "Add a health check endpoint at /health"
```

**Complex task (Consensus mode):**
```bash
./opensack "Design a microservices architecture for real-time chat with authentication"
```

**Architectural decision (Consensus mode):**
```bash
./opensack "Evaluate different approaches for implementing real-time notifications. Compare WebSockets, SSE, and polling."
```

## Complexity Analysis

The orchestrator analyzes tasks based on:

- **Keywords**: architecture, design, refactor, scale, security, performance
- **Scope**: Multiple components, files, or services
- **Decision-making**: Words like "approach", "compare", "evaluate"
- **Requirements**: Security, performance, scalability needs
- **New technology**: Unfamiliar patterns or frameworks

**Scoring:**
- Score < 3: Simple → Fast mode
- Score 3-5: Medium → Fast mode
- Score ≥ 6: Complex → Consensus mode (by default)

## Modes

### Fast Mode ($3.43 per feature)
- Single Claude Opus architect
- Interactive planning chat
- Quick for clear requirements
- Best for: Bug fixes, simple features, well-defined tasks

### Consensus Mode ($3.81 per feature)
- 3 architects (Claude, GPT-4, Gemini)
- Multiple perspectives on complex problems
- User picks the best approach
- Interactive refinement
- Best for: Architecture decisions, complex features, greenfield projects

## Configuration

Edit `main.go` to configure:

```
go
config := orchestrator.Config{
    AnthropicAPIKey:    apiKey,
    AutoMode:           true,     // Auto-select mode
    ConsensusThreshold: 6,        // Score to trigger consensus
    AllowUserOverride:  true,     // Let user override selection
    AlwaysShowAnalysis: true,     // Show complexity analysis
    MonthlyBudget:      300.0,    // Budget tracking
}
```

## Interactive Planning

Once a mode is selected, you'll chat with the architect:

```
Architect: I can help with that. A few questions:
  - What's the expected load?
  - Should we handle authentication?
  - Any existing systems to integrate with?

You: Expected 1000 concurrent users, yes use OAuth, integrate with our Postgres DB

Architect: Got it. Here's my proposal:
  1. Use WebSocket for real-time
  2. Postgres for persistence
  3. OAuth2 middleware
  Does this sound right?

You: approved

✓ Plan approved! Generating detailed implementation plan...
```

## Cost Tracking

At the end of each session:

```
=== Cost Summary ===
Total spent:         $7.62
This session:        $7.62
Features built:      2
Avg cost per feature: $3.81

Mode usage:
  Consensus: 1 (50.0%)
  Fast:      1 (50.0%)
```

## Project Structure

```
opensack/
├── agents/
│   ├── types.go       # Data structures
│   └── claude.go      # Claude API client
├── orchestrator/
│   ├── analyzer.go    # Complexity analysis
│   ├── orchestrator.go # Main orchestration logic
│   └── cost_tracker.go # Cost tracking
├── main.go            # Entry point
├── go.mod
└── README.md
```

## Workflow Details

### 1. Discovery Phase
- Analyzes the goal
- Identifies requirements, constraints, risks
- Determines complexity level

### 2. Architecture Phase
- **Complexity Analysis**: Scores the task
- **Mode Selection**: Fast or Consensus
- **Interactive Planning**: Chat until "approved"
- **Plan Generation**: Detailed implementation plan

### 3. Implementation Phase (TODO)
- Developer agent follows the plan
- Writes code, tests, documentation

### 4. Review Phase (TODO)
- Multiple reviewers critique the code
- Security, performance, quality checks
- Fix issues if needed

## Customization

### Adjust Consensus Threshold

Make consensus mode more/less likely:

```
go
ConsensusThreshold: 4,  // Lower = use consensus more often
ConsensusThreshold: 8,  // Higher = use fast mode more often
```

### Add Custom Keywords

Edit `orchestrator/analyzer.go`:

```
go
complexKeywords: []string{
    "architecture", "design", "refactor",
    "your-custom-keyword", // Add here
}
```

## Future Enhancements

- [ ] Implement Developer agent
- [ ] Implement Reviewer agents (Codex, Gemini)
- [ ] Add GPT-4 and Gemini clients for consensus mode
- [ ] Persistent conversation history
- [ ] Plan visualization
- [ ] Integration tests
- [ ] Web UI

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License

## Credits

Inspired by [Stavros Korokithakis's workflow](https://www.stavros.io/posts/how-i-write-software-with-llms/)
