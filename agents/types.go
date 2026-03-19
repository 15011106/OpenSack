package agents

import (
	"context"
	"time"
)

// Agent interface - all agents implement this
type Agent interface {
	Chat(ctx context.Context, message string) (Response, error)
	GeneratePlan(ctx context.Context) (ImplementationPlan, error)
	GetConversationHistory() []Message
	IsApproved() bool
}

// Message represents a conversation message
type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Response from an agent
type Response struct {
	Message  string `json:"message"`
	Approved bool   `json:"approved"`
}

// ImplementationPlan is the detailed plan after approval
type ImplementationPlan struct {
	Summary             string             `json:"summary"`
	Phases              []Phase            `json:"phases"`
	Files               []FileSpec         `json:"files"`
	SecurityBoundaries  []SecurityBoundary `json:"security_boundaries"`
	TestRequirements    []TestRequirement  `json:"test_requirements"`
	DecisionLog         []Decision         `json:"decision_log"`
	ConversationSummary string             `json:"conversation_summary"`
}

// Phase represents a development phase
type Phase struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Files       []string `json:"files"`
	Order       int      `json:"order"`
}

// FileSpec defines what needs to be implemented in a file
type FileSpec struct {
	Path        string     `json:"path"`
	Action      string     `json:"action"` // "create", "modify", "delete"
	Description string     `json:"description"`
	Functions   []Function `json:"functions"`
	Structs     []Struct   `json:"structs"`
}

// Function specification
type Function struct {
	Name          string   `json:"name"`
	Signature     string   `json:"signature"`
	Description   string   `json:"description"`
	Steps         []string `json:"steps"`
	ErrorHandling string   `json:"error_handling"`
}

// Struct specification
type Struct struct {
	Name        string  `json:"name"`
	Fields      []Field `json:"fields"`
	Description string  `json:"description"`
}

// Field in a struct
type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

// SecurityBoundary marks where security validation is required
type SecurityBoundary struct {
	Location    string   `json:"location"`
	Description string   `json:"description"`
	Validations []string `json:"validations"`
}

// TestRequirement specifies what tests are needed
type TestRequirement struct {
	Type        string `json:"type"` // "unit", "integration", "functional"
	Description string `json:"description"`
	File        string `json:"file"`
}

// Decision captures architectural decisions
type Decision struct {
	Question     string   `json:"question"`
	Decision     string   `json:"decision"`
	Rationale    string   `json:"rationale"`
	Alternatives []string `json:"alternatives"`
	Tradeoffs    string   `json:"tradeoffs"`
}

// Discovery document from phase 1
type Discovery struct {
	Goal                    string                 `json:"goal"`
	Requirements            []string               `json:"requirements"`
	Constraints             []string               `json:"constraints"`
	Risks                   []string               `json:"risks"`
	Assumptions             []string               `json:"assumptions"`
	UserStories             []UserStory            `json:"user_stories"`
	AcceptanceCriteria      []string               `json:"acceptance_criteria"`
	SecurityRequirements    []string               `json:"security_requirements"`
	PerformanceRequirements []string               `json:"performance_requirements"`
	ScaleRequirements       string                 `json:"scale_requirements"`
	EstimatedComplexity     string                 `json:"estimated_complexity"` // "low", "medium", "high"
	Metadata                map[string]interface{} `json:"metadata"`
}

// UserStory represents a user story
type UserStory struct {
	AsA    string `json:"as_a"`
	IWant  string `json:"i_want"`
	SoThat string `json:"so_that"`
}

// ReviewResult from reviewers
type ReviewResult struct {
	Model          string   `json:"model"`
	Summary        string   `json:"summary"`
	CriticalIssues []Issue  `json:"critical_issues"`
	MinorIssues    []Issue  `json:"minor_issues"`
	Suggestions    []string `json:"suggestions"`
	Approved       bool     `json:"approved"`
	QualityScore   float64  `json:"quality_score"` // 0-100
}

// Issue found during review
type Issue struct {
	Location    string `json:"location"`
	Severity    string `json:"severity"` // "critical", "major", "minor"
	Description string `json:"description"`
	Suggestion  string `json:"suggestion"`
}

// ImplementationResult from developer
type ImplementationResult struct {
	FilesCreated  []string `json:"files_created"`
	FilesModified []string `json:"files_modified"`
	TestsWritten  []string `json:"tests_written"`
	Summary       string   `json:"summary"`
	Diff          string   `json:"diff"`
}
