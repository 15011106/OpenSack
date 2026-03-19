package agents

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ClaudeAgent implements the Agent interface
type ClaudeAgent struct {
	apiKey              string
	model               string
	temperature         float64
	conversationHistory []Message
	approved            bool
	systemPrompt        string
}

// NewClaudeAgent creates a new Claude agent
func NewClaudeAgent(apiKey, model string, temperature float64, systemPrompt string) *ClaudeAgent {
	return &ClaudeAgent{
		apiKey:              apiKey,
		model:               model,
		temperature:         temperature,
		conversationHistory: []Message{},
		approved:            false,
		systemPrompt:        systemPrompt,
	}
}

// AnthropicRequest for Claude API
type AnthropicRequest struct {
	Model       string         `json:"model"`
	MaxTokens   int            `json:"max_tokens"`
	Temperature float64        `json:"temperature"`
	System      string         `json:"system"`
	Messages    []AnthropicMsg `json:"messages"`
}

type AnthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AnthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

// Chat sends a message and gets a response
func (c *ClaudeAgent) Chat(ctx context.Context, message string) (Response, error) {
	// Add user message to history
	c.conversationHistory = append(c.conversationHistory, Message{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// Check if user said "approved"
	if strings.Contains(strings.ToLower(message), "approved") {
		c.approved = true
	}

	// Convert history to Anthropic format
	messages := make([]AnthropicMsg, len(c.conversationHistory))
	for i, msg := range c.conversationHistory {
		messages[i] = AnthropicMsg{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call Claude API
	request := AnthropicRequest{
		Model:       c.model,
		MaxTokens:   4096,
		Temperature: c.temperature,
		System:      c.systemPrompt,
		Messages:    messages,
	}

	responseText, err := c.callClaudeAPI(ctx, request)
	if err != nil {
		return Response{}, fmt.Errorf("failed to call Claude API: %w", err)
	}

	// Add assistant response to history
	c.conversationHistory = append(c.conversationHistory, Message{
		Role:      "assistant",
		Content:   responseText,
		Timestamp: time.Now(),
	})

	return Response{
		Message:  responseText,
		Approved: c.approved,
	}, nil
}

// GeneratePlan generates detailed implementation plan after approval
func (c *ClaudeAgent) GeneratePlan(ctx context.Context) (ImplementationPlan, error) {
	if !c.approved {
		return ImplementationPlan{}, fmt.Errorf("plan not approved yet")
	}

	// Generate detailed plan based on conversation
	planPrompt := `Our discussion is complete and approved.
Generate a detailed implementation plan in JSON format.

CRITICAL: The developer CANNOT make any decisions. You MUST provide:
1. EXACT file paths (e.g., "cmd/main.go", "api/handler.go")
2. EXACT function signatures (e.g., "func HandleRequest(w http.ResponseWriter, r *http.Request)")
3. STEP-BY-STEP logic for each function (e.g., ["Parse request body", "Validate input", "Call database", "Return JSON"])
4. What structs/types to define with exact field names and types

EXAMPLE of what you must provide:
{
  "summary": "REST API with authentication",
  "files": [
    {
      "path": "main.go",
      "action": "create",
      "description": "Entry point",
      "functions": [
        {
          "name": "main",
          "signature": "func main()",
          "description": "Start HTTP server",
          "steps": ["Initialize router", "Register handlers", "Start server on :8080"]
        }
      ]
    }
  ]
}

Format as JSON. Wrap in json code block.`

	response, err := c.Chat(ctx, planPrompt)
	if err != nil {
		return ImplementationPlan{}, err
	}

	// Parse JSON from response
	var plan ImplementationPlan

	// Extract JSON from markdown code blocks if present
	content := response.Message
	if strings.Contains(content, "```json") {
		start := strings.Index(content, "```json") + 7
		end := strings.Index(content[start:], "```")
		if end > 0 {
			content = content[start : start+end]
		}
	}

	err = json.Unmarshal([]byte(content), &plan)
	if err != nil {
		return ImplementationPlan{}, fmt.Errorf("failed to parse plan JSON: %w\n\nResponse was:\n%s", err, response.Message[:min(len(response.Message), 500)])
	}

	// Validate plan has minimum required fields
	if len(plan.Files) == 0 {
		return ImplementationPlan{}, fmt.Errorf("plan has no files specified - architect must provide detailed file specifications")
	}

	return plan, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetConversationHistory returns the conversation history
func (c *ClaudeAgent) GetConversationHistory() []Message {
	return c.conversationHistory
}

// IsApproved returns whether the plan is approved
func (c *ClaudeAgent) IsApproved() bool {
	return c.approved
}

// callClaudeAPI makes the actual API call
func (c *ClaudeAgent) callClaudeAPI(ctx context.Context, request AnthropicRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var apiResp AnthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return apiResp.Content[0].Text, nil
}

// summarizeConversation creates a summary of the conversation
func (c *ClaudeAgent) summarizeConversation() string {
	var summary strings.Builder
	summary.WriteString("Conversation Summary:\n\n")

	for _, msg := range c.conversationHistory {
		summary.WriteString(fmt.Sprintf("[%s] %s: %s\n\n",
			msg.Timestamp.Format("15:04:05"),
			msg.Role,
			truncate(msg.Content, 200)))
	}

	return summary.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
