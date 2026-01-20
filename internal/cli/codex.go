package cli

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/xaelophone/ralph-setup/internal/config"
)

// CodexCLI implements CLIRunner for OpenAI Codex CLI
type CodexCLI struct {
	config config.CLIConfig
}

// NewCodexCLI creates a new Codex CLI runner
func NewCodexCLI(cfg config.CLIConfig) *CodexCLI {
	return &CodexCLI{config: cfg}
}

// Name returns the CLI name
func (c *CodexCLI) Name() string {
	return "codex"
}

// SupportsStreamJSON returns true as Codex supports JSON output
func (c *CodexCLI) SupportsStreamJSON() bool {
	return true
}

// BuildCommand creates an exec.Cmd for Codex CLI
func (c *CodexCLI) BuildCommand(prompt string, workDir string) *exec.Cmd {
	// Determine command path
	cmdPath := c.config.Command
	if cmdPath == "" {
		cmdPath = "codex"
	}

	// Build arguments for codex exec
	args := []string{
		"exec",
		"--json",
		"--dangerously-bypass-approvals-and-sandbox",
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

// CodexEvent represents a Codex streaming JSON event
// Codex uses a different event structure than Claude
type CodexEvent struct {
	Type      string          `json:"type"`
	SessionID string          `json:"session_id,omitempty"`
	Item      *CodexItem      `json:"item,omitempty"`
	Turn      *CodexTurn      `json:"turn,omitempty"`
	Error     *CodexError     `json:"error,omitempty"`
}

// CodexItem represents an item in Codex output
type CodexItem struct {
	Type    string       `json:"type"` // agent_message, command_execution, etc.
	ID      string       `json:"id,omitempty"`
	Status  string       `json:"status,omitempty"` // started, completed, etc.
	Content *CodexContent `json:"content,omitempty"`
	Command *CodexCommand `json:"command,omitempty"`
	Output  string        `json:"output,omitempty"`
}

// CodexContent represents message content
type CodexContent struct {
	Text string `json:"text,omitempty"`
}

// CodexCommand represents a command execution
type CodexCommand struct {
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`
}

// CodexTurn represents turn information
type CodexTurn struct {
	Status string `json:"status,omitempty"`
}

// CodexError represents an error
type CodexError struct {
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

// ParseEvent parses a Codex JSONL line into a normalized event
func (c *CodexCLI) ParseEvent(line string) (*NormalizedEvent, error) {
	var event CodexEvent
	if err := json.Unmarshal([]byte(line), &event); err != nil {
		return nil, err
	}

	normalized := &NormalizedEvent{
		Timestamp: time.Now(),
		Raw:       event,
		Type:      EventTypeUnknown,
	}

	switch event.Type {
	case "item.started":
		c.parseItemStarted(event.Item, normalized)

	case "item.completed":
		c.parseItemCompleted(event.Item, normalized)

	case "turn.completed":
		normalized.Type = EventTypeTurnComplete

	case "error":
		normalized.Type = EventTypeError
		normalized.IsError = true
		if event.Error != nil {
			normalized.Content = event.Error.Message
		}
	}

	return normalized, nil
}

// parseItemStarted handles item.started events
func (c *CodexCLI) parseItemStarted(item *CodexItem, normalized *NormalizedEvent) {
	if item == nil {
		return
	}

	if item.Type == "command_execution" || item.Type == "tool_call" {
		normalized.Type = EventTypeToolStart
		normalized.ToolID = item.ID
		if item.Command != nil {
			normalized.ToolName = item.Command.Name
			normalized.ToolInput = item.Command.Input
		}
	}
}

// parseItemCompleted handles item.completed events
func (c *CodexCLI) parseItemCompleted(item *CodexItem, normalized *NormalizedEvent) {
	if item == nil {
		return
	}

	switch item.Type {
	case "agent_message":
		normalized.Type = EventTypeMessage
		if item.Content != nil {
			normalized.Content = item.Content.Text
		}

	case "command_execution", "tool_call":
		normalized.Type = EventTypeToolEnd
		normalized.ToolID = item.ID
		normalized.Content = item.Output
		normalized.IsError = item.Status == "error"
	}
}
