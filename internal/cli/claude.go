package cli

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/xaelophone/ralph-setup/internal/config"
)

// ClaudeCLI implements CLIRunner for Claude Code CLI
type ClaudeCLI struct {
	config config.CLIConfig
}

// NewClaudeCLI creates a new Claude CLI runner
func NewClaudeCLI(cfg config.CLIConfig) *ClaudeCLI {
	return &ClaudeCLI{config: cfg}
}

// Name returns the CLI name
func (c *ClaudeCLI) Name() string {
	return "claude"
}

// SupportsStreamJSON returns true as Claude supports JSONL streaming
func (c *ClaudeCLI) SupportsStreamJSON() bool {
	return true
}

// BuildCommand creates an exec.Cmd for Claude CLI
func (c *ClaudeCLI) BuildCommand(prompt string, workDir string) *exec.Cmd {
	// Determine command path
	cmdPath := c.config.Command
	if cmdPath == "" {
		cmdPath = "claude"
	}

	// Build arguments
	args := []string{
		"--dangerously-skip-permissions",
		"--output-format", "stream-json",
		"--verbose",
	}

	// Add model if specified
	if c.config.Model != "" {
		args = append(args, "--model", c.config.Model)
	}

	// Add any extra arguments
	args = append(args, c.config.ExtraArgs...)

	cmd := exec.Command(cmdPath, args...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	return cmd
}

// ClaudeEvent represents a Claude streaming JSON event
type ClaudeEvent struct {
	Type       string       `json:"type"`
	Subtype    string       `json:"subtype,omitempty"`
	Message    *ClaudeMsg   `json:"message,omitempty"`
	ToolUse    *ClaudeTool  `json:"tool_use,omitempty"`
	ToolResult *ClaudeResult `json:"tool_result,omitempty"`
	Content    string       `json:"content,omitempty"`
	Error      string       `json:"error,omitempty"`
}

// ClaudeMsg represents an assistant message
type ClaudeMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ClaudeTool represents a tool invocation
type ClaudeTool struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ClaudeResult represents a tool result
type ClaudeResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// ParseEvent parses a Claude JSONL line into a normalized event
func (c *ClaudeCLI) ParseEvent(line string) (*NormalizedEvent, error) {
	var event ClaudeEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, err
	}

	normalized := &NormalizedEvent{
		Timestamp: time.Now(),
		Raw:       event,
	}

	switch event.Type {
	case "assistant":
		normalized.Type = EventTypeMessage
		if event.Message != nil {
			normalized.Content = event.Message.Content
		}

	case "tool_use":
		normalized.Type = EventTypeToolStart
		if event.ToolUse != nil {
			normalized.ToolID = event.ToolUse.ID
			normalized.ToolName = event.ToolUse.Name
			normalized.ToolInput = event.ToolUse.Input
		}

	case "tool_result":
		normalized.Type = EventTypeToolEnd
		if event.ToolResult != nil {
			normalized.ToolID = event.ToolResult.ToolUseID
			normalized.Content = event.ToolResult.Content
			normalized.IsError = event.ToolResult.IsError
		}

	case "error":
		normalized.Type = EventTypeError
		normalized.Content = event.Error
		normalized.IsError = true

	default:
		normalized.Type = EventTypeUnknown
		normalized.Content = event.Content
	}

	return normalized, nil
}
