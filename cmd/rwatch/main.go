package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xaelophone/ralph-setup/internal/config"
	"github.com/xaelophone/ralph-setup/internal/model"
	"github.com/xaelophone/ralph-setup/internal/orchestrator"
	"github.com/xaelophone/ralph-setup/internal/runner"
)

var (
	version       = "2.0.0"
	monitorOnly   bool
	legacyMode    bool
	maxIterations int
	cliBackend    string
	cliModel      string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rwatch [-- cli-args...]",
		Short: "TUI orchestrator for Claude/Codex with ralph workflow",
		Long: `rwatch v2.0 - Advanced AI Loop Orchestrator

Supports multiple CLI backends:
  • Claude Code (default): claude --dangerously-skip-permissions
  • OpenAI Codex:          codex exec --dangerously-bypass-approvals-and-sandbox

Three modes of operation:

1. ORCHESTRATOR MODE (default):
   Runs AI autonomously with real-time completion detection.
   Parses <promise>COMPLETE</promise> tokens and continues to next task.

2. LEGACY MODE (--legacy):
   Spawns CLI with PTY, displays output in TUI.
   Manual workflow - CLI stops when context fills.

3. MONITOR MODE (--monitor-only):
   Just watches files, no CLI process.
   Use alongside 'ralph-loop' in another terminal.

Features:
  • Multi-CLI support (Claude, Codex)
  • Real-time completion token detection
  • Subagent tracing (see AI tool calls)
  • Session persistence with crash recovery
  • Cross-iteration context injection

Usage:
  rwatch                        # Claude (default)
  rwatch --cli codex            # OpenAI Codex
  rwatch --cli claude --model claude-sonnet-4-20250514
  rwatch --legacy               # Legacy PTY mode
  rwatch --monitor-only         # Just watch files

Configuration (precedence: flags > env > .ralph-config.json > defaults):
  Flags:       --cli, --model
  Environment: RALPH_CLI, RALPH_MODEL
  File:        .ralph-config.json {"cli": "codex", "model": "gpt-4o"}`,
		Version: version,
		RunE:    runRwatch,
	}

	rootCmd.Flags().BoolVar(&monitorOnly, "monitor-only", false, "Only monitor files, don't run AI")
	rootCmd.Flags().BoolVar(&legacyMode, "legacy", false, "Use legacy PTY mode instead of orchestrator")
	rootCmd.Flags().IntVar(&maxIterations, "max-iterations", 100, "Maximum iterations in orchestrator mode")
	rootCmd.Flags().StringVarP(&cliBackend, "cli", "c", "", "CLI backend: claude (default) or codex")
	rootCmd.Flags().StringVarP(&cliModel, "model", "m", "", "Model to use (e.g., claude-sonnet-4-20250514, gpt-4o)")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRwatch(cmd *cobra.Command, args []string) error {
	// Get extra CLI args (everything after --)
	extraArgs := args

	// Load CLI configuration (flags > env > file > defaults)
	cliConfig := config.LoadCLIConfig(cliBackend, cliModel)

	// Validate CLI backend
	if !cliConfig.Backend.IsValid() {
		return fmt.Errorf("invalid CLI backend: %s (use 'claude' or 'codex')", cliConfig.Backend)
	}

	// Check if we're in a ralph-initialized directory
	if _, err := os.Stat("CLAUDE.md"); os.IsNotExist(err) {
		fmt.Println("Warning: CLAUDE.md not found. Run 'setup-ralph' first to initialize ralph workflow.")
		fmt.Println("Continuing anyway...")
		fmt.Println()
	}

	// Determine mode: orchestrator (default), legacy PTY, or monitor-only
	useOrchestrator := !legacyMode && !monitorOnly

	// Create the model with all options
	m := model.New(model.Options{
		MonitorOnly:      monitorOnly,
		OrchestratorMode: useOrchestrator,
		ClaudeArgs:       extraArgs,
	})

	// Create the Bubbletea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if useOrchestrator {
		// ORCHESTRATOR MODE - New advanced mode
		cliName := cliConfig.Backend.String()
		fmt.Printf("🚀 Starting rwatch in orchestrator mode (CLI: %s)...\n", cliName)
		if cliConfig.Model != "" {
			fmt.Printf("   Model: %s\n", cliConfig.Model)
		}
		fmt.Println("   Real-time completion detection enabled")
		fmt.Println()

		go func() {
			orchConfig := orchestrator.DefaultConfig()
			orchConfig.MaxIterations = maxIterations
			orchConfig.CLIConfig = cliConfig
			orchConfig.CLIConfig.ExtraArgs = extraArgs

			orch := orchestrator.New(orchConfig, p)
			m.SetOrchestrator(orch)

			if err := orch.Start(); err != nil {
				p.Send(orchestrator.ErrorMsg{Error: err})
			}
		}()
	} else if !monitorOnly {
		// LEGACY MODE - PTY wrapper (only supports Claude for now)
		fmt.Println("🔧 Starting rwatch in legacy PTY mode...")
		fmt.Println("   Note: Legacy mode only supports Claude CLI")
		fmt.Println()

		go func() {
			claudeRunner := runner.New(extraArgs, p)
			if err := claudeRunner.Start(); err != nil {
				p.Send(model.ErrorMsg{Error: err})
			}
		}()
	} else {
		// MONITOR MODE - Just watch files
		fmt.Println("👀 Starting rwatch in monitor mode...")
		fmt.Println("   Watching PRD.md and progress.txt for changes")
		fmt.Println()
	}

	// Run the TUI (blocks until quit)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}
