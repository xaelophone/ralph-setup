package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/xaelophone/ralph-setup/internal/model"
	"github.com/xaelophone/ralph-setup/internal/orchestrator"
	"github.com/xaelophone/ralph-setup/internal/runner"
)

var (
	version          = "2.0.0"
	monitorOnly      bool
	orchestratorMode bool
	maxIterations    int
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "rwatch [-- claude-args...]",
		Short: "TUI orchestrator for Claude with ralph workflow",
		Long: `rwatch v2.0 - Advanced Claude Loop Orchestrator

Three modes of operation:

1. ORCHESTRATOR MODE (default):
   Runs Claude autonomously with real-time completion detection.
   Parses <promise>COMPLETE</promise> tokens and continues to next task.

2. LEGACY MODE (--legacy):
   Spawns Claude with PTY, displays output in TUI.
   Manual workflow - Claude stops when context fills.

3. MONITOR MODE (--monitor-only):
   Just watches files, no Claude process.
   Use alongside 'ralph-loop' in another terminal.

Features:
  • Real-time completion token detection
  • Subagent tracing (see Claude's tool calls)
  • Session persistence with crash recovery
  • Cross-iteration context injection

Usage:
  rwatch                    # Orchestrator mode (recommended)
  rwatch --legacy           # Legacy PTY mode
  rwatch --monitor-only     # Just watch files`,
		Version: version,
		RunE:    runRwatch,
	}

	rootCmd.Flags().BoolVar(&monitorOnly, "monitor-only", false, "Only monitor files, don't run Claude")
	rootCmd.Flags().BoolVar(&orchestratorMode, "legacy", false, "Use legacy PTY mode instead of orchestrator")
	rootCmd.Flags().IntVar(&maxIterations, "max-iterations", 100, "Maximum iterations in orchestrator mode")

	// Invert the flag meaning (--legacy means NOT orchestrator)
	orchestratorMode = true

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runRwatch(cmd *cobra.Command, args []string) error {
	// Get claude args (everything after --)
	claudeArgs := args

	// Check if we're in a ralph-initialized directory
	if _, err := os.Stat("CLAUDE.md"); os.IsNotExist(err) {
		fmt.Println("Warning: CLAUDE.md not found. Run 'setup-ralph' first to initialize ralph workflow.")
		fmt.Println("Continuing anyway...")
		fmt.Println()
	}

	// Check for legacy flag
	legacyMode, _ := cmd.Flags().GetBool("legacy")
	useOrchestrator := !legacyMode && !monitorOnly

	// Create the model with all options
	m := model.New(model.Options{
		MonitorOnly:      monitorOnly,
		OrchestratorMode: useOrchestrator,
		ClaudeArgs:       claudeArgs,
	})

	// Create the Bubbletea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
	)

	if useOrchestrator {
		// ORCHESTRATOR MODE - New advanced mode
		fmt.Println("🚀 Starting rwatch in orchestrator mode...")
		fmt.Println("   Real-time completion detection enabled")
		fmt.Println()

		go func() {
			config := orchestrator.DefaultConfig()
			config.MaxIterations = maxIterations

			orch := orchestrator.New(config, p)
			m.SetOrchestrator(orch)

			if err := orch.Start(); err != nil {
				p.Send(orchestrator.ErrorMsg{Error: err})
			}
		}()
	} else if !monitorOnly {
		// LEGACY MODE - PTY wrapper
		fmt.Println("🔧 Starting rwatch in legacy PTY mode...")
		fmt.Println()

		go func() {
			claudeRunner := runner.New(claudeArgs, p)
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
