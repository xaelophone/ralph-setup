package cli

import (
	"os/exec"
	"time"

	"github.com/xaelophone/ralph-setup/internal/config"
)

// CLIRunner defines the interface for CLI backends
type CLIRunner interface {
	// BuildCommand creates an exec.Cmd configured for the CLI
	BuildCommand(prompt string, workDir string) *exec.Cmd

	// ParseEvent parses a JSONL event line into a normalized event
	// Returns nil if the line is not a valid event
	ParseEvent(line string) (*NormalizedEvent, error)

	// Name returns the CLI name for display purposes
	Name() string

	// SupportsStreamJSON returns whether this CLI supports JSONL streaming output
	SupportsStreamJSON() bool
}

// NormalizedEvent represents a CLI event normalized across different backends
type NormalizedEvent struct {
	Type      EventType              // Normalized event type
	Content   string                 // Text content (for messages)
	ToolName  string                 // Tool/command name (for tool events)
	ToolID    string                 // Tool invocation ID
	ToolInput map[string]interface{} // Tool input parameters
	IsError   bool                   // Whether this represents an error
	Raw       interface{}            // Original event for debugging
	Timestamp time.Time              // Event timestamp
}

// EventType represents normalized event types across CLIs
type EventType string

const (
	// EventTypeMessage is emitted when the assistant produces text output
	EventTypeMessage EventType = "message"

	// EventTypeToolStart is emitted when a tool begins execution
	EventTypeToolStart EventType = "tool_start"

	// EventTypeToolEnd is emitted when a tool completes execution
	EventTypeToolEnd EventType = "tool_end"

	// EventTypeTurnComplete is emitted when a full turn completes
	EventTypeTurnComplete EventType = "turn_complete"

	// EventTypeError is emitted for errors
	EventTypeError EventType = "error"

	// EventTypeUnknown is for unrecognized events
	EventTypeUnknown EventType = "unknown"
)

// NewCLIRunner creates a CLIRunner for the specified backend
func NewCLIRunner(cfg config.CLIConfig) CLIRunner {
	switch cfg.Backend {
	case config.CLIBackendCodex:
		return NewCodexCLI(cfg)
	case config.CLIBackendClaude:
		fallthrough
	default:
		return NewClaudeCLI(cfg)
	}
}

// ContainsCompletionToken checks if text contains the completion token
func ContainsCompletionToken(text string) bool {
	return containsToken(text, "<promise>COMPLETE</promise>")
}

// ContainsBlockedToken checks if text contains the blocked token
func ContainsBlockedToken(text string) bool {
	return containsToken(text, "<promise>BLOCKED</promise>")
}

func containsToken(text, token string) bool {
	return len(text) >= len(token) && (text == token ||
		(len(text) > len(token) && (text[:len(token)] == token ||
		text[len(text)-len(token):] == token ||
		contains(text, token))))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
