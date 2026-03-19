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

// BedrockAgent implements the Agent interface for AWS Bedrock
type BedrockAgent struct {
	bearerToken         string
	model               string
	temperature         float64
	conversationHistory []Message
	approved            bool
	systemPrompt        string
}

// NewBedrockAgent creates a new Bedrock agent
func NewBedrockAgent(bearerToken, model string, temperature float64, systemPrompt string) *BedrockAgent {
	return &BedrockAgent{
		bearerToken:         bearerToken,
		model:               model,
		temperature:         temperature,
		conversationHistory: []Message{},
		approved:            false,
		systemPrompt:        systemPrompt,
	}
}

// BedrockRequest for Bedrock API (Claude format)
type BedrockRequest struct {
	AnthropicVersion string       `json:"anthropic_version"`
	MaxTokens        int          `json:"max_tokens"`
	Temperature      float64      `json:"temperature,omitempty"`
	System           string       `json:"system,omitempty"`
	Messages         []BedrockMsg `json:"messages"`
}

type BedrockMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type BedrockResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

// Chat sends a message and gets a response
func (b *BedrockAgent) Chat(ctx context.Context, message string) (Response, error) {
	// Add user message to history
	b.conversationHistory = append(b.conversationHistory, Message{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// Check if user said "approved"
	if strings.Contains(strings.ToLower(message), "approved") {
		b.approved = true
	}

	// Convert history to Bedrock format
	messages := make([]BedrockMsg, len(b.conversationHistory))
	for i, msg := range b.conversationHistory {
		messages[i] = BedrockMsg{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Call Bedrock API
	request := BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		MaxTokens:        4096,
		Temperature:      b.temperature,
		System:           b.systemPrompt,
		Messages:         messages,
	}

	responseText, err := b.callBedrockAPI(ctx, request)
	if err != nil {
		return Response{}, fmt.Errorf("failed to call Bedrock API: %w", err)
	}

	// Add assistant response to history
	b.conversationHistory = append(b.conversationHistory, Message{
		Role:      "assistant",
		Content:   responseText,
		Timestamp: time.Now(),
	})

	return Response{
		Message:  responseText,
		Approved: b.approved,
	}, nil
}

// GeneratePlan generates detailed implementation plan after approval
func (b *BedrockAgent) GeneratePlan(ctx context.Context) (ImplementationPlan, error) {
	if !b.approved {
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

	response, err := b.Chat(ctx, planPrompt)
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
			ConversationSummary: b.summarizeConversation(),
		}
	}

	return plan, nil
}

// GetConversationHistory returns the conversation history
func (b *BedrockAgent) GetConversationHistory() []Message {
	return b.conversationHistory
}

// IsApproved returns whether the plan is approved
func (b *BedrockAgent) IsApproved() bool {
	return b.approved
}

// callBedrockAPI makes the actual API call
func (b *BedrockAgent) callBedrockAPI(ctx context.Context, request BedrockRequest) (string, error) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", err
	}

	// Use Bedrock endpoint with model ID in path
	endpoint := fmt.Sprintf("https://bedrock-runtime.ap-northeast-2.amazonaws.com/model/%s/invoke", b.model)
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.bearerToken)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Bedrock API error: %s - %s", resp.Status, string(body))
	}

	var apiResp BedrockResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w - body: %s", err, string(body))
	}

	if len(apiResp.Content) == 0 {
		return "", fmt.Errorf("empty response from Bedrock API - body: %s", string(body))
	}

	return apiResp.Content[0].Text, nil
}

// summarizeConversation creates a summary of the conversation
func (b *BedrockAgent) summarizeConversation() string {
	var summary strings.Builder
	summary.WriteString("Conversation Summary:\n\n")

	for _, msg := range b.conversationHistory {
		summary.WriteString(fmt.Sprintf("[%s] %s: %s\n\n",
			msg.Timestamp.Format("15:04:05"),
			msg.Role,
			truncate(msg.Content, 200)))
	}

	return summary.String()
}
