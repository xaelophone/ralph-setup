package orchestrator

import (
	"time"
)

// ClaudeEvent represents a line of streaming JSON output from Claude
type ClaudeEvent struct {
	Type      string            `json:"type"`
	Subtype   string            `json:"subtype,omitempty"`
	Message   *Message          `json:"message,omitempty"`
	ToolUse   *ToolUse          `json:"tool_use,omitempty"`
	ToolResult *ToolResult      `json:"tool_result,omitempty"`
	Content   string            `json:"content,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
	Error     string            `json:"error,omitempty"`
}

// Message represents an assistant or user message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ToolUse represents a tool being invoked by Claude
type ToolUse struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error"`
}

// SubagentTrace represents a traced subagent call
type SubagentTrace struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"` // Task, Bash, Read, Write, Edit, etc.
	Input     string        `json:"input"`
	Status    SubagentStatus `json:"status"`
	Output    string        `json:"output,omitempty"`
	StartedAt time.Time     `json:"started_at"`
	EndedAt   *time.Time    `json:"ended_at,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
}

type SubagentStatus string

const (
	SubagentStatusPending  SubagentStatus = "pending"
	SubagentStatusRunning  SubagentStatus = "running"
	SubagentStatusComplete SubagentStatus = "complete"
	SubagentStatusError    SubagentStatus = "error"
)

// Session represents the orchestrator session state
type Session struct {
	ID              string           `json:"id"`
	StartedAt       time.Time        `json:"started_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	Status          SessionStatus    `json:"status"`
	Iteration       int              `json:"iteration"`
	TasksCompleted  int              `json:"tasks_completed"`
	CurrentTask     string           `json:"current_task,omitempty"`
	WorkingDir      string           `json:"working_dir"`
	PID             int              `json:"pid"`
	SubagentTraces  []SubagentTrace  `json:"subagent_traces,omitempty"`
}

type SessionStatus string

const (
	SessionStatusRunning     SessionStatus = "running"
	SessionStatusCompleted   SessionStatus = "completed"
	SessionStatusInterrupted SessionStatus = "interrupted"
	SessionStatusFailed      SessionStatus = "failed"
	SessionStatusRecovered   SessionStatus = "recovered"
)

// Task represents a task from PRD.md
type Task struct {
	Title     string `json:"title"`
	Complete  bool   `json:"complete"`
	IsHuman   bool   `json:"is_human"`  // 🧑 task
	IsAI      bool   `json:"is_ai"`     // 🤖 task
	Line      int    `json:"line"`
}

// IterationResult represents the result of a single iteration
type IterationResult struct {
	Iteration   int
	Status      IterationStatus
	Task        string
	Duration    time.Duration
	Subagents   []SubagentTrace
	LogFile     string
}

type IterationStatus string

const (
	IterationStatusComplete IterationStatus = "complete"
	IterationStatusBlocked  IterationStatus = "blocked"
	IterationStatusFailed   IterationStatus = "failed"
	IterationStatusTimeout  IterationStatus = "timeout"
)

// CompletionToken constants
const (
	CompletionTokenComplete = "<promise>COMPLETE</promise>"
	CompletionTokenBlocked  = "<promise>BLOCKED</promise>"
)

// PromptContext holds the context for building prompts
type PromptContext struct {
	Iteration       int
	CurrentTask     string
	RecentProgress  string
	TasksSummary    string
	SubagentHistory string
}
