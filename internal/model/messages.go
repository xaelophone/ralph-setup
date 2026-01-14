package model

import "github.com/xaelophone/ralph-setup/internal/parser"

// OutputMsg contains Claude output to display
type OutputMsg struct {
	Content string
}

// ClaudeStartedMsg indicates Claude process has started
type ClaudeStartedMsg struct{}

// ClaudeStoppedMsg indicates Claude process has stopped
type ClaudeStoppedMsg struct {
	ExitCode int
	Error    error
}

// TasksUpdatedMsg indicates PRD.md has been parsed
type TasksUpdatedMsg struct {
	Tasks []parser.Task
}

// ProgressUpdatedMsg indicates progress.txt has been parsed
type ProgressUpdatedMsg struct {
	Entries []parser.ProgressEntry
}

// FileChangedMsg indicates a watched file has changed
type FileChangedMsg struct {
	File string
}

// GitUpdatedMsg indicates git commits have been refreshed
type GitUpdatedMsg struct {
	Commits []string
}

// ErrorMsg indicates an error occurred
type ErrorMsg struct {
	Error error
}
