# OpenSack

A Go-based AI agent orchestrator that manages the complete software development lifecycle from planning to review, with intelligent mode selection based on task complexity.

## Features

- **🌐 Multi-Provider**: Supports both Anthropic API and AWS Bedrock
- **🎯 Interactive Workflow**: No command-line arguments needed - fully guided experience
- **📊 Smart Mode Selection**: Three complexity modes (Fast/Explore/Consensus)
- **🔍 Optional Discovery**: Quick 3-step requirements gathering
- **💬 One-Question-at-a-Time**: Architect asks focused questions, not overwhelming
- **👨‍💻 Developer Phase**: Claude Haiku implements the plan efficiently
- **👀 Parallel Review**: Opus + Sonnet review in parallel for quality assurance
- **💰 Cost Tracking**: Monitor spending and mode usage

## Architecture

```
Provider Selection → Goal Input → Optional Discovery (3 steps)
                                         ↓
                              Complexity Analysis
                    ├─ Score < 3: Simple → Fast Mode
                    ├─ Score 3-5: Medium → Explore Mode
                    └─ Score ≥ 6: Complex → Consensus Mode
                                         ↓
                        Interactive Architecture (one question at a time)
                                         ↓
                            Developer (Claude Haiku)
                                         ↓
                          Review (Opus + Sonnet in parallel)
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

3. Set up API keys (choose one):

**Option 1: Anthropic API (default)**
```bash
export ANTHROPIC_API_KEY='your-anthropic-key'
```

**Option 2: AWS Bedrock**
```bash
export CLAUDE_CODE_USE_BEDROCK=1
export AWS_BEARER_TOKEN_BEDROCK='your-bearer-token'
```

4. Build:
```bash
go build -o opensack
```

## Usage

### Interactive Workflow

Simply run:
```bash
./opensack
```

You'll be guided through:

1. **Provider Selection**
   ```
   Which provider would you like to use?
   1. Anthropic API
   2. AWS Bedrock
   Choice [1]:
   ```

2. **Goal Input**
   ```
   What would you like to build?
   > Build a REST API for todo management
   ```

3. **Optional Discovery** (y/N to skip)
   ```
   Run detailed discovery? (helps architect understand better) [y/N]:
   ```
   - If yes: Answer 3 quick questions about requirements, constraints, and concerns
   - If no: Skip straight to architecture

4. **Complexity Analysis & Mode Selection**
   - System analyzes your goal
   - Recommends Fast/Explore/Consensus mode
   - You can accept or override

5. **Interactive Architecture**
   - Architect asks **one question at a time**
   - Answer each question naturally
   - Type `approved` when ready for proposal
   - Review proposal and type `approved` again to finalize

6. **Development & Review**
   - Developer (Haiku) implements the plan
   - Reviewers (Opus + Sonnet) review in parallel
   - Consensus on quality and issues

## Complexity Analysis

The orchestrator analyzes tasks based on:

- **Keywords**: architecture, design, refactor, scale, security, performance
- **Scope**: Multiple components, files, or services
- **Decision-making**: Words like "approach", "compare", "evaluate"
- **Requirements**: Security, performance, scalability needs
- **New technology**: Unfamiliar patterns or frameworks

**Scoring:**
- Score < 3: Simple → **Fast mode** (single architect)
- Score 3-5: Medium → **Explore mode** (balanced approach)
- Score ≥ 6: Complex → **Consensus mode** (multiple architects)

## Modes

### Fast Mode (Score < 3)
- **Single Claude Opus architect**
- Interactive one-question-at-a-time planning
- Claude Haiku developer
- Opus + Sonnet reviewers
- **Best for:** Bug fixes, simple features, well-defined tasks

### Explore Mode (Score 3-5)
- **Single Claude Sonnet architect** (balanced approach)
- Interactive planning with exploration
- Claude Haiku developer
- Opus + Sonnet reviewers
- **Best for:** Medium complexity features, new integrations

### Consensus Mode (Score ≥ 6)
- **3 parallel architects** (Opus, Sonnet, Haiku with different focuses)
- Multiple perspectives on complex problems
- User picks the best approach
- Interactive refinement
- Claude Haiku developer
- Opus + Sonnet reviewers
- **Best for:** Architecture decisions, complex features, greenfield projects

## Configuration

The orchestrator is configured in `main.go`:

```go
config := orchestrator.Config{
    APIKey:             apiKey,           // Your API key
    Provider:           provider,         // "anthropic" or "bedrock"
    AutoMode:           true,             // Auto-select mode based on complexity
    ConsensusThreshold: 6,                // Score to trigger consensus (default: 6)
    AllowUserOverride:  true,             // Let user override mode selection
    AlwaysShowAnalysis: true,             // Show complexity analysis
    MonthlyBudget:      300.0,            // Budget tracking
}
```

## Interactive Planning

Once a mode is selected, you'll chat with the architect **one question at a time**:

```
Architect: What's the expected load for this API?

You: Around 1000 concurrent users

Architect: Should we handle authentication?

You: Yes, OAuth2

Architect: Any existing systems to integrate with?

You: Yes, our Postgres database

Architect: I have enough information. Type 'approved' when ready for my proposal.

You: approved

Architect: Here's my proposal:
  1. REST API with Express.js
  2. Postgres for persistence
  3. OAuth2 middleware with JWT
  4. Rate limiting for 1000 concurrent users
  Does this sound right?

You: approved

✓ Plan approved! Generating detailed implementation plan...
```

The key improvement: **one focused question at a time**, not overwhelming you with multiple questions at once.

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
│   ├── types.go           # Data structures
│   ├── claude.go          # Claude API client (Anthropic)
│   └── bedrock.go         # Bedrock API client (AWS)
├── orchestrator/
│   ├── analyzer.go        # Complexity analysis (Fast/Explore/Consensus)
│   ├── orchestrator.go    # Main orchestration logic
│   ├── cost_tracker.go    # Cost tracking
│   ├── orchestrator_test.go        # Developer & Review tests
│   └── architect_flow_test.go      # One-question-at-a-time test
├── cmd/
│   └── test-bedrock/      # Bedrock connection test utility
├── main.go                # Entry point with interactive flow
├── go.mod
└── README.md
```

## Workflow Details

### 1. Discovery Phase (Optional)
- **Optional 3-step process** (can skip with N)
- Step 1: Requirements - key features
- Step 2: Constraints - technical, security, performance
- Step 3: Other concerns - risks, assumptions
- Auto-categorizes responses into appropriate fields

### 2. Architecture Phase
- **Complexity Analysis**: Scores the task (0-10+)
- **Mode Selection**: Fast (< 3), Explore (3-5), Consensus (≥ 6)
- **Interactive Planning**: One question at a time approach
- **Two-stage approval**:
  1. First `approved` → See proposal
  2. Second `approved` → Generate detailed plan
- **Plan Generation**: Detailed implementation plan with files, functions, steps

### 3. Implementation Phase ✅
- **Developer agent** (Claude Haiku) implements the plan
- Strictly follows the plan - no creative decisions
- Returns implementation summary
- Efficient and token-optimized

### 4. Review Phase ✅
- **Parallel reviews** from 2 reviewers:
  - Claude Opus - Quality & architecture focus
  - Claude Sonnet - Practical implementation focus
- **Consensus handling**:
  - All approve → Success ✅
  - Mixed reviews → Some concerns ⚠️
  - No approvals → Escalate to architect ❌

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

- [x] ~~Implement Developer agent~~ ✅ Done (Claude Haiku)
- [x] ~~Implement Reviewer agents~~ ✅ Done (Opus + Sonnet)
- [x] ~~AWS Bedrock support~~ ✅ Done
- [x] ~~Interactive workflow~~ ✅ Done
- [x] ~~One-question-at-a-time architecture~~ ✅ Done
- [x] ~~Optional discovery phase~~ ✅ Done
- [x] ~~Explore mode~~ ✅ Done
- [ ] Add GPT-4 and Gemini clients for additional reviewers
- [ ] Actual file writing/modification in Developer phase
- [ ] Persistent conversation history
- [ ] Plan visualization
- [ ] Review feedback iteration loop
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
