package model

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/xaelophone/ralph-setup/internal/parser"
	"github.com/xaelophone/ralph-setup/internal/watcher"
)

// loadTasks loads and parses PRD.md
func (m Model) loadTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, err := parser.ParsePRD("PRD.md")
		if err != nil {
			// Not an error - PRD might not exist yet
			return TasksUpdatedMsg{Tasks: []parser.Task{}}
		}
		return TasksUpdatedMsg{Tasks: tasks}
	}
}

// loadProgress loads and parses progress.txt
func (m Model) loadProgress() tea.Cmd {
	return func() tea.Msg {
		entries, err := parser.ParseProgress("progress.txt")
		if err != nil {
			// Not an error - progress.txt might be empty
			return ProgressUpdatedMsg{Entries: []parser.ProgressEntry{}}
		}
		return ProgressUpdatedMsg{Entries: entries}
	}
}

// watchFiles starts watching PRD.md and progress.txt for changes
func (m Model) watchFiles() tea.Cmd {
	return func() tea.Msg {
		changes := make(chan string)
		go watcher.Watch([]string{"PRD.md", "progress.txt"}, changes)

		// This will block and send file change messages
		// In a real implementation, we'd use a subscription pattern
		go func() {
			for file := range changes {
				// Note: This is a simplified approach
				// In production, we'd send messages through a channel to the program
				_ = file
			}
		}()

		return nil
	}
}

// loadGitCommits loads recent git commits
func (m Model) loadGitCommits() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement git log parsing
		return GitUpdatedMsg{Commits: []string{}}
	}
}
