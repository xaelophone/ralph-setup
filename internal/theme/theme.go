package theme

import "github.com/charmbracelet/lipgloss"

// Theme defines the visual styling for rwatch
type Theme struct {
	// Status bar
	Title         lipgloss.Style
	StatusRunning lipgloss.Style
	StatusStopped lipgloss.Style
	StatusMonitor lipgloss.Style
	ProjectName   lipgloss.Style
	Help          lipgloss.Style

	// Sidebar
	Sidebar           lipgloss.Style
	SidebarHeader     lipgloss.Style
	SidebarItem       lipgloss.Style
	SidebarItemActive lipgloss.Style
	Divider           lipgloss.Style
	CurrentTask       lipgloss.Style

	// Main area
	Main      lipgloss.Style
	MainTitle lipgloss.Style

	// Tasks
	TaskComplete lipgloss.Style
	TaskCurrent  lipgloss.Style
	TaskPending  lipgloss.Style

	// Progress
	ProgressTime   lipgloss.Style
	ProgressTitle  lipgloss.Style
	ProgressDetail lipgloss.Style

	// General
	Muted   lipgloss.Style
	Success lipgloss.Style
	Error   lipgloss.Style
}

// Colors
var (
	colorPrimary   = lipgloss.Color("#7C3AED") // Purple
	colorSecondary = lipgloss.Color("#06B6D4") // Cyan
	colorSuccess   = lipgloss.Color("#22C55E") // Green
	colorWarning   = lipgloss.Color("#F59E0B") // Amber
	colorError     = lipgloss.Color("#EF4444") // Red
	colorMuted     = lipgloss.Color("#6B7280") // Gray
	colorText      = lipgloss.Color("#F9FAFB") // Light
	colorBorder    = lipgloss.Color("#374151") // Dark gray
	colorBg        = lipgloss.Color("#111827") // Very dark
	colorBgAlt     = lipgloss.Color("#1F2937") // Slightly lighter
)

// Default returns the default dark theme
func Default() Theme {
	return Theme{
		// Status bar
		Title: lipgloss.NewStyle().
			Background(colorPrimary).
			Foreground(colorText).
			Bold(true).
			Padding(0, 1),

		StatusRunning: lipgloss.NewStyle().
			Foreground(colorSuccess).
			Bold(true),

		StatusStopped: lipgloss.NewStyle().
			Foreground(colorError),

		StatusMonitor: lipgloss.NewStyle().
			Foreground(colorSecondary),

		ProjectName: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true),

		Help: lipgloss.NewStyle().
			Foreground(colorMuted),

		// Sidebar
		Sidebar: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1),

		SidebarHeader: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true),

		SidebarItem: lipgloss.NewStyle().
			Foreground(colorMuted),

		SidebarItemActive: lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true),

		Divider: lipgloss.NewStyle().
			Foreground(colorBorder),

		CurrentTask: lipgloss.NewStyle().
			Foreground(colorText),

		// Main area
		Main: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1),

		MainTitle: lipgloss.NewStyle().
			Foreground(colorSecondary).
			Bold(true),

		// Tasks
		TaskComplete: lipgloss.NewStyle().
			Foreground(colorSuccess),

		TaskCurrent: lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true),

		TaskPending: lipgloss.NewStyle().
			Foreground(colorMuted),

		// Progress
		ProgressTime: lipgloss.NewStyle().
			Foreground(colorMuted),

		ProgressTitle: lipgloss.NewStyle().
			Foreground(colorText).
			Bold(true),

		ProgressDetail: lipgloss.NewStyle().
			Foreground(colorMuted),

		// General
		Muted: lipgloss.NewStyle().
			Foreground(colorMuted),

		Success: lipgloss.NewStyle().
			Foreground(colorSuccess),

		Error: lipgloss.NewStyle().
			Foreground(colorError),
	}
}

// Light returns a light theme
func Light() Theme {
	// For light theme, we'd invert colors
	// For now, return default
	return Default()
}

// DetectTheme tries to detect if terminal is light or dark
func DetectTheme() Theme {
	// TODO: Implement terminal background detection
	// Using OSC 11 escape sequence
	return Default()
}
