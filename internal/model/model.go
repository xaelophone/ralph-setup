package model

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xaelophone/ralph-setup/internal/orchestrator"
	"github.com/xaelophone/ralph-setup/internal/parser"
	"github.com/xaelophone/ralph-setup/internal/runner"
	"github.com/xaelophone/ralph-setup/internal/theme"
)

// Re-export runner message types so main.go doesn't need to import runner
type RunnerOutputMsg = runner.OutputMsg
type RunnerStartedMsg = runner.StartedMsg
type RunnerStoppedMsg = runner.StoppedMsg

// Re-export orchestrator message types
type OrchestratorOutputMsg = orchestrator.OutputMsg
type OrchestratorStatusMsg = orchestrator.StatusMsg
type OrchestratorSubagentMsg = orchestrator.SubagentMsg
type OrchestratorCompletionMsg = orchestrator.CompletionMsg
type OrchestratorSessionMsg = orchestrator.SessionMsg
type OrchestratorStartedMsg = orchestrator.StartedMsg
type OrchestratorStoppedMsg = orchestrator.StoppedMsg
type OrchestratorErrorMsg = orchestrator.ErrorMsg

// View represents the current active view
type View int

const (
	ViewOutput View = iota
	ViewTasks
	ViewSubagents
	ViewProgress
	ViewGit
)

// Options for creating a new model
type Options struct {
	MonitorOnly    bool
	OrchestratorMode bool
	ClaudeArgs     []string
}

// Model is the main Bubbletea model
type Model struct {
	// Configuration
	monitorOnly      bool
	orchestratorMode bool
	claudeArgs       []string

	// Dimensions
	width  int
	height int

	// Views
	activeView    View
	sidebarFocus  bool
	showHelp      bool

	// Content
	claudeOutput   string
	outputViewport viewport.Model
	tasks         []parser.Task
	progressLog   []parser.ProgressEntry
	gitCommits    []string
	subagents     []orchestrator.SubagentTrace

	// State
	claudeRunning   bool
	startTime       time.Time
	projectName     string
	iteration       int
	tasksCompleted  int
	tasksRemaining  int
	currentTask     string
	sessionID       string
	lastCompletion  string

	// Theme
	theme theme.Theme

	// Runner reference (for cleanup) - legacy mode
	runner *runner.Runner

	// Orchestrator reference - new mode
	orchestrator *orchestrator.Orchestrator
}

// New creates a new model
func New(opts Options) *Model {
	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle()

	return &Model{
		monitorOnly:      opts.MonitorOnly,
		orchestratorMode: opts.OrchestratorMode,
		claudeArgs:       opts.ClaudeArgs,
		activeView:       ViewOutput,
		outputViewport:   vp,
		tasks:            []parser.Task{},
		progressLog:      []parser.ProgressEntry{},
		gitCommits:       []string{},
		subagents:        []orchestrator.SubagentTrace{},
		startTime:        time.Now(),
		projectName:      getProjectName(),
		theme:            theme.Default(),
	}
}

// SetRunner sets the Claude runner reference (legacy mode)
func (m *Model) SetRunner(r *runner.Runner) {
	m.runner = r
}

// SetOrchestrator sets the orchestrator reference (new mode)
func (m *Model) SetOrchestrator(o *orchestrator.Orchestrator) {
	m.orchestrator = o
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadTasks(),
		m.loadProgress(),
		m.watchFiles(),
	)
}

// Update handles messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateViewportSize()

	// Legacy runner messages
	case OutputMsg:
		m.claudeOutput += msg.Content
		m.outputViewport.SetContent(m.claudeOutput)
		m.outputViewport.GotoBottom()

	case runner.OutputMsg:
		m.claudeOutput += msg.Content
		m.outputViewport.SetContent(m.claudeOutput)
		m.outputViewport.GotoBottom()

	case ClaudeStartedMsg:
		m.claudeRunning = true

	case runner.StartedMsg:
		m.claudeRunning = true

	case ClaudeStoppedMsg:
		m.claudeRunning = false

	case runner.StoppedMsg:
		m.claudeRunning = false

	// Orchestrator messages
	case orchestrator.OutputMsg:
		if msg.Raw {
			m.claudeOutput += msg.Content + "\n"
		} else {
			m.claudeOutput += msg.Content + "\n"
		}
		m.outputViewport.SetContent(m.claudeOutput)
		m.outputViewport.GotoBottom()

	case orchestrator.StatusMsg:
		m.iteration = msg.Iteration
		m.currentTask = msg.CurrentTask
		m.tasksCompleted = msg.TasksCompleted
		m.tasksRemaining = msg.TasksRemaining
		m.claudeRunning = msg.Status == "running"

	case orchestrator.SubagentMsg:
		m.updateSubagent(msg.Trace)

	case orchestrator.CompletionMsg:
		m.lastCompletion = fmt.Sprintf("[%s] %s", msg.Status, msg.Task)
		cmds = append(cmds, m.loadTasks())
		cmds = append(cmds, m.loadProgress())

	case orchestrator.SessionMsg:
		if msg.Session != nil {
			m.sessionID = msg.Session.ID
			m.iteration = msg.Session.Iteration
			m.tasksCompleted = msg.Session.TasksCompleted
		}

	case orchestrator.StartedMsg:
		m.claudeRunning = true

	case orchestrator.StoppedMsg:
		m.claudeRunning = false

	case orchestrator.ErrorMsg:
		m.claudeOutput += fmt.Sprintf("\n[ERROR] %v\n", msg.Error)
		m.outputViewport.SetContent(m.claudeOutput)
		m.outputViewport.GotoBottom()

	// File watching messages
	case TasksUpdatedMsg:
		m.tasks = msg.Tasks

	case ProgressUpdatedMsg:
		m.progressLog = msg.Entries

	case FileChangedMsg:
		switch msg.File {
		case "PRD.md":
			cmds = append(cmds, m.loadTasks())
		case "progress.txt":
			cmds = append(cmds, m.loadProgress())
		}
	}

	// Update viewport
	var cmd tea.Cmd
	m.outputViewport, cmd = m.outputViewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// updateSubagent updates or adds a subagent trace
func (m *Model) updateSubagent(trace orchestrator.SubagentTrace) {
	// Find and update existing trace
	for i := range m.subagents {
		if m.subagents[i].ID == trace.ID {
			m.subagents[i] = trace
			return
		}
	}
	// Add new trace
	m.subagents = append(m.subagents, trace)

	// Keep only last 50 traces
	if len(m.subagents) > 50 {
		m.subagents = m.subagents[len(m.subagents)-50:]
	}
}

// handleKeyPress handles keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "q", "ctrl+c":
		if m.runner != nil {
			m.runner.Stop()
		}
		return m, tea.Quit

	case "?":
		m.showHelp = !m.showHelp
		return m, nil

	case "tab":
		m.sidebarFocus = !m.sidebarFocus
		return m, nil

	case "1":
		m.activeView = ViewOutput
		return m, nil
	case "2":
		m.activeView = ViewTasks
		return m, nil
	case "3":
		m.activeView = ViewSubagents
		return m, nil
	case "4":
		m.activeView = ViewProgress
		return m, nil
	case "5":
		m.activeView = ViewGit
		return m, nil

	case "esc":
		m.showHelp = false
		m.sidebarFocus = false
		return m, nil
	}

	// View-specific keys when not in sidebar
	if !m.sidebarFocus {
		switch msg.String() {
		case "j", "down":
			m.outputViewport.LineDown(1)
		case "k", "up":
			m.outputViewport.LineUp(1)
		case "g":
			m.outputViewport.GotoTop()
		case "G":
			m.outputViewport.GotoBottom()
		case "ctrl+d":
			m.outputViewport.HalfViewDown()
		case "ctrl+u":
			m.outputViewport.HalfViewUp()
		}
	}

	return m, nil
}

// View renders the model
func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	// Calculate layout
	sidebarWidth := m.calculateSidebarWidth()
	mainWidth := m.width - sidebarWidth - 1 // -1 for border

	// Render components
	statusBar := m.renderStatusBar()
	sidebar := m.renderSidebar(sidebarWidth, m.height-3)
	main := m.renderMain(mainWidth, m.height-3)

	// Compose layout
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		sidebar,
		main,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusBar,
		content,
	)
}

// renderStatusBar renders the top status bar
func (m Model) renderStatusBar() string {
	status := "● STOPPED"
	statusStyle := m.theme.StatusStopped
	if m.claudeRunning {
		status = "● RUNNING"
		statusStyle = m.theme.StatusRunning
	}
	if m.monitorOnly {
		status = "◌ MONITOR"
		statusStyle = m.theme.StatusMonitor
	}

	// Task progress
	completed := 0
	total := len(m.tasks)
	for _, t := range m.tasks {
		if t.Complete {
			completed++
		}
	}
	taskStatus := ""
	if total > 0 {
		taskStatus = m.theme.Muted.Render(
			" │ " + itoa(completed) + "/" + itoa(total) + " tasks",
		)
	}

	// Iteration info for orchestrator mode
	iterInfo := ""
	if m.orchestratorMode && m.iteration > 0 {
		iterInfo = m.theme.Muted.Render(
			" │ iter " + itoa(m.iteration),
		)
	}

	left := lipgloss.JoinHorizontal(
		lipgloss.Center,
		m.theme.Title.Render(" rwatch v2.0.0 "),
		" │ ",
		statusStyle.Render(status),
		" │ ",
		m.theme.ProjectName.Render(m.projectName),
		taskStatus,
		iterInfo,
	)

	right := m.theme.Help.Render("?=help")

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		left,
		strings.Repeat(" ", gap),
		right,
	)
}

// renderSidebar renders the sidebar navigation and current task
func (m Model) renderSidebar(width, height int) string {
	if width <= 0 {
		return ""
	}

	style := m.theme.Sidebar.Width(width).Height(height)

	// Navigation items
	items := []struct {
		view  View
		label string
		count int
	}{
		{ViewOutput, "Output", 0},
		{ViewTasks, "Tasks", len(m.tasks)},
		{ViewSubagents, "Subagents", len(m.subagents)},
		{ViewProgress, "Progress", len(m.progressLog)},
		{ViewGit, "Git", len(m.gitCommits)},
	}

	var nav strings.Builder
	nav.WriteString(m.theme.SidebarHeader.Render("Navigation") + "\n")

	for _, item := range items {
		prefix := "  ○ "
		labelStyle := m.theme.SidebarItem
		if item.view == m.activeView {
			prefix = "  ► "
			labelStyle = m.theme.SidebarItemActive
		}

		label := item.label
		if item.count > 0 {
			label += " (" + itoa(item.count) + ")"
		}

		nav.WriteString(prefix + labelStyle.Render(label) + "\n")
	}

	// Current task section
	nav.WriteString("\n")
	nav.WriteString(m.theme.SidebarHeader.Render("Current Task") + "\n")
	nav.WriteString(m.theme.Divider.Render(strings.Repeat("─", width-4)) + "\n")

	currentTask := m.getCurrentTask()
	if currentTask != nil {
		taskText := wrapText(currentTask.Title, width-4)
		nav.WriteString(m.theme.CurrentTask.Render("☐ " + taskText) + "\n")
	} else if len(m.tasks) > 0 && m.allTasksComplete() {
		nav.WriteString(m.theme.Success.Render("✓ All done!") + "\n")
	} else {
		nav.WriteString(m.theme.Muted.Render("No PRD.md") + "\n")
	}

	return style.Render(nav.String())
}

// renderMain renders the main content area
func (m Model) renderMain(width, height int) string {
	if width <= 0 {
		return ""
	}

	m.outputViewport.Width = width - 2
	m.outputViewport.Height = height - 4

	var content string
	var title string

	switch m.activeView {
	case ViewOutput:
		title = "CLAUDE OUTPUT"
		if len(m.claudeOutput) == 0 {
			if m.monitorOnly {
				content = m.theme.Muted.Render("Monitor mode - Claude output from other terminal will not appear here")
			} else if !m.claudeRunning {
				content = m.theme.Muted.Render("Waiting for Claude to start...")
			} else {
				content = m.theme.Muted.Render("Claude is running, waiting for output...")
			}
		} else {
			content = m.outputViewport.View()
		}

	case ViewTasks:
		title = "TASKS"
		content = m.renderTaskList(width - 4)

	case ViewSubagents:
		title = "SUBAGENT ACTIVITY"
		content = m.renderSubagentsView(width - 4)

	case ViewProgress:
		title = "PROGRESS LOG"
		content = m.renderProgressLog(width - 4)

	case ViewGit:
		title = "GIT COMMITS"
		content = m.renderGitView(width - 4)
	}

	titleBar := m.theme.MainTitle.Render(" " + title + " ")
	titleBar = lipgloss.PlaceHorizontal(width, lipgloss.Center, titleBar)

	footer := ""
	if m.activeView == ViewOutput {
		if len(m.claudeOutput) > 0 {
			footer = m.theme.Muted.Render("(auto-scrolling) [Esc] pause  [j/k] scroll")
		}
	}

	mainStyle := m.theme.Main.Width(width).Height(height)

	return mainStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			titleBar,
			"",
			content,
			"",
			footer,
		),
	)
}

// renderTaskList renders the task list from PRD.md
func (m Model) renderTaskList(width int) string {
	if len(m.tasks) == 0 {
		return m.theme.Muted.Render("No PRD.md found or no tasks defined.\nRun 'setup-ralph' and let Claude create a PRD.")
	}

	var sb strings.Builder
	completed := 0
	for _, t := range m.tasks {
		if t.Complete {
			completed++
		}
	}

	sb.WriteString(m.theme.Muted.Render("(" + itoa(completed) + "/" + itoa(len(m.tasks)) + " complete)") + "\n\n")

	current := m.getCurrentTask()
	for _, task := range m.tasks {
		icon := "◌"
		style := m.theme.TaskPending
		suffix := ""

		if task.Complete {
			icon = "✓"
			style = m.theme.TaskComplete
		} else if current != nil && task.Title == current.Title {
			icon = "⟳"
			style = m.theme.TaskCurrent
			suffix = "  ← CURRENT"
		}

		line := icon + " " + wrapText(task.Title, width-4) + suffix
		sb.WriteString(style.Render(line) + "\n")
	}

	return sb.String()
}

// renderProgressLog renders the progress.txt entries
func (m Model) renderProgressLog(width int) string {
	if len(m.progressLog) == 0 {
		return m.theme.Muted.Render("No progress entries yet.\nCompleted tasks will appear here.")
	}

	var sb strings.Builder
	for i := len(m.progressLog) - 1; i >= 0; i-- {
		entry := m.progressLog[i]
		sb.WriteString(m.theme.ProgressTime.Render("["+entry.Timestamp+"]") + " ")
		sb.WriteString(m.theme.ProgressTitle.Render(entry.Title) + "\n")
		for _, detail := range entry.Details {
			sb.WriteString(m.theme.ProgressDetail.Render("  - "+detail) + "\n")
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// renderSubagentsView renders the subagent activity trace
func (m Model) renderSubagentsView(width int) string {
	if len(m.subagents) == 0 {
		return m.theme.Muted.Render("No subagent activity yet.\nTool calls will appear here when Claude runs.")
	}

	var sb strings.Builder

	// Show iteration info
	if m.iteration > 0 {
		sb.WriteString(m.theme.Muted.Render(fmt.Sprintf("Iteration %d | %d completed this session", m.iteration, m.tasksCompleted)) + "\n\n")
	}

	// Show recent subagent traces (most recent first)
	for i := len(m.subagents) - 1; i >= 0 && i >= len(m.subagents)-20; i-- {
		trace := m.subagents[i]

		// Status icon
		icon := "○"
		style := m.theme.TaskPending
		switch trace.Status {
		case orchestrator.SubagentStatusRunning:
			icon = "◐"
			style = m.theme.TaskCurrent
		case orchestrator.SubagentStatusComplete:
			icon = "✓"
			style = m.theme.TaskComplete
		case orchestrator.SubagentStatusError:
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
		}

		// Tool type badge
		typeBadge := trace.Type
		switch trace.Type {
		case "Task":
			typeBadge = "🤖 Task"
		case "Bash":
			typeBadge = "⌨️  Bash"
		case "Read":
			typeBadge = "📄 Read"
		case "Write":
			typeBadge = "✏️  Write"
		case "Edit":
			typeBadge = "🔧 Edit"
		case "Glob":
			typeBadge = "🔍 Glob"
		case "Grep":
			typeBadge = "🔎 Grep"
		}

		// Duration
		duration := ""
		if trace.Duration > 0 {
			duration = fmt.Sprintf(" (%s)", trace.Duration.Round(time.Millisecond))
		}

		line := fmt.Sprintf("%s %s: %s%s", icon, typeBadge, wrapText(trace.Input, width-30), duration)
		sb.WriteString(style.Render(line) + "\n")

		// Show output for completed traces (truncated)
		if trace.Output != "" && trace.Status == orchestrator.SubagentStatusComplete {
			output := wrapText(trace.Output, width-6)
			sb.WriteString(m.theme.Muted.Render("    → "+output) + "\n")
		}
	}

	return sb.String()
}

// renderGitView renders git commit history
func (m Model) renderGitView(width int) string {
	if len(m.gitCommits) == 0 {
		return m.theme.Muted.Render("No git commits found.\nCommits from ralph will appear here.")
	}

	var sb strings.Builder
	for _, commit := range m.gitCommits {
		sb.WriteString(commit + "\n")
	}
	return sb.String()
}

// renderHelp renders the help overlay
func (m Model) renderHelp() string {
	help := `
 ╭─────────────────────────────────────────╮
 │           RWATCH v2.0 HELP              │
 ├─────────────────────────────────────────┤
 │  Navigation                             │
 │  ─────────────────────────────────────  │
 │  Tab        Toggle sidebar focus        │
 │  1-5        Switch views                │
 │  j/↓        Scroll down                 │
 │  k/↑        Scroll up                   │
 │  g          Go to top                   │
 │  G          Go to bottom                │
 │  Ctrl+d/u   Half page down/up           │
 │                                         │
 │  General                                │
 │  ─────────────────────────────────────  │
 │  ?          Toggle this help            │
 │  Esc        Close help/unfocus sidebar  │
 │  q          Quit                        │
 │                                         │
 │  Views                                  │
 │  ─────────────────────────────────────  │
 │  1  Output    - Live Claude output      │
 │  2  Tasks     - PRD.md task list        │
 │  3  Subagents - Tool call activity      │
 │  4  Progress  - progress.txt log        │
 │  5  Git       - Recent commits          │
 │                                         │
 │  Orchestrator Mode                      │
 │  ─────────────────────────────────────  │
 │  Automatically runs Claude loop with    │
 │  completion token detection             │
 ╰─────────────────────────────────────────╯

           Press ? or Esc to close
`
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		m.theme.Help.Render(help),
	)
}

// Helper functions

func (m *Model) updateViewportSize() {
	sidebarWidth := m.calculateSidebarWidth()
	mainWidth := m.width - sidebarWidth - 4
	m.outputViewport.Width = mainWidth
	m.outputViewport.Height = m.height - 6
}

func (m Model) calculateSidebarWidth() int {
	if m.width < 80 {
		return 0 // Hide sidebar on small terminals
	}
	if m.width < 120 {
		return 20 // Narrow sidebar
	}
	return 25 // Full sidebar
}

func (m Model) getCurrentTask() *parser.Task {
	for i := range m.tasks {
		if !m.tasks[i].Complete {
			return &m.tasks[i]
		}
	}
	return nil
}

func (m Model) allTasksComplete() bool {
	for _, t := range m.tasks {
		if !t.Complete {
			return false
		}
	}
	return true
}

func getProjectName() string {
	// Get current directory name
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	parts := strings.Split(dir, string(os.PathSeparator))
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}

func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}
	return text[:width-3] + "..."
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
