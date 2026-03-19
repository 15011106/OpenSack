# Quick Start Guide

## 1. Setup (First Time)

### Set your API key:
```bash
export ANTHROPIC_API_KEY='your-api-key-here'
```

### Build the orchestrator:
```bash
cd ~/opensack
go build -o opensack
```

## 2. Test the Complexity Analyzer

```bash
go run test_analyzer.go
```

This shows how different tasks are classified:
- **Simple tasks** → Fast mode ($3.43)
- **Complex tasks** → Consensus mode ($3.81)

## 3. Run a Simple Example

**Try a simple task (Fast mode):**
```bash
./opensack "Add a health check endpoint at /health that returns 200 OK"
```

Expected output:
```
=== Task Complexity Analysis ===
Complexity: Simple (score: 0)
Recommended mode: fast
Estimated cost: $3.43

→ Using Fast mode

=== Fast Mode: Single Architect ===

Chat with the architect. Type 'approved' when ready.

Architect: I can help with that. Let me understand the requirements...
```

**Try a complex task (Consensus mode):**
```bash
./opensack "Design a microservices architecture for real-time chat with authentication and scalability"
```

Expected output:
```
=== Task Complexity Analysis ===
Complexity: Complex (score: 9)
Recommended mode: consensus
Estimated cost: $3.81

Reasons:
  • Complex keyword: 'architecture'
  • Complex keyword: 'design'
  • Complex keyword: 'microservices'

→ Using Consensus mode

=== Consensus Mode: 3 Architects ===

Generating 3 architectural proposals...
```

## 4. How to Use

### Interactive Chat Flow:

1. **Start with your goal:**
   ```bash
   ./opensack "your goal here"
   ```

2. **Review complexity analysis:**
   - See the score and recommended mode
   - Choose to accept or override

3. **Chat with architect:**
   ```
   Architect: I can help. A few questions:
     - What's the expected user count?
     - Should we handle authentication?

   You: 1000 users, yes OAuth

   Architect: Got it. Here's my proposal:
     1. Use WebSocket for real-time
     2. Postgres for storage
     3. OAuth2 middleware

   You: approved
   ```

4. **Get detailed plan:**
   - After "approved", architect generates implementation plan
   - Plan saved to `plan.json`

## 5. Configuration

Edit `main.go` to customize:

```
go
config := orchestrator.Config{
    ConsensusThreshold: 6,    // Lower = use consensus more
    AllowUserOverride:  true, // Let user choose mode
    MonthlyBudget:      300.0,
}
```

## 6. Example Commands

### Simple (Fast mode ~$3.43):
```bash
./opensack "Add logging to the API handlers"
./opensack "Fix the validation bug in user registration"
./opensack "Add rate limiting to the API"
```

### Complex (Consensus mode ~$3.81):
```bash
./opensack "Design a caching strategy with Redis for 100k users"
./opensack "Refactor authentication to support multiple providers"
./opensack "Compare approaches for implementing real-time features"
```

## 7. Project Structure

```
opensack/
├── opensack      # Compiled binary
├── main.go                 # Entry point
├── agents/
│   ├── types.go           # Data structures
│   └── claude.go          # Claude API client
├── orchestrator/
│   ├── analyzer.go        # Complexity analysis
│   ├── orchestrator.go    # Main logic
│   └── cost_tracker.go    # Cost tracking
├── test_analyzer.go       # Test script
├── examples.sh            # Example commands
└── README.md              # Full documentation
```

## 8. What Happens Next

Current implementation includes:
- ✅ Discovery phase (simplified)
- ✅ Architecture phase (Fast + Consensus modes)
- ✅ Interactive planning chat
- ✅ Complexity analysis
- ✅ Cost tracking

To be implemented:
- ⏳ Developer agent (code implementation)
- ⏳ Reviewer agents (multi-model review)
- ⏳ Full workflow execution

## 9. Troubleshooting

**"ANTHROPIC_API_KEY not set"**
```bash
export ANTHROPIC_API_KEY='sk-ant-...'
```

**"API error: 401"**
- Check your API key is valid
- Make sure it starts with `sk-ant-`

**Build errors**
```bash
go mod tidy
go build
```

## 10. Next Steps

1. **Try the test analyzer:** `go run test_analyzer.go`
2. **Run a simple example:** `./opensack "Add health check endpoint"`
3. **Try consensus mode:** Use a complex architectural task
4. **Experiment with threshold:** Lower it in `main.go` to use consensus more

## Tips

- **Type "approved"** to finalize the plan
- **Override mode** if you disagree with analysis
- **Check cost summary** at the end of each session
- **Start small** with simple tasks to learn the flow
