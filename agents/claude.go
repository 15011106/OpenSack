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

The plan must be PRESCRIPTIVE and include:
1. Exact files to modify/create
2. Exact functions to add/modify with step-by-step logic
3. Security boundaries and validations
4. Test requirements
5. Decision log with rationales

Format as JSON matching the ImplementationPlan structure.`

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
		// If JSON parsing fails, create a basic plan from the text
		plan = ImplementationPlan{
			Summary:             "Implementation plan",
			ConversationSummary: c.summarizeConversation(),
		}
	}

	return plan, nil
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
