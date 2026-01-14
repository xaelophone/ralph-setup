package parser

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Task represents a task from PRD.md
type Task struct {
	Title    string
	Complete bool
	Line     int
}

var (
	// Match incomplete tasks: - [ ] Task description
	incompletePattern = regexp.MustCompile(`^[\s]*-\s*\[\s*\]\s*(.+)$`)
	// Match complete tasks: - [x] Task description or - [X] Task description
	completePattern = regexp.MustCompile(`^[\s]*-\s*\[[xX]\]\s*(.+)$`)
)

// ParsePRD parses a PRD.md file and extracts tasks
func ParsePRD(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tasks []Task
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for incomplete task
		if matches := incompletePattern.FindStringSubmatch(line); matches != nil {
			tasks = append(tasks, Task{
				Title:    strings.TrimSpace(matches[1]),
				Complete: false,
				Line:     lineNum,
			})
			continue
		}

		// Check for complete task
		if matches := completePattern.FindStringSubmatch(line); matches != nil {
			tasks = append(tasks, Task{
				Title:    strings.TrimSpace(matches[1]),
				Complete: true,
				Line:     lineNum,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tasks, nil
}

// CountCompleted returns the number of completed tasks
func CountCompleted(tasks []Task) int {
	count := 0
	for _, t := range tasks {
		if t.Complete {
			count++
		}
	}
	return count
}

// GetCurrentTask returns the first incomplete task
func GetCurrentTask(tasks []Task) *Task {
	for i := range tasks {
		if !tasks[i].Complete {
			return &tasks[i]
		}
	}
	return nil
}
