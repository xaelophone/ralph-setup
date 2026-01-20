package orchestrator

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/xaelophone/ralph-setup/internal/cli"
	"github.com/xaelophone/ralph-setup/internal/config"
)

// Config holds orchestrator configuration
type Config struct {
	MaxIterations int
	RestartDelay  time.Duration
	ContextLines  int
	LogDir        string
	SessionFile   string
	LockFile      string
	CLIConfig     config.CLIConfig // CLI backend configuration
}

// DefaultConfig returns default orchestrator configuration
func DefaultConfig() Config {
	return Config{
		MaxIterations: 100,
		RestartDelay:  3 * time.Second,
		ContextLines:  20,
		LogDir:        ".ralph-logs",
		SessionFile:   ".ralph-session.json",
		LockFile:      ".ralph.lock",
		CLIConfig:     config.DefaultCLIConfig(),
	}
}

// Orchestrator manages the Claude loop
type Orchestrator struct {
	config  Config
	program *tea.Program
	session *Session
	mu      sync.Mutex

	// CLI backend
	cliRunner cli.CLIRunner

	// Current iteration state
	currentSubagents []SubagentTrace
	outputBuffer     strings.Builder

	// Process management
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	running bool
	stopCh  chan struct{}
}

// New creates a new orchestrator
func New(config Config, program *tea.Program) *Orchestrator {
	return &Orchestrator{
		config:    config,
		program:   program,
		stopCh:    make(chan struct{}),
		cliRunner: cli.NewCLIRunner(config.CLIConfig),
	}
}

// Messages sent to the TUI

// OutputMsg contains raw output from Claude
type OutputMsg struct {
	Content string
	Raw     bool // If true, this is raw output; if false, it's parsed
}

// EventMsg contains a parsed Claude event
type EventMsg struct {
	Event ClaudeEvent
}

// SubagentMsg contains subagent activity
type SubagentMsg struct {
	Trace SubagentTrace
}

// StatusMsg contains orchestrator status updates
type StatusMsg struct {
	Iteration      int
	Status         string
	CurrentTask    string
	TasksCompleted int
	TasksRemaining int
}

// CompletionMsg is sent when a task is completed
type CompletionMsg struct {
	Status IterationStatus
	Task   string
}

// SessionMsg contains session state
type SessionMsg struct {
	Session *Session
}

// ErrorMsg contains error information
type ErrorMsg struct {
	Error error
}

// StartedMsg is sent when the orchestrator starts
type StartedMsg struct{}

// StoppedMsg is sent when the orchestrator stops
type StoppedMsg struct {
	Reason string
}

// Start begins the orchestration loop
func (o *Orchestrator) Start() error {
	o.mu.Lock()
	if o.running {
		o.mu.Unlock()
		return fmt.Errorf("orchestrator already running")
	}
	o.running = true
	o.mu.Unlock()

	// Initialize session
	if err := o.initSession(); err != nil {
		return err
	}

	o.program.Send(StartedMsg{})
	o.program.Send(SessionMsg{Session: o.session})

	// Run the main loop
	go o.runLoop()

	return nil
}

// Stop stops the orchestrator
func (o *Orchestrator) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.running {
		return
	}

	close(o.stopCh)
	o.running = false

	if o.cmd != nil && o.cmd.Process != nil {
		o.cmd.Process.Kill()
	}

	if o.session != nil {
		o.session.Status = SessionStatusInterrupted
		o.saveSession()
	}
}

// initSession creates or resumes a session
func (o *Orchestrator) initSession() error {
	// Check for existing lock
	if _, err := os.Stat(o.config.LockFile); err == nil {
		// Lock exists - check if stale
		lockData, _ := os.ReadFile(o.config.LockFile)
		var lock struct {
			PID int `json:"pid"`
		}
		json.Unmarshal(lockData, &lock)

		// Check if process is still running
		if lock.PID > 0 {
			process, err := os.FindProcess(lock.PID)
			if err == nil {
				// On Unix, FindProcess always succeeds. Use signal 0 to check.
				if err := process.Signal(os.Signal(nil)); err == nil {
					return fmt.Errorf("another ralph-loop is running (PID: %d)", lock.PID)
				}
			}
		}
		// Stale lock - clean it up
		os.Remove(o.config.LockFile)
	}

	// Create new session
	o.session = &Session{
		ID:         uuid.New().String(),
		StartedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Status:     SessionStatusRunning,
		Iteration:  0,
		WorkingDir: mustGetwd(),
		PID:        os.Getpid(),
	}

	// Create lock file
	lockData, _ := json.Marshal(map[string]interface{}{
		"pid":        os.Getpid(),
		"session_id": o.session.ID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
	os.WriteFile(o.config.LockFile, lockData, 0644)

	// Create log directory
	os.MkdirAll(o.config.LogDir, 0755)

	return o.saveSession()
}

// saveSession persists the session state
func (o *Orchestrator) saveSession() error {
	o.session.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(o.session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(o.config.SessionFile, data, 0644)
}

// runLoop is the main orchestration loop
func (o *Orchestrator) runLoop() {
	defer func() {
		os.Remove(o.config.LockFile)
		o.program.Send(StoppedMsg{Reason: "loop ended"})
	}()

	consecutiveFailures := 0

	for o.session.Iteration < o.config.MaxIterations {
		select {
		case <-o.stopCh:
			return
		default:
		}

		o.session.Iteration++
		o.saveSession()

		// Check if we should continue
		shouldContinue, currentTask, tasksRemaining := o.checkTasks()
		if !shouldContinue {
			o.session.Status = SessionStatusCompleted
			o.saveSession()
			return
		}

		o.session.CurrentTask = currentTask
		o.saveSession()

		// Send status update
		o.program.Send(StatusMsg{
			Iteration:      o.session.Iteration,
			Status:         "running",
			CurrentTask:    currentTask,
			TasksCompleted: o.session.TasksCompleted,
			TasksRemaining: tasksRemaining,
		})

		// Run Claude iteration
		result := o.runIteration()

		switch result.Status {
		case IterationStatusComplete:
			o.session.TasksCompleted++
			consecutiveFailures = 0
			o.program.Send(CompletionMsg{Status: result.Status, Task: result.Task})

		case IterationStatusBlocked:
			o.writeHandoff(currentTask, result.LogFile)
			consecutiveFailures = 0
			o.program.Send(CompletionMsg{Status: result.Status, Task: result.Task})

		case IterationStatusFailed:
			consecutiveFailures++
			if consecutiveFailures >= 3 {
				o.session.Status = SessionStatusFailed
				o.saveSession()
				o.program.Send(ErrorMsg{Error: fmt.Errorf("too many consecutive failures")})
				return
			}
		}

		// Brief delay before next iteration
		time.Sleep(o.config.RestartDelay)
	}

	o.session.Status = SessionStatusCompleted
	o.saveSession()
}

// runIteration runs a single Claude iteration
func (o *Orchestrator) runIteration() IterationResult {
	startTime := time.Now()
	result := IterationResult{
		Iteration: o.session.Iteration,
		Task:      o.session.CurrentTask,
	}

	// Reset iteration state
	o.currentSubagents = nil
	o.outputBuffer.Reset()

	// Build the prompt
	prompt := o.buildPrompt()

	// Create log file
	logFile := filepath.Join(o.config.LogDir, fmt.Sprintf("iteration-%d.log", o.session.Iteration))
	result.LogFile = logFile
	logWriter, err := os.Create(logFile)
	if err != nil {
		result.Status = IterationStatusFailed
		return result
	}
	defer logWriter.Close()

	// Run CLI with streaming JSON output
	o.cmd = o.cliRunner.BuildCommand(prompt, o.session.WorkingDir)

	// Set up pipes
	stdin, _ := o.cmd.StdinPipe()
	stdout, _ := o.cmd.StdoutPipe()
	stderr, _ := o.cmd.StderrPipe()

	if err := o.cmd.Start(); err != nil {
		result.Status = IterationStatusFailed
		return result
	}

	// Send prompt via stdin
	stdin.Write([]byte(prompt))
	stdin.Close()

	// Parse stdout (JSONL)
	completionDetected := false
	blockedDetected := false

	var wg sync.WaitGroup
	wg.Add(2)

	// Read stdout (JSONL)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		// Increase buffer size for long lines
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		for scanner.Scan() {
			line := scanner.Text()
			logWriter.WriteString(line + "\n")

			// Try to parse using CLI runner
			if normalizedEvent, err := o.cliRunner.ParseEvent(line); err == nil && normalizedEvent != nil {
				o.processNormalizedEvent(normalizedEvent)

				// Check for completion tokens in content
				if normalizedEvent.Type == cli.EventTypeMessage {
					if cli.ContainsCompletionToken(normalizedEvent.Content) {
						completionDetected = true
					}
					if cli.ContainsBlockedToken(normalizedEvent.Content) {
						blockedDetected = true
					}
				}
			} else {
				// Raw output
				o.program.Send(OutputMsg{Content: line, Raw: true})

				// Check raw output for tokens
				if cli.ContainsCompletionToken(line) {
					completionDetected = true
				}
				if cli.ContainsBlockedToken(line) {
					blockedDetected = true
				}
			}
		}
	}()

	// Read stderr
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			logWriter.WriteString("[stderr] " + line + "\n")
			o.program.Send(OutputMsg{Content: "[stderr] " + line, Raw: true})
		}
	}()

	wg.Wait()
	o.cmd.Wait()

	result.Duration = time.Since(startTime)
	result.Subagents = o.currentSubagents

	if completionDetected {
		result.Status = IterationStatusComplete
	} else if blockedDetected {
		result.Status = IterationStatusBlocked
	} else {
		result.Status = IterationStatusFailed
	}

	return result
}

// processNormalizedEvent handles a normalized CLI event (works with any CLI backend)
func (o *Orchestrator) processNormalizedEvent(event *cli.NormalizedEvent) {
	switch event.Type {
	case cli.EventTypeMessage:
		o.program.Send(OutputMsg{Content: event.Content, Raw: false})

	case cli.EventTypeToolStart:
		trace := SubagentTrace{
			ID:        event.ToolID,
			Type:      event.ToolName,
			Status:    SubagentStatusRunning,
			StartedAt: event.Timestamp,
			Input:     extractToolInputSummary(event.ToolInput),
		}
		o.currentSubagents = append(o.currentSubagents, trace)
		o.program.Send(SubagentMsg{Trace: trace})

	case cli.EventTypeToolEnd:
		o.completeSubagent(event.ToolID, event.Content, event.IsError)

	case cli.EventTypeError:
		o.program.Send(OutputMsg{Content: "[error] " + event.Content, Raw: true})
	}
}

// completeSubagent marks a subagent trace as complete
func (o *Orchestrator) completeSubagent(toolID, output string, isError bool) {
	for i := range o.currentSubagents {
		if o.currentSubagents[i].ID == toolID {
			now := time.Now()
			o.currentSubagents[i].EndedAt = &now
			o.currentSubagents[i].Duration = now.Sub(o.currentSubagents[i].StartedAt)
			o.currentSubagents[i].Output = truncate(output, 200)

			if isError {
				o.currentSubagents[i].Status = SubagentStatusError
			} else {
				o.currentSubagents[i].Status = SubagentStatusComplete
			}

			o.program.Send(SubagentMsg{Trace: o.currentSubagents[i]})
			return
		}
	}
}

// extractToolInputSummary extracts a human-readable summary from tool input
func extractToolInputSummary(input map[string]interface{}) string {
	if input == nil {
		return ""
	}

	// Check common input fields in priority order
	if cmd, ok := input["command"].(string); ok {
		return truncate(cmd, 100)
	}
	if path, ok := input["file_path"].(string); ok {
		return path
	}
	if prompt, ok := input["prompt"].(string); ok {
		return truncate(prompt, 100)
	}
	return ""
}

// buildPrompt creates the prompt for Claude
func (o *Orchestrator) buildPrompt() string {
	recentProgress := o.getRecentProgress()

	return fmt.Sprintf(`You are running under ralph-loop (iteration %d).

## Current Task
%s

## Recent Progress (Last Few Iterations)
%s

## Instructions
1. Complete the current task (or the highest-priority 🤖 task in PRD.md)
2. Run tests and type checks - they MUST pass
3. Update PRD.md to mark the task complete (- [x])
4. Append to progress.txt with what you did
5. Commit with a descriptive message
6. Output the completion token: <promise>COMPLETE</promise>

IMPORTANT: You MUST output <promise>COMPLETE</promise> after completing each task.
This signals the orchestrator to continue to the next task.

If you encounter an error you cannot resolve after 3 attempts, explain the issue
and output <promise>BLOCKED</promise> instead.

Begin working on the task now.
`, o.session.Iteration, o.session.CurrentTask, recentProgress)
}

// checkTasks reads PRD.md and determines if we should continue
func (o *Orchestrator) checkTasks() (shouldContinue bool, currentTask string, remaining int) {
	data, err := os.ReadFile("PRD.md")
	if err != nil {
		// No PRD.md - Claude should create one
		return true, "Create PRD.md with task list", 0
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	incompletePattern := regexp.MustCompile(`^- \[ \] (.+)$`)
	aiPattern := regexp.MustCompile(`🤖`)

	var aiTasks []string

	for _, line := range lines {
		if match := incompletePattern.FindStringSubmatch(line); match != nil {
			task := match[1]
			if aiPattern.MatchString(line) {
				aiTasks = append(aiTasks, strings.ReplaceAll(task, "🤖", ""))
			}
		}
	}

	if len(aiTasks) == 0 {
		return false, "", 0
	}

	return true, strings.TrimSpace(aiTasks[0]), len(aiTasks)
}

// getRecentProgress reads recent entries from progress.txt
func (o *Orchestrator) getRecentProgress() string {
	data, err := os.ReadFile("progress.txt")
	if err != nil {
		return "No previous progress recorded."
	}

	lines := strings.Split(string(data), "\n")
	start := len(lines) - o.config.ContextLines
	if start < 0 {
		start = 0
	}

	recent := lines[start:]
	result := strings.Join(recent, "\n")

	// Limit size
	if len(result) > 4000 {
		result = result[:4000] + "\n... (truncated)"
	}

	return result
}

// writeHandoff writes blocked task info to HANDOFF.md
func (o *Orchestrator) writeHandoff(task, logFile string) {
	f, err := os.OpenFile("HANDOFF.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf("\n## Blocked Task (%s)\n", time.Now().Format("2006-01-02 15:04")))
	f.WriteString(fmt.Sprintf("- Task: %s\n", task))
	f.WriteString(fmt.Sprintf("- See log: %s\n", logFile))
}

// GetSession returns the current session
func (o *Orchestrator) GetSession() *Session {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.session
}

// Helper functions

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
