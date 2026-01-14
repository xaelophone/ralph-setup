package runner

import (
	"io"
	"os"
	"os/exec"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/creack/pty"
)

// Runner manages the Claude process
type Runner struct {
	args    []string
	program *tea.Program
	cmd     *exec.Cmd
	ptmx    *os.File
	mu      sync.Mutex
	running bool
}

// New creates a new Runner
func New(args []string, program *tea.Program) *Runner {
	return &Runner{
		args:    args,
		program: program,
	}
}

// OutputMsg is sent when Claude produces output
type OutputMsg struct {
	Content string
}

// StartedMsg is sent when Claude starts
type StartedMsg struct{}

// StoppedMsg is sent when Claude stops
type StoppedMsg struct {
	ExitCode int
	Error    error
}

// Start starts the Claude process with a PTY
func (r *Runner) Start() error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return nil
	}

	// Find claude binary
	claudePath, err := exec.LookPath("claude")
	if err != nil {
		r.mu.Unlock()
		r.program.Send(StoppedMsg{ExitCode: 1, Error: err})
		return err
	}

	// Build command
	r.cmd = exec.Command(claudePath, r.args...)

	// Start with PTY for proper terminal emulation
	r.ptmx, err = pty.Start(r.cmd)
	if err != nil {
		r.mu.Unlock()
		r.program.Send(StoppedMsg{ExitCode: 1, Error: err})
		return err
	}

	r.running = true
	r.mu.Unlock()

	// Notify that Claude has started
	r.program.Send(StartedMsg{})

	// Read output in goroutine
	go r.readOutput()

	// Handle stdin pass-through
	go r.handleStdin()

	// Wait for process to complete
	go r.waitForCompletion()

	return nil
}

// readOutput reads from the PTY and sends to the program
func (r *Runner) readOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := r.ptmx.Read(buf)
		if err != nil {
			if err != io.EOF {
				// Log error but don't crash
			}
			return
		}
		if n > 0 {
			// Send output to the TUI
			r.program.Send(OutputMsg{Content: string(buf[:n])})
		}
	}
}

// handleStdin passes stdin to Claude
func (r *Runner) handleStdin() {
	// Read from original stdin and write to PTY
	// Note: In a full TUI, we'd need a more sophisticated approach
	// to handle keyboard input that shouldn't go to Claude
	io.Copy(r.ptmx, os.Stdin)
}

// waitForCompletion waits for the process to finish
func (r *Runner) waitForCompletion() {
	err := r.cmd.Wait()

	r.mu.Lock()
	r.running = false
	if r.ptmx != nil {
		r.ptmx.Close()
	}
	r.mu.Unlock()

	exitCode := 0
	if r.cmd.ProcessState != nil {
		exitCode = r.cmd.ProcessState.ExitCode()
	}

	r.program.Send(StoppedMsg{ExitCode: exitCode, Error: err})
}

// Stop stops the Claude process
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	if r.cmd != nil && r.cmd.Process != nil {
		r.cmd.Process.Kill()
	}

	if r.ptmx != nil {
		r.ptmx.Close()
	}

	r.running = false
}

// IsRunning returns whether Claude is running
func (r *Runner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.running
}

// SendInput sends input to Claude
func (r *Runner) SendInput(input string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running || r.ptmx == nil {
		return nil
	}

	_, err := r.ptmx.WriteString(input)
	return err
}
